package jobs

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

var ErrNoJob = errors.New("no job available")

type Store struct {
	pool        *pgxpool.Pool
	defaultMax  int
	orphanAfter time.Duration
}

func NewStore(pool *pgxpool.Pool, defaultMaxAttempts int) *Store {
	if defaultMaxAttempts <= 0 {
		defaultMaxAttempts = 3
	}
	return &Store{
		pool:        pool,
		defaultMax:  defaultMaxAttempts,
		orphanAfter: 10 * time.Minute,
	}
}

type EnqueueParams struct {
	OutfitID    uuid.UUID
	Kind        string
	Payload     any
	MaxAttempts int
	RunAt       time.Time
}

func (s *Store) Enqueue(ctx context.Context, p EnqueueParams) (*Job, error) {
	if p.Kind != KindURL && p.Kind != KindUpload {
		return nil, fmt.Errorf("invalid job kind %q", p.Kind)
	}
	raw, err := EncodePayload(p.Payload)
	if err != nil {
		return nil, fmt.Errorf("encode payload: %w", err)
	}
	maxAttempts := p.MaxAttempts
	if maxAttempts <= 0 {
		maxAttempts = s.defaultMax
	}
	runAt := p.RunAt
	if runAt.IsZero() {
		runAt = time.Now()
	}

	var j Job
	err = s.pool.QueryRow(ctx, `
		INSERT INTO jobs (outfit_id, kind, payload, status, max_attempts, run_at)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id, outfit_id, kind, payload, status, attempts, max_attempts, run_at, locked_at, last_error, created_at
	`, p.OutfitID, p.Kind, raw, StatusQueued, maxAttempts, runAt).Scan(
		&j.ID, &j.OutfitID, &j.Kind, &j.Payload, &j.Status, &j.Attempts, &j.MaxAttempts,
		&j.RunAt, &j.LockedAt, &j.LastError, &j.CreatedAt,
	)
	if err != nil {
		return nil, err
	}
	return &j, nil
}

// Claim locks the next queued job ready to run (FOR UPDATE SKIP LOCKED).
func (s *Store) Claim(ctx context.Context) (*Job, error) {
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(ctx)

	var j Job
	err = tx.QueryRow(ctx, `
		UPDATE jobs SET status=$1, locked_at=NOW(), attempts=attempts+1
		WHERE id = (
			SELECT id FROM jobs
			WHERE status=$2 AND run_at <= NOW()
			ORDER BY run_at
			FOR UPDATE SKIP LOCKED
			LIMIT 1
		)
		RETURNING id, outfit_id, kind, payload, status, attempts, max_attempts, run_at, locked_at, last_error, created_at
	`, StatusRunning, StatusQueued).Scan(
		&j.ID, &j.OutfitID, &j.Kind, &j.Payload, &j.Status, &j.Attempts, &j.MaxAttempts,
		&j.RunAt, &j.LockedAt, &j.LastError, &j.CreatedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrNoJob
	}
	if err != nil {
		return nil, err
	}
	if err := tx.Commit(ctx); err != nil {
		return nil, err
	}
	return &j, nil
}

func (s *Store) Complete(ctx context.Context, jobID uuid.UUID) error {
	_, err := s.pool.Exec(ctx, `
		UPDATE jobs SET status=$2, locked_at=NULL, last_error=NULL WHERE id=$1
	`, jobID, StatusDone)
	return err
}

// Fail records an error. If attempts remain, requeues with exponential backoff;
// otherwise marks the job dead. Returns whether the job is permanently dead.
func (s *Store) Fail(ctx context.Context, jobID uuid.UUID, errMsg string) (dead bool, err error) {
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return false, err
	}
	defer tx.Rollback(ctx)

	var attempts, maxAttempts int
	var outfitID uuid.UUID
	err = tx.QueryRow(ctx, `
		SELECT attempts, max_attempts, outfit_id FROM jobs WHERE id=$1 FOR UPDATE
	`, jobID).Scan(&attempts, &maxAttempts, &outfitID)
	if err != nil {
		return false, err
	}

	if attempts >= maxAttempts {
		_, err = tx.Exec(ctx, `
			UPDATE jobs SET status=$2, locked_at=NULL, last_error=$3 WHERE id=$1
		`, jobID, StatusDead, errMsg)
		if err != nil {
			return false, err
		}
		if err := tx.Commit(ctx); err != nil {
			return false, err
		}
		return true, nil
	}

	runAt := time.Now().Add(Backoff(attempts))
	_, err = tx.Exec(ctx, `
		UPDATE jobs SET status=$2, locked_at=NULL, last_error=$3, run_at=$4 WHERE id=$1
	`, jobID, StatusQueued, errMsg, runAt)
	if err != nil {
		return false, err
	}
	return false, tx.Commit(ctx)
}

// ReapOrphans requeues jobs stuck in running past the orphan window.
func (s *Store) ReapOrphans(ctx context.Context) (int64, error) {
	ct, err := s.pool.Exec(ctx, `
		UPDATE jobs SET status=$1, locked_at=NULL
		WHERE status=$2
		  AND locked_at IS NOT NULL
		  AND locked_at < NOW() - ($3 * INTERVAL '1 second')
	`, StatusQueued, StatusRunning, s.orphanAfter.Seconds())
	if err != nil {
		return 0, err
	}
	return ct.RowsAffected(), nil
}

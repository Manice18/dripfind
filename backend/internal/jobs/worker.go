package jobs

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/manice18/dripfind/backend/internal/history"
)

// Runner executes a claimed analysis job.
type Runner interface {
	RunURL(ctx context.Context, outfitID uuid.UUID, sourceURL string) error
	RunUpload(ctx context.Context, outfitID uuid.UUID, imageKey, sourceLabel string) error
}

type Worker struct {
	Jobs         *Store
	Outfits      *history.Store
	Pipeline     Runner
	Log          *slog.Logger
	Concurrency  int
	PollInterval time.Duration
	ReapInterval time.Duration
}

func (w *Worker) Run(ctx context.Context) {
	if w.Concurrency <= 0 {
		w.Concurrency = 4
	}
	if w.PollInterval <= 0 {
		w.PollInterval = 500 * time.Millisecond
	}
	if w.ReapInterval <= 0 {
		w.ReapInterval = time.Minute
	}
	if w.Log == nil {
		w.Log = slog.New(slog.NewJSONHandler(os.Stdout, nil))
	}

	// Startup sweep for orphans left by a previous crash.
	if n, err := w.Jobs.ReapOrphans(ctx); err != nil {
		w.Log.Error("reap orphans on start", "error", err)
	} else if n > 0 {
		w.Log.Info("requeued orphaned jobs", "count", n)
	}

	var wg sync.WaitGroup
	sem := make(chan struct{}, w.Concurrency)

	reapTicker := time.NewTicker(w.ReapInterval)
	defer reapTicker.Stop()

	pollTicker := time.NewTicker(w.PollInterval)
	defer pollTicker.Stop()

	w.Log.Info("worker started", "concurrency", w.Concurrency)

	for {
		select {
		case <-ctx.Done():
			w.Log.Info("worker shutting down, waiting for in-flight jobs")
			wg.Wait()
			w.Log.Info("worker stopped")
			return

		case <-reapTicker.C:
			if n, err := w.Jobs.ReapOrphans(ctx); err != nil {
				w.Log.Error("reap orphans", "error", err)
			} else if n > 0 {
				w.Log.Info("requeued orphaned jobs", "count", n)
			}

		case <-pollTicker.C:
			// Try to fill free slots without blocking the select loop.
			for {
				select {
				case <-ctx.Done():
					wg.Wait()
					return
				case sem <- struct{}{}:
				default:
					goto waitNext
				}

				job, err := w.Jobs.Claim(ctx)
				if err != nil {
					<-sem
					if !errors.Is(err, ErrNoJob) {
						w.Log.Error("claim job", "error", err)
					}
					goto waitNext
				}

				wg.Add(1)
				go func(j *Job) {
					defer wg.Done()
					defer func() { <-sem }()
					w.process(ctx, j)
				}(job)
			}
		waitNext:
		}
	}
}

func (w *Worker) process(_ context.Context, job *Job) {
	// Detach from parent cancel so in-flight work can finish on shutdown signal,
	// but still bound total runtime.
	runCtx, cancel := context.WithTimeout(context.Background(), 3*time.Minute)
	defer cancel()

	log := w.Log.With("job_id", job.ID, "outfit_id", job.OutfitID, "kind", job.Kind, "attempt", job.Attempts)
	log.Info("job started")

	err := w.execute(runCtx, job)
	if err == nil {
		if cerr := w.Jobs.Complete(context.Background(), job.ID); cerr != nil {
			log.Error("complete job", "error", cerr)
		} else {
			log.Info("job completed")
		}
		return
	}

	log.Error("job failed", "error", err)
	dead, ferr := w.Jobs.Fail(context.Background(), job.ID, err.Error())
	if ferr != nil {
		log.Error("record job failure", "error", ferr)
		return
	}
	if dead {
		msg := fmt.Sprintf("analysis failed after %d attempts: %s", job.Attempts, err.Error())
		if merr := w.Outfits.MarkFailed(context.Background(), job.OutfitID, msg); merr != nil {
			log.Error("mark outfit failed", "error", merr)
		}
		log.Info("job marked dead")
	}
}

func (w *Worker) execute(ctx context.Context, job *Job) error {
	switch job.Kind {
	case KindURL:
		p, err := DecodeURLPayload(job.Payload)
		if err != nil {
			return fmt.Errorf("decode url payload: %w", err)
		}
		if p.URL == "" {
			return fmt.Errorf("url payload missing url")
		}
		return w.Pipeline.RunURL(ctx, job.OutfitID, p.URL)

	case KindUpload:
		p, err := DecodeUploadPayload(job.Payload)
		if err != nil {
			return fmt.Errorf("decode upload payload: %w", err)
		}
		if p.ImageKey == "" {
			return fmt.Errorf("upload payload missing image_key")
		}
		return w.Pipeline.RunUpload(ctx, job.OutfitID, p.ImageKey, p.SourceLabel)

	default:
		return fmt.Errorf("unknown job kind %q", job.Kind)
	}
}

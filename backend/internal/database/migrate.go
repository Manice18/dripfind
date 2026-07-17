package database

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/jackc/pgx/v5/pgxpool"
)

// RunMigrations applies *.up.sql files from migrationsDir in lexical order.
// Tracks applied migrations in schema_migrations.
func RunMigrations(ctx context.Context, pool *pgxpool.Pool, migrationsDir string) error {
	_, err := pool.Exec(ctx, `
		CREATE TABLE IF NOT EXISTS schema_migrations (
			version TEXT PRIMARY KEY,
			applied_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
		)
	`)
	if err != nil {
		return fmt.Errorf("create schema_migrations: %w", err)
	}

	entries, err := os.ReadDir(migrationsDir)
	if err != nil {
		return fmt.Errorf("read migrations: %w", err)
	}

	var ups []string
	for _, e := range entries {
		name := e.Name()
		if strings.HasSuffix(name, ".up.sql") {
			ups = append(ups, name)
		}
	}
	sort.Strings(ups)

	for _, name := range ups {
		version := strings.TrimSuffix(name, ".up.sql")
		var exists bool
		err := pool.QueryRow(ctx, `SELECT EXISTS(SELECT 1 FROM schema_migrations WHERE version=$1)`, version).Scan(&exists)
		if err != nil {
			return err
		}
		if exists {
			continue
		}

		body, err := os.ReadFile(filepath.Join(migrationsDir, name))
		if err != nil {
			return err
		}

		tx, err := pool.Begin(ctx)
		if err != nil {
			return err
		}

		// Strip migrate comments for raw execution
		sql := string(body)
		sql = strings.ReplaceAll(sql, "-- +migrate Up", "")
		sql = strings.ReplaceAll(sql, "-- +migrate Down", "")

		if _, err := tx.Exec(ctx, sql); err != nil {
			_ = tx.Rollback(ctx)
			return fmt.Errorf("apply %s: %w", name, err)
		}
		if _, err := tx.Exec(ctx, `INSERT INTO schema_migrations(version) VALUES($1)`, version); err != nil {
			_ = tx.Rollback(ctx)
			return err
		}
		if err := tx.Commit(ctx); err != nil {
			return err
		}
	}
	return nil
}

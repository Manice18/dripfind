CREATE TABLE IF NOT EXISTS jobs (
    id           UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    outfit_id    UUID NOT NULL REFERENCES outfits(id) ON DELETE CASCADE,
    kind         TEXT NOT NULL,
    payload      JSONB NOT NULL,
    status       TEXT NOT NULL DEFAULT 'queued',
    attempts     INT  NOT NULL DEFAULT 0,
    max_attempts INT  NOT NULL DEFAULT 3,
    run_at       TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    locked_at    TIMESTAMPTZ,
    last_error   TEXT,
    created_at   TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_jobs_claim ON jobs(status, run_at) WHERE status = 'queued';
CREATE INDEX IF NOT EXISTS idx_jobs_outfit_id ON jobs(outfit_id);
CREATE INDEX IF NOT EXISTS idx_jobs_running ON jobs(status, locked_at) WHERE status = 'running';

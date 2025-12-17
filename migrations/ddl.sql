CREATE TABLE jobs (
    id UUID PRIMARY KEY,
    type TEXT NOT NULL,                 -- send_email, generate_invoice, etc
    callback_url TEXT NOT NULL,         -- endpoint real que ejecuta la lógica
    payload JSONB NOT NULL,             -- datos del job
    status TEXT NOT NULL,
    max_retries INT NOT NULL,
    scheduled_at TIMESTAMPTZ NOT NULL, -- para jobs futuros o cron
    locked_at TIMESTAMPTZ,              -- cuando un worker lo tomó
    locked_by TEXT,                     -- cuando un worker lo tomó
    completed_at TIMESTAMPTZ,
    priority INT NOT NULL,    -- si luego usás prioridades en Rabbit
    created_at TIMESTAMPTZ NOT NULL,
    updated_at TIMESTAMPTZ NOT NULL
);

CREATE INDEX idx_jobs_status ON jobs(status);
CREATE INDEX idx_jobs_scheduled_at ON jobs(scheduled_at);

CREATE TABLE job_attempts (
    id UUID PRIMARY KEY,
    job_id UUID NOT NULL REFERENCES jobs(id),
    attempt_number INT NOT NULL,
    started_at TIMESTAMPTZ NOT NULL,
    finished_at TIMESTAMPTZ,
    status TEXT NOT NULL, 
    error_message TEXT,
    http_status INT,
    created_at TIMESTAMPTZ NOT NULL
);

CREATE INDEX idx_job_attempts_job_id ON job_attempts(job_id);


CREATE TABLE job_events (
    id UUID PRIMARY KEY,
    job_id UUID NOT NULL REFERENCES jobs(id),
    event_type TEXT NOT NULL,   -- created, queued, started, retried, failed, completed, dead
    message TEXT,
    metadata JSONB,
    created_at TIMESTAMPTZ NOT NULL
);

CREATE INDEX idx_job_events_job_id ON job_events(job_id);


CREATE TABLE job_dead_letters (
    id UUID PRIMARY KEY,
    job_id UUID NOT NULL REFERENCES jobs(id),
    reason TEXT NOT NULL,
    last_error TEXT,
    failed_at TIMESTAMPTZ NOT NULL
);


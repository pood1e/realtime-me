ALTER TABLE processing_jobs
    ADD COLUMN lease_token TEXT;

UPDATE processing_jobs
SET status = 'pending', lease_until = NULL, available_time = now(), update_time = now()
WHERE status = 'running';

DELETE FROM processing_jobs WHERE status = 'completed';

ALTER TABLE processing_jobs
    ADD CONSTRAINT processing_jobs_lease_check CHECK (
        (status = 'running' AND lease_token IS NOT NULL AND lease_until IS NOT NULL)
        OR (status <> 'running' AND lease_token IS NULL AND lease_until IS NULL)
    );

CREATE UNIQUE INDEX processing_jobs_lease_token_idx
    ON processing_jobs (lease_token)
    WHERE lease_token IS NOT NULL;

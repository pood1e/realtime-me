UPDATE books
SET processing_status = 'pending', update_time = now()
WHERE processing_status = 'failed'
  AND EXISTS (
      SELECT 1
      FROM processing_jobs
      WHERE processing_jobs.kind = 'book'
        AND processing_jobs.resource_uid = books.uid
        AND processing_jobs.status = 'failed'
  );

UPDATE processing_jobs
SET status = 'pending',
    attempts = 0,
    available_time = now(),
    lease_until = NULL,
    error_code = '',
    update_time = now()
WHERE kind = 'book' AND status = 'failed';

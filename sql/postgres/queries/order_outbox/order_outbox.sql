что -- name: Enqueue :exec
-- Queues an order event for the ERP. Must run in the transaction that makes
-- the change the event describes.
INSERT INTO order_outbox (event_id, event_type, schema_version, order_id, payload)
VALUES ($1, $2, $3, $4, $5);

-- name: LockPending :many
-- Next batch of unpublished events in publication order, row-locked until
-- commit. A concurrent relay blocks on the lock instead of skipping ahead, so
-- events are never published out of order. Transaction only.
SELECT * FROM order_outbox
WHERE published_at IS NULL
ORDER BY id
LIMIT $1
FOR UPDATE;

-- name: MarkPublished :exec
UPDATE order_outbox
SET published_at = now(),
    attempts     = attempts + 1,
    last_error   = NULL
WHERE id = $1;

-- name: MarkFailed :exec
-- Records a failed publish attempt; the event stays pending.
UPDATE order_outbox
SET attempts   = attempts + 1,
    last_error = $2
WHERE id = $1;

-- name: DeletePublishedBefore :execrows
-- Retention sweep: drops events published before the cutoff.
DELETE FROM order_outbox
WHERE published_at IS NOT NULL
  AND published_at < $1;

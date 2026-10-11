-- name: Register :execrows
-- Records an inbound ERP event. Affects 0 rows when event_id was already
-- registered, i.e. the event is a redelivery and must be skipped.
INSERT INTO order_inbound_events (event_id, event_type, order_id)
VALUES ($1, $2, $3)
ON CONFLICT (event_id) DO NOTHING;

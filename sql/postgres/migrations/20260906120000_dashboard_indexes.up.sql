-- Supports the admin dashboard aggregation (internal/service/dashboard): full
-- scans over orders filtered/bucketed by created_at, status and payment_status.
-- idx_orders_open already exists but only covers status in ('pending','paid').
create index if not exists idx_orders_created_at on orders (created_at);
create index if not exists idx_orders_status on orders (status);
create index if not exists idx_orders_payment_status on orders (payment_status);

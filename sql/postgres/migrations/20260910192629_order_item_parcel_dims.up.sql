-- Parcel snapshot per line, captured at checkout from products.weight / length /
-- width / height (kg / cm). Kept on the line itself — like every other order_item
-- column — so an admin can re-quote shipping when editing the order even after
-- the catalogue product is gone.
ALTER TABLE order_items
    ADD COLUMN weight_kg  decimal(15, 8) NOT NULL DEFAULT 0,
    ADD COLUMN length_cm  decimal(15, 8) NOT NULL DEFAULT 0,
    ADD COLUMN width_cm   decimal(15, 8) NOT NULL DEFAULT 0,
    ADD COLUMN height_cm  decimal(15, 8) NOT NULL DEFAULT 0;

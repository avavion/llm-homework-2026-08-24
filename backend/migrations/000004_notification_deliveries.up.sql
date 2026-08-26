CREATE TABLE notification_deliveries (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    product_id uuid NOT NULL REFERENCES products(id) ON DELETE CASCADE,
    channel text NOT NULL,
    scheduled_for timestamptz NOT NULL,
    delivered_at timestamptz NOT NULL DEFAULT now(),
    CONSTRAINT notification_deliveries_channel_not_empty CHECK (channel <> ''),
    CONSTRAINT notification_deliveries_unique_slot UNIQUE (product_id, scheduled_for, channel)
);

CREATE INDEX notification_deliveries_product_id_idx ON notification_deliveries(product_id);

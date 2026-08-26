CREATE TABLE products (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    account_id uuid NOT NULL REFERENCES accounts(id) ON DELETE CASCADE,
    name text NOT NULL,
    date_type text NOT NULL,
    expiry_date timestamptz NOT NULL,
    quantity double precision,
    unit text,
    product_group text,
    storage_location text,
    country_code text,
    alert_threshold_minutes integer,
    lifecycle_status text NOT NULL DEFAULT 'active',
    completed_at timestamptz,
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now(),
    CONSTRAINT products_name_not_empty CHECK (name <> ''),
    CONSTRAINT products_date_type_valid CHECK (date_type IN ('use_by', 'best_before')),
    CONSTRAINT products_lifecycle_status_valid CHECK (lifecycle_status IN ('active', 'used', 'discarded')),
    CONSTRAINT products_alert_threshold_minimum CHECK (alert_threshold_minutes IS NULL OR alert_threshold_minutes >= 60),
    CONSTRAINT products_completed_at_matches_status CHECK (
        (lifecycle_status = 'active' AND completed_at IS NULL) OR
        (lifecycle_status <> 'active' AND completed_at IS NOT NULL)
    )
);

CREATE INDEX products_account_id_idx ON products(account_id);
CREATE INDEX products_account_id_lifecycle_status_idx ON products(account_id, lifecycle_status);

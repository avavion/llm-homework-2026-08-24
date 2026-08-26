CREATE TABLE product_drafts (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    account_id uuid NOT NULL REFERENCES accounts(id) ON DELETE CASCADE,
    source_reference text NOT NULL DEFAULT '',
    raw_text text NOT NULL DEFAULT '',
    name text,
    date_type text,
    expiry_date timestamptz,
    quantity double precision,
    unit text,
    product_group text,
    storage_location text,
    country_code text,
    status text NOT NULL DEFAULT 'pending',
    approved_product_id uuid REFERENCES products(id) ON DELETE SET NULL,
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now(),
    CONSTRAINT product_drafts_status_valid CHECK (status IN ('pending', 'approved', 'rejected')),
    CONSTRAINT product_drafts_date_type_valid CHECK (date_type IS NULL OR date_type IN ('use_by', 'best_before')),
    CONSTRAINT product_drafts_approved_has_product CHECK (
        (status = 'approved' AND approved_product_id IS NOT NULL) OR
        (status <> 'approved' AND approved_product_id IS NULL)
    )
);

CREATE INDEX product_drafts_account_id_idx ON product_drafts(account_id);

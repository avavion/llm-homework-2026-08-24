CREATE TABLE accounts (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    email_normalized text NOT NULL UNIQUE,
    password_hash text NOT NULL,
    created_at timestamptz NOT NULL DEFAULT now(),
    CONSTRAINT accounts_email_normalized_lowercase
        CHECK (email_normalized = lower(btrim(email_normalized))),
    CONSTRAINT accounts_email_normalized_not_empty
        CHECK (email_normalized <> ''),
    CONSTRAINT accounts_password_hash_not_empty
        CHECK (password_hash <> '')
);

CREATE TABLE auth_sessions (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    account_id uuid NOT NULL REFERENCES accounts(id) ON DELETE CASCADE,
    token_hash bytea NOT NULL UNIQUE,
    expires_at timestamptz NOT NULL,
    created_at timestamptz NOT NULL DEFAULT now(),
    CONSTRAINT auth_sessions_token_hash_sha256_length
        CHECK (octet_length(token_hash) = 32)
);

CREATE INDEX auth_sessions_account_id_idx ON auth_sessions(account_id);
CREATE INDEX auth_sessions_expires_at_idx ON auth_sessions(expires_at);

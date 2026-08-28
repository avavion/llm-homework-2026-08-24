CREATE TABLE account_profiles (
    account_id uuid PRIMARY KEY REFERENCES accounts(id) ON DELETE CASCADE,
    country_code char(2) NOT NULL,
    language text NOT NULL,
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now(),
    CONSTRAINT account_profiles_country_code_uppercase CHECK (country_code = upper(country_code)),
    CONSTRAINT account_profiles_language_supported CHECK (language IN ('ru', 'en'))
);

CREATE TABLE account_notification_settings (
    account_id uuid NOT NULL REFERENCES accounts(id) ON DELETE CASCADE,
    product_group text NOT NULL,
    alert_threshold_minutes integer NOT NULL,
    updated_at timestamptz NOT NULL DEFAULT now(),
    PRIMARY KEY (account_id, product_group),
    CONSTRAINT account_notification_settings_group_supported CHECK (
        product_group IN ('refrigerated_perishable', 'fresh_produce', 'frozen', 'shelf_stable', 'other')
    ),
    CONSTRAINT account_notification_settings_threshold_minimum CHECK (alert_threshold_minutes >= 60)
);

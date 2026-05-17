CREATE EXTENSION IF NOT EXISTS pgcrypto;

CREATE TABLE IF NOT EXISTS users (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    external_id TEXT,
    email TEXT UNIQUE NOT NULL,
    full_name TEXT NOT NULL,
    status TEXT NOT NULL,
    user_type TEXT NOT NULL DEFAULT 'PF',
    password_hash TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1
          FROM information_schema.columns
         WHERE table_name = 'users'
           AND column_name = 'user_type'
    ) THEN
        ALTER TABLE users
            ADD COLUMN user_type TEXT NOT NULL DEFAULT 'PF';
    END IF;
END $$;

DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1 FROM pg_constraint WHERE conname = 'users_user_type_check'
    ) THEN
        ALTER TABLE users
            ADD CONSTRAINT users_user_type_check
            CHECK (user_type IN ('PF', 'PJ'));
    END IF;
END $$;

CREATE TABLE IF NOT EXISTS roles (
    code TEXT PRIMARY KEY,
    description TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS user_roles (
    user_id UUID NOT NULL REFERENCES users(id),
    role_code TEXT NOT NULL REFERENCES roles(code),
    granted_by UUID REFERENCES users(id),
    granted_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    PRIMARY KEY (user_id, role_code)
);

CREATE TABLE IF NOT EXISTS role_separation_rules (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    role_code_a TEXT NOT NULL REFERENCES roles(code),
    role_code_b TEXT NOT NULL REFERENCES roles(code),
    reason TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS regulatory_profiles (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL UNIQUE REFERENCES users(id),
    jurisdiction_code TEXT NOT NULL,
    jurisdiction_risk TEXT NOT NULL DEFAULT 'LOW',
    aml_tier TEXT NOT NULL DEFAULT 'BASIC',
    travel_rule_required BOOLEAN NOT NULL DEFAULT FALSE,
    sanctions_screening_required BOOLEAN NOT NULL DEFAULT FALSE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1 FROM pg_constraint WHERE conname = 'regulatory_profiles_risk_check'
    ) THEN
        ALTER TABLE regulatory_profiles
            ADD CONSTRAINT regulatory_profiles_risk_check
            CHECK (jurisdiction_risk IN ('LOW', 'MEDIUM', 'HIGH'));
    END IF;
    IF NOT EXISTS (
        SELECT 1 FROM pg_constraint WHERE conname = 'regulatory_profiles_aml_check'
    ) THEN
        ALTER TABLE regulatory_profiles
            ADD CONSTRAINT regulatory_profiles_aml_check
            CHECK (aml_tier IN ('BASIC', 'ENHANCED', 'INSTITUTIONAL'));
    END IF;
END $$;

INSERT INTO roles (code, description)
VALUES
    ('ADMIN', 'Acesso administrativo completo'),
    ('COMPLIANCE', 'Operacoes de compliance e risco'),
    ('AUDIT', 'Acesso de auditoria e trilhas'),
    ('OPS', 'Operacoes e reconciliacao'),
    ('VIEWER', 'Leitura e acompanhamento')
ON CONFLICT (code) DO NOTHING;

CREATE INDEX IF NOT EXISTS user_roles_user_idx
    ON user_roles (user_id);

CREATE INDEX IF NOT EXISTS user_roles_role_idx
    ON user_roles (role_code);

CREATE UNIQUE INDEX IF NOT EXISTS role_separation_unique_idx
    ON role_separation_rules (role_code_a, role_code_b);

CREATE TABLE IF NOT EXISTS refresh_tokens (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id),
    token_hash TEXT NOT NULL UNIQUE,
    expires_at TIMESTAMPTZ NOT NULL,
    revoked_at TIMESTAMPTZ,
    replaced_by UUID,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS login_audits (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID REFERENCES users(id),
    email TEXT,
    ip TEXT,
    success BOOLEAN NOT NULL,
    reason TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS pre_registrations (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    full_name TEXT NOT NULL,
    email TEXT NOT NULL,
    phone TEXT NOT NULL,
    status TEXT NOT NULL,
    email_status TEXT NOT NULL,
    phone_status TEXT NOT NULL,
    email_token_hash TEXT NOT NULL,
    phone_code_hash TEXT NOT NULL,
    email_verified_at TIMESTAMPTZ,
    phone_verified_at TIMESTAMPTZ,
    expires_at TIMESTAMPTZ NOT NULL,
    email_attempts INT NOT NULL DEFAULT 0,
    phone_attempts INT NOT NULL DEFAULT 0,
    email_blocked_until TIMESTAMPTZ,
    phone_blocked_until TIMESTAMPTZ,
    created_ip TEXT,
    user_agent TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS pre_registration_attempts (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    pre_registration_id UUID NOT NULL REFERENCES pre_registrations(id),
    channel TEXT NOT NULL,
    success BOOLEAN NOT NULL,
    reason TEXT,
    ip TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS kyc_profiles (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL UNIQUE REFERENCES users(id),
    level TEXT NOT NULL,
    status TEXT NOT NULL,
    provider_ref TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS kyc_limits (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    level TEXT NOT NULL UNIQUE,
    daily_limit BIGINT NOT NULL,
    monthly_limit BIGINT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT kyc_limits_daily_check CHECK (daily_limit >= 0),
    CONSTRAINT kyc_limits_monthly_check CHECK (monthly_limit >= 0)
);

CREATE TABLE IF NOT EXISTS accounts (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id),
    currency TEXT NOT NULL,
    scale SMALLINT NOT NULL,
    balance BIGINT NOT NULL DEFAULT 0,
    status TEXT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT accounts_scale_check CHECK (scale >= 0),
    CONSTRAINT accounts_balance_check CHECK (balance >= 0)
);

CREATE TABLE IF NOT EXISTS transactions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    account_id UUID NOT NULL REFERENCES accounts(id),
    user_id UUID NOT NULL REFERENCES users(id),
    type TEXT NOT NULL,
    status TEXT NOT NULL,
    amount BIGINT NOT NULL,
    fee BIGINT NOT NULL DEFAULT 0,
    net_amount BIGINT NOT NULL,
    idempotency_key TEXT NOT NULL,
    external_ref TEXT,
    reversal_of_transaction_id UUID REFERENCES transactions(id),
    occurred_at TIMESTAMPTZ NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT transactions_amount_check CHECK (amount > 0),
    CONSTRAINT transactions_fee_check CHECK (fee >= 0 AND fee <= amount),
    CONSTRAINT transactions_net_amount_check CHECK (net_amount >= 0 AND net_amount = amount - fee)
);

ALTER TABLE transactions
    ADD COLUMN IF NOT EXISTS reversal_of_transaction_id UUID REFERENCES transactions(id);

CREATE TABLE IF NOT EXISTS conversion_audits (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id),
    operation_type TEXT NOT NULL,
    asset TEXT NOT NULL,
    gross_amount BIGINT NOT NULL,
    fee BIGINT NOT NULL DEFAULT 0,
    net_amount BIGINT NOT NULL,
    quote_price BIGINT NOT NULL,
    spread_bps BIGINT NOT NULL DEFAULT 0,
    liquidity_source TEXT NOT NULL,
    quoted_at TIMESTAMPTZ,
    related_type TEXT,
    related_id UUID,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT conversion_audits_amount_check CHECK (gross_amount >= 0 AND fee >= 0 AND net_amount >= 0),
    CONSTRAINT conversion_audits_quote_check CHECK (quote_price >= 0),
    CONSTRAINT conversion_audits_spread_check CHECK (spread_bps >= 0)
);

CREATE OR REPLACE FUNCTION prevent_transactions_mutation() RETURNS trigger AS $$
BEGIN
    IF TG_OP = 'DELETE' THEN
        RAISE EXCEPTION 'transactions are immutable';
    END IF;

    IF NEW.id IS DISTINCT FROM OLD.id
        OR NEW.account_id IS DISTINCT FROM OLD.account_id
        OR NEW.user_id IS DISTINCT FROM OLD.user_id
        OR NEW.type IS DISTINCT FROM OLD.type
        OR NEW.amount IS DISTINCT FROM OLD.amount
        OR NEW.fee IS DISTINCT FROM OLD.fee
        OR NEW.net_amount IS DISTINCT FROM OLD.net_amount
        OR NEW.idempotency_key IS DISTINCT FROM OLD.idempotency_key
        OR NEW.external_ref IS DISTINCT FROM OLD.external_ref
        OR NEW.reversal_of_transaction_id IS DISTINCT FROM OLD.reversal_of_transaction_id
        OR NEW.occurred_at IS DISTINCT FROM OLD.occurred_at
        OR NEW.created_at IS DISTINCT FROM OLD.created_at
    THEN
        RAISE EXCEPTION 'transactions are immutable';
    END IF;

    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1 FROM pg_trigger WHERE tgname = 'transactions_immutable_trg'
    ) THEN
        CREATE TRIGGER transactions_immutable_trg
            BEFORE UPDATE OR DELETE ON transactions
            FOR EACH ROW EXECUTE FUNCTION prevent_transactions_mutation();
    END IF;
END $$;

CREATE TABLE IF NOT EXISTS audit_logs (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID REFERENCES users(id),
    action TEXT NOT NULL,
    entity_type TEXT,
    entity_id UUID,
    data JSONB,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS audit_logs_archive (
    id UUID PRIMARY KEY,
    user_id UUID,
    action TEXT NOT NULL,
    entity_type TEXT,
    entity_id UUID,
    data JSONB,
    created_at TIMESTAMPTZ NOT NULL
);

CREATE OR REPLACE FUNCTION prevent_transaction_mutation()
RETURNS trigger AS $$
BEGIN
    RAISE EXCEPTION 'transactions are append-only';
END;
$$ LANGUAGE plpgsql;

DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1 FROM pg_trigger WHERE tgname = 'transactions_prevent_update'
    ) THEN
        CREATE TRIGGER transactions_prevent_update
        BEFORE UPDATE OF amount, fee, net_amount, account_id, user_id, type, idempotency_key,
                     external_ref, reversal_of_transaction_id, occurred_at, created_at
        ON transactions
        FOR EACH ROW
        EXECUTE FUNCTION prevent_transaction_mutation();
    END IF;
    IF NOT EXISTS (
        SELECT 1 FROM pg_trigger WHERE tgname = 'transactions_prevent_delete'
    ) THEN
        CREATE TRIGGER transactions_prevent_delete
        BEFORE DELETE ON transactions
        FOR EACH ROW
        EXECUTE FUNCTION prevent_transaction_mutation();
    END IF;
END $$;

CREATE TABLE IF NOT EXISTS compliance_cases (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id),
    type TEXT NOT NULL,
    status TEXT NOT NULL,
    risk_level TEXT NOT NULL,
    title TEXT,
    summary TEXT,
    metadata JSONB,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS compliance_events (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    case_id UUID NOT NULL REFERENCES compliance_cases(id),
    event_type TEXT NOT NULL,
    payload JSONB,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS pix_transfers (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    transaction_id UUID REFERENCES transactions(id),
    user_id UUID NOT NULL REFERENCES users(id),
    account_id UUID NOT NULL REFERENCES accounts(id),
    direction TEXT NOT NULL,
    status TEXT NOT NULL,
    amount BIGINT NOT NULL,
    fee BIGINT NOT NULL DEFAULT 0,
    net_amount BIGINT NOT NULL,
    idempotency_key TEXT NOT NULL,
    end_to_end_id TEXT,
    external_ref TEXT,
    metadata JSONB,
    occurred_at TIMESTAMPTZ NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    confirmed_at TIMESTAMPTZ,
    CONSTRAINT pix_amount_check CHECK (amount > 0),
    CONSTRAINT pix_fee_check CHECK (fee >= 0 AND fee <= amount),
    CONSTRAINT pix_net_amount_check CHECK (net_amount >= 0 AND net_amount = amount - fee)
);

CREATE TABLE IF NOT EXISTS pix_keys (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id),
    account_id UUID NOT NULL REFERENCES accounts(id),
    type TEXT NOT NULL,
    key TEXT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT pix_key_type_check CHECK (type IN ('CPF', 'EMAIL', 'PHONE', 'EVP'))
);

CREATE TABLE IF NOT EXISTS payments (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id),
    account_id UUID NOT NULL REFERENCES accounts(id),
    status TEXT NOT NULL,
    amount BIGINT NOT NULL,
    fee BIGINT NOT NULL DEFAULT 0,
    net_amount BIGINT NOT NULL,
    idempotency_key TEXT NOT NULL,
    barcode TEXT,
    scheduled_at TIMESTAMPTZ,
    due_date DATE,
    external_ref TEXT,
    transaction_id UUID REFERENCES transactions(id),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT payments_amount_check CHECK (amount > 0),
    CONSTRAINT payments_fee_check CHECK (fee >= 0 AND fee <= amount),
    CONSTRAINT payments_net_amount_check CHECK (net_amount >= 0 AND net_amount = amount - fee)
);

CREATE TABLE IF NOT EXISTS card_authorizations (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id),
    account_id UUID NOT NULL REFERENCES accounts(id),
    status TEXT NOT NULL,
    amount BIGINT NOT NULL,
    fee BIGINT NOT NULL DEFAULT 0,
    net_amount BIGINT NOT NULL,
    merchant_name TEXT,
    merchant_mcc TEXT,
    auth_code TEXT,
    external_ref TEXT,
    transaction_id UUID REFERENCES transactions(id),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT card_amount_check CHECK (amount > 0),
    CONSTRAINT card_fee_check CHECK (fee >= 0 AND fee <= amount),
    CONSTRAINT card_net_amount_check CHECK (net_amount >= 0 AND net_amount = amount - fee)
);

CREATE TABLE IF NOT EXISTS trade_orders (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id),
    status TEXT NOT NULL,
    side TEXT NOT NULL,
    base_currency TEXT NOT NULL,
    quote_currency TEXT NOT NULL,
    price BIGINT NOT NULL,
    quantity BIGINT NOT NULL,
    fee BIGINT NOT NULL DEFAULT 0,
    idempotency_key TEXT NOT NULL,
    external_ref TEXT,
    debit_account_id UUID REFERENCES accounts(id),
    credit_account_id UUID REFERENCES accounts(id),
    debit_transaction_id UUID REFERENCES transactions(id),
    credit_transaction_id UUID REFERENCES transactions(id),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT trade_price_check CHECK (price > 0),
    CONSTRAINT trade_quantity_check CHECK (quantity > 0),
    CONSTRAINT trade_fee_check CHECK (fee >= 0)
);

CREATE TABLE IF NOT EXISTS conversion_audits (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id),
    operation_type TEXT NOT NULL,
    asset TEXT NOT NULL,
    gross_amount BIGINT NOT NULL,
    fee BIGINT NOT NULL DEFAULT 0,
    net_amount BIGINT NOT NULL,
    quote_price BIGINT NOT NULL,
    spread_bps BIGINT NOT NULL DEFAULT 0,
    liquidity_source TEXT NOT NULL,
    quoted_at TIMESTAMPTZ,
    related_type TEXT,
    related_id UUID,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT conversion_audits_amount_check CHECK (gross_amount > 0),
    CONSTRAINT conversion_audits_fee_check CHECK (fee >= 0 AND fee <= gross_amount),
    CONSTRAINT conversion_audits_net_amount_check CHECK (net_amount >= 0 AND net_amount = gross_amount - fee),
    CONSTRAINT conversion_audits_quote_check CHECK (quote_price > 0),
    CONSTRAINT conversion_audits_spread_check CHECK (spread_bps >= 0)
);

CREATE TABLE IF NOT EXISTS user_settings (
    user_id UUID PRIMARY KEY REFERENCES users(id),
    conversion_priority TEXT[] NOT NULL DEFAULT ARRAY[]::TEXT[],
    allow_crypto_to_fiat BOOLEAN NOT NULL DEFAULT TRUE,
    auto_convert_pix_in BOOLEAN NOT NULL DEFAULT FALSE,
    pix_in_target_asset TEXT,
    pix_in_percentage SMALLINT NOT NULL DEFAULT 100,
    auto_convert_enabled BOOLEAN NOT NULL DEFAULT FALSE,
    auto_convert_asset TEXT,
    auto_convert_min_amount BIGINT NOT NULL DEFAULT 0,
    fallback_asset TEXT,
    ux_mode TEXT NOT NULL DEFAULT 'LITE',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1 FROM information_schema.columns
         WHERE table_name = 'user_settings'
           AND column_name = 'conversion_priority'
    ) THEN
        ALTER TABLE user_settings
            ADD COLUMN conversion_priority TEXT[] NOT NULL DEFAULT ARRAY[]::TEXT[];
    END IF;
    IF NOT EXISTS (
        SELECT 1 FROM information_schema.columns
         WHERE table_name = 'user_settings'
           AND column_name = 'allow_crypto_to_fiat'
    ) THEN
        ALTER TABLE user_settings
            ADD COLUMN allow_crypto_to_fiat BOOLEAN NOT NULL DEFAULT TRUE;
    END IF;
    IF NOT EXISTS (
        SELECT 1 FROM information_schema.columns
         WHERE table_name = 'user_settings'
           AND column_name = 'auto_convert_pix_in'
    ) THEN
        ALTER TABLE user_settings
            ADD COLUMN auto_convert_pix_in BOOLEAN NOT NULL DEFAULT FALSE;
    END IF;
    IF NOT EXISTS (
        SELECT 1 FROM information_schema.columns
         WHERE table_name = 'user_settings'
           AND column_name = 'pix_in_target_asset'
    ) THEN
        ALTER TABLE user_settings
            ADD COLUMN pix_in_target_asset TEXT;
    END IF;
    IF NOT EXISTS (
        SELECT 1 FROM information_schema.columns
         WHERE table_name = 'user_settings'
           AND column_name = 'pix_in_percentage'
    ) THEN
        ALTER TABLE user_settings
            ADD COLUMN pix_in_percentage SMALLINT NOT NULL DEFAULT 100;
    END IF;
    IF NOT EXISTS (
        SELECT 1 FROM information_schema.columns
         WHERE table_name = 'user_settings'
           AND column_name = 'auto_convert_enabled'
    ) THEN
        ALTER TABLE user_settings
            ADD COLUMN auto_convert_enabled BOOLEAN NOT NULL DEFAULT FALSE;
    END IF;
    IF NOT EXISTS (
        SELECT 1 FROM information_schema.columns
         WHERE table_name = 'user_settings'
           AND column_name = 'auto_convert_asset'
    ) THEN
        ALTER TABLE user_settings
            ADD COLUMN auto_convert_asset TEXT;
    END IF;
    IF NOT EXISTS (
        SELECT 1 FROM information_schema.columns
         WHERE table_name = 'user_settings'
           AND column_name = 'auto_convert_min_amount'
    ) THEN
        ALTER TABLE user_settings
            ADD COLUMN auto_convert_min_amount BIGINT NOT NULL DEFAULT 0;
    END IF;
    IF NOT EXISTS (
        SELECT 1 FROM information_schema.columns
         WHERE table_name = 'user_settings'
           AND column_name = 'fallback_asset'
    ) THEN
        ALTER TABLE user_settings
            ADD COLUMN fallback_asset TEXT;
    END IF;
    IF NOT EXISTS (
        SELECT 1 FROM information_schema.columns
         WHERE table_name = 'user_settings'
           AND column_name = 'ux_mode'
    ) THEN
        ALTER TABLE user_settings
            ADD COLUMN ux_mode TEXT NOT NULL DEFAULT 'LITE';
    END IF;
END $$;

DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1 FROM pg_constraint WHERE conname = 'user_settings_auto_convert_min_check'
    ) THEN
        ALTER TABLE user_settings
            ADD CONSTRAINT user_settings_auto_convert_min_check
            CHECK (auto_convert_min_amount >= 0);
    END IF;
    IF NOT EXISTS (
        SELECT 1 FROM pg_constraint WHERE conname = 'user_settings_ux_mode_check'
    ) THEN
        ALTER TABLE user_settings
            ADD CONSTRAINT user_settings_ux_mode_check
            CHECK (ux_mode IN ('LITE', 'PRO', 'ADVANCED'));
    END IF;
END $$;

CREATE TABLE IF NOT EXISTS plans (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    code TEXT NOT NULL,
    description TEXT,
    monthly_price_cents BIGINT NOT NULL DEFAULT 0,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE UNIQUE INDEX IF NOT EXISTS plans_code_uq
    ON plans (code);

DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1 FROM information_schema.columns
         WHERE table_name = 'plans'
           AND column_name = 'status'
    ) THEN
        ALTER TABLE plans
            ADD COLUMN status TEXT NOT NULL DEFAULT 'ACTIVE';
    END IF;
    IF NOT EXISTS (
        SELECT 1 FROM information_schema.columns
         WHERE table_name = 'plans'
           AND column_name = 'valid_from'
    ) THEN
        ALTER TABLE plans
            ADD COLUMN valid_from TIMESTAMPTZ NOT NULL DEFAULT NOW();
    END IF;
    IF NOT EXISTS (
        SELECT 1 FROM information_schema.columns
         WHERE table_name = 'plans'
           AND column_name = 'valid_until'
    ) THEN
        ALTER TABLE plans
            ADD COLUMN valid_until TIMESTAMPTZ;
    END IF;
END $$;

DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1 FROM pg_constraint WHERE conname = 'plans_status_check'
    ) THEN
        ALTER TABLE plans
            ADD CONSTRAINT plans_status_check
            CHECK (status IN ('ACTIVE', 'INACTIVE'));
    END IF;
END $$;

CREATE TABLE IF NOT EXISTS plan_limits (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    plan_id UUID NOT NULL REFERENCES plans(id),
    limit_code TEXT NOT NULL,
    limit_value BIGINT NOT NULL,
    limit_window TEXT NOT NULL DEFAULT 'MONTHLY',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE UNIQUE INDEX IF NOT EXISTS plan_limits_unique_idx
    ON plan_limits (plan_id, limit_code, limit_window);

CREATE TABLE IF NOT EXISTS plan_history (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    plan_id UUID NOT NULL REFERENCES plans(id),
    code TEXT NOT NULL,
    description TEXT,
    monthly_price_cents BIGINT NOT NULL DEFAULT 0,
    status TEXT NOT NULL DEFAULT 'ACTIVE',
    valid_from TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    valid_until TIMESTAMPTZ,
    change_type TEXT NOT NULL DEFAULT 'UPSERT',
    change_reason TEXT,
    changed_by UUID REFERENCES users(id),
    changed_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS plan_feature_history (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    plan_id UUID NOT NULL REFERENCES plans(id),
    feature_code TEXT NOT NULL,
    enabled BOOLEAN NOT NULL DEFAULT TRUE,
    change_type TEXT NOT NULL DEFAULT 'UPSERT',
    change_reason TEXT,
    changed_by UUID REFERENCES users(id),
    changed_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS plan_limit_history (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    plan_id UUID NOT NULL REFERENCES plans(id),
    limit_code TEXT NOT NULL,
    limit_value BIGINT NOT NULL,
    limit_window TEXT NOT NULL DEFAULT 'MONTHLY',
    change_type TEXT NOT NULL DEFAULT 'UPSERT',
    change_reason TEXT,
    changed_by UUID REFERENCES users(id),
    changed_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS plan_history_plan_time_idx
    ON plan_history (plan_id, changed_at DESC);

CREATE INDEX IF NOT EXISTS plan_feature_history_plan_time_idx
    ON plan_feature_history (plan_id, changed_at DESC);

CREATE INDEX IF NOT EXISTS plan_limit_history_plan_time_idx
    ON plan_limit_history (plan_id, changed_at DESC);

CREATE TABLE IF NOT EXISTS plans (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    code TEXT NOT NULL,
    description TEXT,
    monthly_price_cents BIGINT NOT NULL DEFAULT 0,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE UNIQUE INDEX IF NOT EXISTS plans_code_uq
    ON plans (code);

CREATE TABLE IF NOT EXISTS user_plans (
    user_id UUID NOT NULL REFERENCES users(id),
    plan_id UUID NOT NULL REFERENCES plans(id),
    valid_from TIMESTAMPTZ NOT NULL,
    valid_until TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS user_plans_user_idx
    ON user_plans (user_id, valid_from DESC);

CREATE TABLE IF NOT EXISTS pricing_versions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    code TEXT NOT NULL UNIQUE,
    description TEXT,
    status TEXT NOT NULL DEFAULT 'ACTIVE',
    valid_from TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    valid_until TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1 FROM pg_constraint WHERE conname = 'pricing_versions_status_check'
    ) THEN
        ALTER TABLE pricing_versions
            ADD CONSTRAINT pricing_versions_status_check
            CHECK (status IN ('ACTIVE', 'INACTIVE'));
    END IF;
END $$;

CREATE TABLE IF NOT EXISTS pricing_rules (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    plan_id UUID NOT NULL REFERENCES plans(id),
    pricing_version_id UUID REFERENCES pricing_versions(id),
    user_type TEXT NOT NULL,
    operation_type TEXT NOT NULL,
    asset TEXT NOT NULL,
    fee_type TEXT NOT NULL,
    fee_value BIGINT NOT NULL,
    min_fee BIGINT,
    max_fee BIGINT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT pricing_rules_fee_value_check CHECK (fee_value >= 0),
    CONSTRAINT pricing_rules_min_fee_check CHECK (min_fee IS NULL OR min_fee >= 0),
    CONSTRAINT pricing_rules_max_fee_check CHECK (max_fee IS NULL OR max_fee >= 0)
);

DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1 FROM pg_constraint WHERE conname = 'pricing_rules_user_type_check'
    ) THEN
        ALTER TABLE pricing_rules
            ADD CONSTRAINT pricing_rules_user_type_check
            CHECK (user_type IN ('PF', 'PJ'));
    END IF;
    IF NOT EXISTS (
        SELECT 1 FROM pg_constraint WHERE conname = 'pricing_rules_operation_type_check'
    ) THEN
        ALTER TABLE pricing_rules
            ADD CONSTRAINT pricing_rules_operation_type_check
            CHECK (operation_type IN ('PIX_IN', 'PIX_OUT', 'PIX_TO_CRYPTO', 'CRYPTO_TO_PIX', 'SWAP', 'CARD_CRYPTO', 'INVOICE'));
    END IF;
    IF NOT EXISTS (
        SELECT 1 FROM pg_constraint WHERE conname = 'pricing_rules_asset_check'
    ) THEN
        ALTER TABLE pricing_rules
            ADD CONSTRAINT pricing_rules_asset_check
            CHECK (asset IN ('BRL', 'BTC', 'ETH', 'USDT', 'ANY'));
    END IF;
    IF NOT EXISTS (
        SELECT 1 FROM pg_constraint WHERE conname = 'pricing_rules_fee_type_check'
    ) THEN
        ALTER TABLE pricing_rules
            ADD CONSTRAINT pricing_rules_fee_type_check
            CHECK (fee_type IN ('PERCENTAGE', 'FIXED'));
    END IF;
END $$;

CREATE INDEX IF NOT EXISTS pricing_rules_lookup_idx
    ON pricing_rules (plan_id, user_type, operation_type, asset);

CREATE INDEX IF NOT EXISTS pricing_rules_version_idx
    ON pricing_rules (pricing_version_id);

INSERT INTO pricing_versions (id, code, description, status, valid_from, created_at)
VALUES (gen_random_uuid(), 'DEFAULT', 'Versao inicial de pricing', 'ACTIVE', NOW(), NOW())
ON CONFLICT (code) DO NOTHING;

UPDATE pricing_rules
   SET pricing_version_id = pv.id
  FROM pricing_versions pv
 WHERE pv.code = 'DEFAULT'
   AND pricing_rules.pricing_version_id IS NULL;

CREATE TABLE IF NOT EXISTS plan_features (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    plan_id UUID NOT NULL REFERENCES plans(id),
    feature_code TEXT NOT NULL,
    enabled BOOLEAN NOT NULL DEFAULT TRUE,
    metadata JSONB,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE UNIQUE INDEX IF NOT EXISTS plan_features_unique_idx
    ON plan_features (plan_id, feature_code);

CREATE TABLE IF NOT EXISTS user_feature_overrides (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id),
    feature_code TEXT NOT NULL,
    enabled BOOLEAN NOT NULL DEFAULT TRUE,
    reason TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE UNIQUE INDEX IF NOT EXISTS user_feature_overrides_unique_idx
    ON user_feature_overrides (user_id, feature_code);

CREATE TABLE IF NOT EXISTS user_limit_overrides (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id),
    limit_code TEXT NOT NULL,
    limit_value BIGINT NOT NULL DEFAULT 0,
    limit_window TEXT NOT NULL DEFAULT 'MONTHLY',
    reason TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT user_limit_value_check CHECK (limit_value >= 0)
);

CREATE UNIQUE INDEX IF NOT EXISTS user_limit_overrides_unique_idx
    ON user_limit_overrides (user_id, limit_code, limit_window);

CREATE TABLE IF NOT EXISTS pricing_campaigns (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    code TEXT NOT NULL UNIQUE,
    description TEXT,
    status TEXT NOT NULL DEFAULT 'ACTIVE',
    priority INTEGER NOT NULL DEFAULT 0,
    valid_from TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    valid_until TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1 FROM pg_constraint WHERE conname = 'pricing_campaigns_status_check'
    ) THEN
        ALTER TABLE pricing_campaigns
            ADD CONSTRAINT pricing_campaigns_status_check
            CHECK (status IN ('ACTIVE', 'INACTIVE'));
    END IF;
END $$;

CREATE TABLE IF NOT EXISTS pricing_campaign_rules (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    campaign_id UUID NOT NULL REFERENCES pricing_campaigns(id),
    plan_id UUID REFERENCES plans(id),
    user_type TEXT NOT NULL,
    operation_type TEXT NOT NULL,
    asset TEXT NOT NULL,
    fee_type TEXT NOT NULL,
    fee_value BIGINT NOT NULL,
    min_fee BIGINT,
    max_fee BIGINT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT pricing_campaign_rules_fee_value_check CHECK (fee_value >= 0),
    CONSTRAINT pricing_campaign_rules_min_fee_check CHECK (min_fee IS NULL OR min_fee >= 0),
    CONSTRAINT pricing_campaign_rules_max_fee_check CHECK (max_fee IS NULL OR max_fee >= 0)
);

DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1 FROM pg_constraint WHERE conname = 'pricing_campaign_rules_user_type_check'
    ) THEN
        ALTER TABLE pricing_campaign_rules
            ADD CONSTRAINT pricing_campaign_rules_user_type_check
            CHECK (user_type IN ('PF', 'PJ'));
    END IF;
    IF NOT EXISTS (
        SELECT 1 FROM pg_constraint WHERE conname = 'pricing_campaign_rules_operation_type_check'
    ) THEN
        ALTER TABLE pricing_campaign_rules
            ADD CONSTRAINT pricing_campaign_rules_operation_type_check
            CHECK (operation_type IN ('PIX_IN', 'PIX_OUT', 'PIX_TO_CRYPTO', 'CRYPTO_TO_PIX', 'SWAP', 'CARD_CRYPTO', 'INVOICE'));
    END IF;
    IF NOT EXISTS (
        SELECT 1 FROM pg_constraint WHERE conname = 'pricing_campaign_rules_asset_check'
    ) THEN
        ALTER TABLE pricing_campaign_rules
            ADD CONSTRAINT pricing_campaign_rules_asset_check
            CHECK (asset IN ('BRL', 'BTC', 'ETH', 'USDT', 'ANY'));
    END IF;
    IF NOT EXISTS (
        SELECT 1 FROM pg_constraint WHERE conname = 'pricing_campaign_rules_fee_type_check'
    ) THEN
        ALTER TABLE pricing_campaign_rules
            ADD CONSTRAINT pricing_campaign_rules_fee_type_check
            CHECK (fee_type IN ('PERCENTAGE', 'FIXED'));
    END IF;
END $$;

CREATE INDEX IF NOT EXISTS pricing_campaign_rules_lookup_idx
    ON pricing_campaign_rules (campaign_id, user_type, operation_type, asset);

INSERT INTO plans (id, code, description, monthly_price_cents, created_at)
VALUES
    (gen_random_uuid(), 'FREE', 'Plano gratuito', 0, NOW()),
    (gen_random_uuid(), 'PRO', 'Plano profissional', 9900, NOW()),
    (gen_random_uuid(), 'BUSINESS', 'Plano corporativo', 29900, NOW())
ON CONFLICT (code) DO NOTHING;

INSERT INTO plan_features (id, plan_id, feature_code, enabled, created_at, updated_at)
SELECT gen_random_uuid(), p.id, v.feature_code, v.enabled, NOW(), NOW()
FROM plans p
JOIN (
    VALUES
        ('FREE', 'PAYMENTS', TRUE),
        ('FREE', 'PIX_OUT', TRUE),
        ('FREE', 'PIX_TO_CRYPTO', TRUE),
        ('FREE', 'CRYPTO_TO_PIX', TRUE),
        ('FREE', 'SWAP', TRUE),
        ('FREE', 'CARD_CRYPTO', TRUE),
        ('FREE', 'INVOICE', TRUE),
        ('FREE', 'UX_ADVANCED', FALSE),
        ('PRO', 'PAYMENTS', TRUE),
        ('PRO', 'PIX_OUT', TRUE),
        ('PRO', 'PIX_TO_CRYPTO', TRUE),
        ('PRO', 'CRYPTO_TO_PIX', TRUE),
        ('PRO', 'SWAP', TRUE),
        ('PRO', 'CARD_CRYPTO', TRUE),
        ('PRO', 'INVOICE', TRUE),
        ('PRO', 'UX_ADVANCED', TRUE),
        ('BUSINESS', 'PAYMENTS', TRUE),
        ('BUSINESS', 'PIX_OUT', TRUE),
        ('BUSINESS', 'PIX_TO_CRYPTO', TRUE),
        ('BUSINESS', 'CRYPTO_TO_PIX', TRUE),
        ('BUSINESS', 'SWAP', TRUE),
        ('BUSINESS', 'CARD_CRYPTO', TRUE),
        ('BUSINESS', 'INVOICE', TRUE),
        ('BUSINESS', 'UX_ADVANCED', TRUE)
) AS v(plan_code, feature_code, enabled)
  ON p.code = v.plan_code
ON CONFLICT (plan_id, feature_code) DO NOTHING;

INSERT INTO plan_limits (id, plan_id, limit_code, limit_value, limit_window, created_at, updated_at)
SELECT gen_random_uuid(), p.id, v.limit_code, v.limit_value, v.limit_window, NOW(), NOW()
FROM plans p
JOIN (
    VALUES
        ('FREE', 'MONTHLY_VOLUME', 5000000, 'MONTHLY'),
        ('FREE', 'AUTOMATIONS_MONTHLY', 20, 'MONTHLY'),
        ('FREE', 'HISTORY_ITEMS', 200, 'TOTAL'),
        ('FREE', 'HISTORY_DAYS', 30, 'TOTAL'),
        ('PRO', 'MONTHLY_VOLUME', 50000000, 'MONTHLY'),
        ('PRO', 'AUTOMATIONS_MONTHLY', 200, 'MONTHLY'),
        ('PRO', 'HISTORY_ITEMS', 2000, 'TOTAL'),
        ('PRO', 'HISTORY_DAYS', 365, 'TOTAL'),
        ('BUSINESS', 'MONTHLY_VOLUME', 500000000, 'MONTHLY'),
        ('BUSINESS', 'AUTOMATIONS_MONTHLY', 2000, 'MONTHLY'),
        ('BUSINESS', 'HISTORY_ITEMS', 10000, 'TOTAL'),
        ('BUSINESS', 'HISTORY_DAYS', 3650, 'TOTAL')
) AS v(plan_code, limit_code, limit_value, limit_window)
  ON p.code = v.plan_code
ON CONFLICT (plan_id, limit_code, limit_window) DO NOTHING;

INSERT INTO pricing_rules (id, plan_id, user_type, operation_type, asset, fee_type, fee_value, min_fee, max_fee, created_at)
SELECT gen_random_uuid(), p.id, v.user_type, v.operation_type, v.asset, v.fee_type,
       v.fee_value::BIGINT, v.min_fee::BIGINT, v.max_fee::BIGINT, NOW()
FROM plans p
CROSS JOIN (
    VALUES
        ('PF', 'PIX_IN', 'ANY', 'FIXED', 0, NULL, NULL),
        ('PF', 'PIX_OUT', 'ANY', 'PERCENTAGE', 30, 50, NULL),
        ('PF', 'PIX_TO_CRYPTO', 'ANY', 'PERCENTAGE', 40, 100, NULL),
        ('PF', 'CRYPTO_TO_PIX', 'ANY', 'PERCENTAGE', 40, 100, NULL),
        ('PF', 'SWAP', 'ANY', 'PERCENTAGE', 35, 100, NULL),
        ('PF', 'CARD_CRYPTO', 'ANY', 'PERCENTAGE', 50, 150, NULL),
        ('PF', 'INVOICE', 'ANY', 'PERCENTAGE', 20, 50, NULL),
        ('PJ', 'PIX_IN', 'ANY', 'FIXED', 0, NULL, NULL),
        ('PJ', 'PIX_OUT', 'ANY', 'PERCENTAGE', 30, 50, NULL),
        ('PJ', 'PIX_TO_CRYPTO', 'ANY', 'PERCENTAGE', 40, 100, NULL),
        ('PJ', 'CRYPTO_TO_PIX', 'ANY', 'PERCENTAGE', 40, 100, NULL),
        ('PJ', 'SWAP', 'ANY', 'PERCENTAGE', 35, 100, NULL),
        ('PJ', 'CARD_CRYPTO', 'ANY', 'PERCENTAGE', 50, 150, NULL),
        ('PJ', 'INVOICE', 'ANY', 'PERCENTAGE', 20, 50, NULL)
) AS v(user_type, operation_type, asset, fee_type, fee_value, min_fee, max_fee)
WHERE p.code = 'FREE'
  AND NOT EXISTS (
    SELECT 1
      FROM pricing_rules pr
     WHERE pr.plan_id = p.id
       AND pr.user_type = v.user_type
       AND pr.operation_type = v.operation_type
       AND pr.asset = v.asset
  );

INSERT INTO pricing_rules (id, plan_id, user_type, operation_type, asset, fee_type, fee_value, min_fee, max_fee, created_at)
SELECT gen_random_uuid(), p.id, v.user_type, v.operation_type, v.asset, v.fee_type,
       v.fee_value::BIGINT, v.min_fee::BIGINT, v.max_fee::BIGINT, NOW()
FROM plans p
CROSS JOIN (
    VALUES
        ('PF', 'PIX_IN', 'ANY', 'FIXED', 0, NULL, NULL),
        ('PF', 'PIX_OUT', 'ANY', 'PERCENTAGE', 15, 25, NULL),
        ('PF', 'PIX_TO_CRYPTO', 'ANY', 'PERCENTAGE', 25, 50, NULL),
        ('PF', 'CRYPTO_TO_PIX', 'ANY', 'PERCENTAGE', 25, 50, NULL),
        ('PF', 'SWAP', 'ANY', 'PERCENTAGE', 20, 50, NULL),
        ('PF', 'CARD_CRYPTO', 'ANY', 'PERCENTAGE', 35, 75, NULL),
        ('PF', 'INVOICE', 'ANY', 'PERCENTAGE', 10, 25, NULL),
        ('PJ', 'PIX_IN', 'ANY', 'FIXED', 0, NULL, NULL),
        ('PJ', 'PIX_OUT', 'ANY', 'PERCENTAGE', 15, 25, NULL),
        ('PJ', 'PIX_TO_CRYPTO', 'ANY', 'PERCENTAGE', 25, 50, NULL),
        ('PJ', 'CRYPTO_TO_PIX', 'ANY', 'PERCENTAGE', 25, 50, NULL),
        ('PJ', 'SWAP', 'ANY', 'PERCENTAGE', 20, 50, NULL),
        ('PJ', 'CARD_CRYPTO', 'ANY', 'PERCENTAGE', 35, 75, NULL),
        ('PJ', 'INVOICE', 'ANY', 'PERCENTAGE', 10, 25, NULL)
) AS v(user_type, operation_type, asset, fee_type, fee_value, min_fee, max_fee)
WHERE p.code = 'PRO'
  AND NOT EXISTS (
    SELECT 1
      FROM pricing_rules pr
     WHERE pr.plan_id = p.id
       AND pr.user_type = v.user_type
       AND pr.operation_type = v.operation_type
       AND pr.asset = v.asset
  );

INSERT INTO pricing_rules (id, plan_id, user_type, operation_type, asset, fee_type, fee_value, min_fee, max_fee, created_at)
SELECT gen_random_uuid(), p.id, v.user_type, v.operation_type, v.asset, v.fee_type,
       v.fee_value::BIGINT, v.min_fee::BIGINT, v.max_fee::BIGINT, NOW()
FROM plans p
CROSS JOIN (
    VALUES
        ('PF', 'PIX_IN', 'ANY', 'FIXED', 0, NULL, NULL),
        ('PF', 'PIX_OUT', 'ANY', 'PERCENTAGE', 10, 10, NULL),
        ('PF', 'PIX_TO_CRYPTO', 'ANY', 'PERCENTAGE', 20, 25, NULL),
        ('PF', 'CRYPTO_TO_PIX', 'ANY', 'PERCENTAGE', 20, 25, NULL),
        ('PF', 'SWAP', 'ANY', 'PERCENTAGE', 15, 25, NULL),
        ('PF', 'CARD_CRYPTO', 'ANY', 'PERCENTAGE', 25, 50, NULL),
        ('PF', 'INVOICE', 'ANY', 'PERCENTAGE', 8, 10, NULL),
        ('PJ', 'PIX_IN', 'ANY', 'FIXED', 0, NULL, NULL),
        ('PJ', 'PIX_OUT', 'ANY', 'PERCENTAGE', 10, 10, NULL),
        ('PJ', 'PIX_TO_CRYPTO', 'ANY', 'PERCENTAGE', 20, 25, NULL),
        ('PJ', 'CRYPTO_TO_PIX', 'ANY', 'PERCENTAGE', 20, 25, NULL),
        ('PJ', 'SWAP', 'ANY', 'PERCENTAGE', 15, 25, NULL),
        ('PJ', 'CARD_CRYPTO', 'ANY', 'PERCENTAGE', 25, 50, NULL),
        ('PJ', 'INVOICE', 'ANY', 'PERCENTAGE', 8, 10, NULL)
) AS v(user_type, operation_type, asset, fee_type, fee_value, min_fee, max_fee)
WHERE p.code = 'BUSINESS'
  AND NOT EXISTS (
    SELECT 1
      FROM pricing_rules pr
     WHERE pr.plan_id = p.id
       AND pr.user_type = v.user_type
       AND pr.operation_type = v.operation_type
       AND pr.asset = v.asset
  );

CREATE TABLE IF NOT EXISTS conversion_rules (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id),
    trigger_event TEXT NOT NULL,
    source_asset TEXT NOT NULL,
    target_asset TEXT NOT NULL,
    enabled BOOLEAN NOT NULL DEFAULT TRUE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT conversion_trigger_check CHECK (trigger_event IN ('PIX_IN', 'CARD_AUTH', 'PAYMENT')),
    CONSTRAINT conversion_source_check CHECK (source_asset IN ('BRL', 'USDT', 'BTC', 'ETH', 'MATIC')),
    CONSTRAINT conversion_target_check CHECK (target_asset IN ('BRL', 'USDT', 'BTC', 'ETH', 'MATIC'))
);
ALTER TABLE conversion_rules
    DROP CONSTRAINT IF EXISTS conversion_trigger_check;
ALTER TABLE conversion_rules
    ADD CONSTRAINT conversion_trigger_check CHECK (trigger_event IN ('PIX_IN', 'CARD_AUTH', 'PAYMENT'));

CREATE TABLE IF NOT EXISTS invoices (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id),
    amount_brl BIGINT NOT NULL,
    pix_copy_paste TEXT NOT NULL,
    usdt_address TEXT NOT NULL,
    idempotency_key TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT invoice_amount_check CHECK (amount_brl > 0)
);

ALTER TABLE invoices
    ADD COLUMN IF NOT EXISTS idempotency_key TEXT;

CREATE UNIQUE INDEX IF NOT EXISTS invoices_user_idempotency_idx
    ON invoices (user_id, idempotency_key)
    WHERE idempotency_key IS NOT NULL;

CREATE TABLE IF NOT EXISTS crypto_transfers (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id),
    account_id UUID NOT NULL REFERENCES accounts(id),
    transaction_id UUID NOT NULL REFERENCES transactions(id),
    asset TEXT NOT NULL,
    network TEXT NOT NULL,
    address TEXT NOT NULL,
    amount BIGINT NOT NULL,
    fee BIGINT NOT NULL DEFAULT 0,
    status TEXT NOT NULL,
    direction TEXT NOT NULL DEFAULT 'BUY',
    tx_hash TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT crypto_amount_check CHECK (amount > 0),
    CONSTRAINT crypto_fee_check CHECK (fee >= 0),
    CONSTRAINT crypto_status_check CHECK (status IN ('PENDING_EXCHANGE', 'CONFIRMED', 'REJECTED')),
    CONSTRAINT crypto_asset_check CHECK (asset IN ('USDT', 'BTC', 'ETH', 'MATIC')),
    CONSTRAINT crypto_direction_check CHECK (direction IN ('BUY', 'SELL'))
);

CREATE TABLE IF NOT EXISTS outbox_events (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    event_type TEXT NOT NULL,
    payload JSONB NOT NULL,
    status TEXT NOT NULL DEFAULT 'PENDING',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    processed_at TIMESTAMPTZ
);

CREATE TABLE IF NOT EXISTS webhook_events (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    event_type TEXT NOT NULL,
    reference_id TEXT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS webhook_retry_jobs (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    event_type TEXT NOT NULL,
    path TEXT NOT NULL,
    payload JSONB NOT NULL,
    headers JSONB,
    attempts INTEGER NOT NULL DEFAULT 0,
    status TEXT NOT NULL DEFAULT 'PENDING',
    next_retry_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    last_error TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1 FROM pg_constraint WHERE conname = 'webhook_retry_status_check'
    ) THEN
        ALTER TABLE webhook_retry_jobs
            ADD CONSTRAINT webhook_retry_status_check
            CHECK (status IN ('PENDING', 'PROCESSING', 'SUCCEEDED', 'DEAD'));
    END IF;
END $$;

CREATE UNIQUE INDEX IF NOT EXISTS transactions_idempotency_uq
    ON transactions (account_id, idempotency_key);

CREATE INDEX IF NOT EXISTS transactions_account_time_idx
    ON transactions (account_id, occurred_at DESC);

CREATE INDEX IF NOT EXISTS transactions_time_idx
    ON transactions (occurred_at DESC);

CREATE INDEX IF NOT EXISTS transactions_reversal_idx
    ON transactions (reversal_of_transaction_id);

CREATE INDEX IF NOT EXISTS audit_logs_user_time_idx
    ON audit_logs (user_id, created_at DESC);

CREATE INDEX IF NOT EXISTS audit_logs_time_idx
    ON audit_logs (created_at DESC);

CREATE INDEX IF NOT EXISTS audit_logs_archive_user_time_idx
    ON audit_logs_archive (user_id, created_at DESC);

CREATE INDEX IF NOT EXISTS audit_logs_archive_time_idx
    ON audit_logs_archive (created_at DESC);

CREATE INDEX IF NOT EXISTS audit_logs_archive_user_time_idx
    ON audit_logs_archive (user_id, created_at DESC);

CREATE INDEX IF NOT EXISTS audit_logs_archive_time_idx
    ON audit_logs_archive (created_at DESC);

CREATE INDEX IF NOT EXISTS refresh_tokens_user_idx
    ON refresh_tokens (user_id, created_at DESC);

CREATE INDEX IF NOT EXISTS refresh_tokens_expires_idx
    ON refresh_tokens (expires_at);

CREATE INDEX IF NOT EXISTS login_audits_email_ip_time_idx
    ON login_audits (email, ip, created_at DESC);

CREATE INDEX IF NOT EXISTS login_audits_time_idx
    ON login_audits (created_at DESC);

CREATE UNIQUE INDEX IF NOT EXISTS pre_registrations_email_uq
    ON pre_registrations (email);

CREATE UNIQUE INDEX IF NOT EXISTS pre_registrations_phone_uq
    ON pre_registrations (phone);

CREATE INDEX IF NOT EXISTS pre_registrations_status_idx
    ON pre_registrations (status, created_at DESC);

CREATE INDEX IF NOT EXISTS pre_registrations_expiry_idx
    ON pre_registrations (expires_at);

CREATE INDEX IF NOT EXISTS pre_registrations_email_status_idx
    ON pre_registrations (email_status, created_at DESC);

CREATE INDEX IF NOT EXISTS pre_registrations_phone_status_idx
    ON pre_registrations (phone_status, created_at DESC);

CREATE INDEX IF NOT EXISTS pre_registration_attempts_pre_idx
    ON pre_registration_attempts (pre_registration_id, created_at DESC);

CREATE INDEX IF NOT EXISTS compliance_cases_user_idx
    ON compliance_cases (user_id, created_at DESC);

CREATE INDEX IF NOT EXISTS compliance_cases_status_idx
    ON compliance_cases (status, created_at DESC);

CREATE INDEX IF NOT EXISTS compliance_events_case_idx
    ON compliance_events (case_id, created_at DESC);

CREATE UNIQUE INDEX IF NOT EXISTS pix_idempotency_uq
    ON pix_transfers (account_id, idempotency_key);

CREATE UNIQUE INDEX IF NOT EXISTS pix_idempotency_global_uq
    ON pix_transfers (idempotency_key);

CREATE UNIQUE INDEX IF NOT EXISTS pix_end_to_end_uq
    ON pix_transfers (end_to_end_id);

CREATE INDEX IF NOT EXISTS pix_account_time_idx
    ON pix_transfers (account_id, occurred_at DESC);

CREATE INDEX IF NOT EXISTS pix_status_idx
    ON pix_transfers (status);

CREATE INDEX IF NOT EXISTS pix_created_at_idx
    ON pix_transfers (created_at DESC);

CREATE UNIQUE INDEX IF NOT EXISTS pix_key_unique_idx
    ON pix_keys (key);

CREATE INDEX IF NOT EXISTS pix_keys_account_idx
    ON pix_keys (account_id);

CREATE UNIQUE INDEX IF NOT EXISTS payments_idempotency_uq
    ON payments (account_id, idempotency_key);

CREATE INDEX IF NOT EXISTS payments_account_time_idx
    ON payments (account_id, created_at DESC);

CREATE INDEX IF NOT EXISTS card_auth_account_time_idx
    ON card_authorizations (account_id, created_at DESC);

CREATE INDEX IF NOT EXISTS trade_user_time_idx
    ON trade_orders (user_id, created_at DESC);

CREATE UNIQUE INDEX IF NOT EXISTS trade_idempotency_uq
    ON trade_orders (user_id, idempotency_key);

CREATE INDEX IF NOT EXISTS outbox_status_time_idx
    ON outbox_events (status, created_at);

CREATE UNIQUE INDEX IF NOT EXISTS webhook_events_unique_idx
    ON webhook_events (event_type, reference_id);

CREATE INDEX IF NOT EXISTS webhook_retry_next_idx
    ON webhook_retry_jobs (status, next_retry_at);

CREATE INDEX IF NOT EXISTS kyc_limits_level_idx
    ON kyc_limits (level);

CREATE INDEX IF NOT EXISTS crypto_transfers_account_idx
    ON crypto_transfers (account_id);

CREATE INDEX IF NOT EXISTS crypto_transfers_status_idx
    ON crypto_transfers (status);

CREATE INDEX IF NOT EXISTS crypto_transfers_created_at_idx
    ON crypto_transfers (created_at DESC);

CREATE INDEX IF NOT EXISTS conversion_audits_user_time_idx
    ON conversion_audits (user_id, created_at DESC);

CREATE INDEX IF NOT EXISTS conversion_audits_related_idx
    ON conversion_audits (related_type, related_id);

CREATE INDEX IF NOT EXISTS conversion_rules_user_idx
    ON conversion_rules (user_id);

CREATE INDEX IF NOT EXISTS conversion_rules_trigger_idx
    ON conversion_rules (trigger_event, enabled);

CREATE INDEX IF NOT EXISTS invoices_user_idx
    ON invoices (user_id);

CREATE INDEX IF NOT EXISTS conversion_audits_user_time_idx
    ON conversion_audits (user_id, created_at DESC);

CREATE INDEX IF NOT EXISTS conversion_audits_related_idx
    ON conversion_audits (related_id);

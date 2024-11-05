CREATE TABLE IF NOT EXISTS accounts (
    account_id UUID PRIMARY KEY,
    user_id UUID NOT NULL,
    account_number BYTEA NOT NULL,
    phone_number BYTEA NOT NULL,
    account_name VARCHAR(255) NOT NULL,
    currency CHAR(3) NOT NULL,
    created_at TIMESTAMP DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_accounts_user_id_created_at ON accounts (user_id, created_at DESC);

CREATE TABLE IF NOT EXISTS cards (
    card_id UUID PRIMARY KEY,
    account_id UUID NOT NULL,
    user_id UUID NOT NULL,
    card_number BYTEA NOT NULL UNIQUE,
    phone_number BYTEA NOT NULL,
    created_at TIMESTAMP DEFAULT NOW(),
    is_active BOOLEAN DEFAULT TRUE
);

CREATE INDEX IF NOT EXISTS idx_cards_user_id_created_at ON cards (user_id, created_at DESC);

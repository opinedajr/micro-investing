CREATE TABLE positions (
    id TEXT PRIMARY KEY,
    wallet_id TEXT NOT NULL,
    stock_id TEXT NOT NULL,
    quantity INTEGER NOT NULL,
    average_price INTEGER NOT NULL,
    current_price INTEGER NOT NULL DEFAULT 0,
    invested INTEGER NOT NULL DEFAULT 0,
    balance INTEGER NOT NULL DEFAULT 0,
    variation_value INTEGER NOT NULL DEFAULT 0,
    variation_percent REAL NOT NULL DEFAULT 0,
    portfolio_percent REAL NOT NULL DEFAULT 0,
    created_at DATETIME NOT NULL,
    updated_at DATETIME NOT NULL,
    UNIQUE(wallet_id, stock_id)
);

CREATE INDEX idx_positions_wallet_id ON positions(wallet_id);

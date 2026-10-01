CREATE TABLE dividends (
    id TEXT PRIMARY KEY,
    wallet_id TEXT NOT NULL,
    year INTEGER NOT NULL,
    amount INTEGER NOT NULL,
    created_at DATETIME NOT NULL,
    updated_at DATETIME NOT NULL,
    FOREIGN KEY (wallet_id) REFERENCES wallets(id)
);

CREATE UNIQUE INDEX idx_dividends_wallet_year ON dividends(wallet_id, year);

CREATE TABLE assets (
    id TEXT PRIMARY KEY,
    wallet_id TEXT NOT NULL,
    type TEXT NOT NULL,
    date DATETIME NOT NULL,
    description TEXT NOT NULL,
    amount INTEGER NOT NULL,
    created_at DATETIME NOT NULL,
    updated_at DATETIME NOT NULL,
    FOREIGN KEY (wallet_id) REFERENCES wallets(id)
);

CREATE INDEX idx_assets_wallet_id ON assets(wallet_id);
CREATE INDEX idx_assets_wallet_type_date ON assets(wallet_id, type, date);

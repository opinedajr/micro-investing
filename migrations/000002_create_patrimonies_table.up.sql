CREATE TABLE patrimonies (
    id TEXT PRIMARY KEY,
    wallet_id TEXT NOT NULL,
    year INTEGER NOT NULL,
    month INTEGER NOT NULL,
    type TEXT NOT NULL,
    amount INTEGER NOT NULL,
    created_at DATETIME NOT NULL,
    updated_at DATETIME NOT NULL,
    UNIQUE(wallet_id, year, month, type),
    FOREIGN KEY (wallet_id) REFERENCES wallets(id)
);

CREATE INDEX idx_patrimonies_wallet_id ON patrimonies(wallet_id);
CREATE INDEX idx_patrimonies_year_month ON patrimonies(year, month);

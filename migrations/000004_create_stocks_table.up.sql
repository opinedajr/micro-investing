CREATE TABLE stocks (
    id TEXT PRIMARY KEY,
    ticker TEXT NOT NULL,
    name TEXT NOT NULL,
    sector TEXT NOT NULL,
    rank INTEGER NOT NULL,
    website TEXT,
    created_at DATETIME NOT NULL,
    updated_at DATETIME NOT NULL,
    UNIQUE(ticker)
);

CREATE INDEX idx_stocks_ticker ON stocks(ticker);

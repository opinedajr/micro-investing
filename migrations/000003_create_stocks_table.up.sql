CREATE TABLE stocks (
    id TEXT PRIMARY KEY,
    ticker VARCHAR(6) NOT NULL UNIQUE,
    name VARCHAR(100) NOT NULL,
    sector VARCHAR(50) NOT NULL,
    rank TINYINT NOT NULL CHECK(rank >= 0 AND rank <= 10),
    website TEXT,
    created_at DATETIME NOT NULL,
    updated_at DATETIME NOT NULL
);

CREATE INDEX idx_stocks_ticker ON stocks(ticker);

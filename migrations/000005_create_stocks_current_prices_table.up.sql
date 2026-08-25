CREATE TABLE stocks_current_prices (
    stock_id TEXT PRIMARY KEY,
    price INTEGER NOT NULL,
    updated_at DATETIME NOT NULL
);

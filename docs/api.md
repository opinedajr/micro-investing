# API Documentation

## Wallets

### List Wallets
- **URL**: `GET /api/v1/wallets`
- **Response**: `200 OK`

```json
{
  "data": [
    {
      "id": "wallet-id",
      "name": "My Wallet",
      "description": "Stocks and FIIs",
      "created_at": "2026-07-25T12:00:00Z",
      "updated_at": "2026-07-25T12:00:00Z"
    }
  ]
}
```

### Create Wallet
- **URL**: `POST /api/v1/wallets`
- **Request Body**:

```json
{
  "name": "My Wallet",
  "description": "Stocks and FIIs"
}
```

- **Response**: `201 Created`

### Find Wallet
- **URL**: `GET /api/v1/wallets/:id`
- **Response**: `200 OK` or `404 Not Found`

### Update Wallet
- **URL**: `PUT /api/v1/wallets/:id`
- **Request Body**:

```json
{
  "name": "My Wallet Updated",
  "description": "Updated description"
}
```

- **Response**: `200 OK` or `404 Not Found`

### Delete Wallet
- **URL**: `DELETE /api/v1/wallets/:id`
- **Response**: `204 No Content` or `404 Not Found`

---

## Patrimonies

Manual patrimony records consolidated by wallet, year, month and asset type. All monetary values are integer cents (e.g. `150000` means R$ 1.500,00).

Valid asset types (`type`):
- `stocks`
- `fiis`
- `fixed_income`
- `emergency_reserve`
- `liquid_cash`

### List Patrimonies
- **URL**: `GET /api/v1/wallets/:id/patrimonies?type=&year=&month=`
- **Query Parameters** (all optional):
  - `type`: asset type filter
  - `year`: integer year filter
  - `month`: integer month filter (1-12)
- **Response**: `200 OK`

```json
{
  "data": [
    {
      "id": "patrimony-id",
      "wallet_id": "wallet-id",
      "year": 2026,
      "month": 7,
      "type": "stocks",
      "amount": 150000,
      "created_at": "2026-07-25T12:00:00Z",
      "updated_at": "2026-07-25T12:00:00Z"
    }
  ]
}
```

### Create Patrimony
- **URL**: `POST /api/v1/wallets/:id/patrimonies`
- **Request Body**:

```json
{
  "year": 2026,
  "month": 7,
  "type": "stocks",
  "amount": 150000
}
```

- **Response**: `201 Created` or `409 Conflict` (duplicate) or `422 Unprocessable Entity` (validation)

### Update Patrimony
- **URL**: `PUT /api/v1/wallets/:id/patrimonies/:id`
- **Request Body**:

```json
{
  "year": 2026,
  "month": 7,
  "type": "stocks",
  "amount": 200000
}
```

- **Response**: `200 OK` or `404 Not Found` or `409 Conflict` (duplicate) or `422 Unprocessable Entity` (validation)

### Error Responses

All error responses follow the same structure:

```json
{
  "error": {
    "code": "ERROR_CODE",
    "message": "Human readable message",
    "details": {
      "field": ["error message"]
    }
  }
}
```

Common error codes:
- `VALIDATION_ERROR`: invalid input, missing fields or invalid values
- `PATRIMONY_ALREADY_EXISTS`: duplicate patrimony for the same wallet, year, month and type
- `PATRIMONY_NOT_FOUND`: patrimony id not found or does not belong to the wallet
- `ASSET_NOT_FOUND`: asset id not found
- `WALLET_NOT_FOUND`: wallet does not exist (returned by the wallet validation middleware)
- `INTERNAL_ERROR`: unexpected server error

---

## Assets

Individual investment launches. All monetary values are integer cents (e.g. `150000` means R$ 1.500,00). The `date` must be an RFC3339 timestamp and cannot be in the future. The `description` must be between 3 and 100 characters. The `amount` must be greater than zero.

Creating, updating or deleting an asset automatically recalculates the corresponding patrimony record (`SUM(amount)` of all assets for the same wallet/type/year/month) within the same transaction. When an update changes the asset's `date` and/or `type` across year/month or type boundaries, both the original and the new (wallet/type/year/month) patrimony records are recalculated.

Valid asset types (`type`): same as patrimony.

### Create Asset
- **URL**: `POST /api/v1/wallets/:id/assets`
- **Request Body**:

```json
{
  "type": "stocks",
  "date": "2026-07-15T12:00:00Z",
  "description": "PETR4 - Petrobras",
  "amount": 150000
}
```

- **Response**: `201 Created`

```json
{
  "data": {
    "id": "asset-id",
    "wallet_id": "wallet-id",
    "type": "stocks",
    "date": "2026-07-15T12:00:00Z",
    "description": "PETR4 - Petrobras",
    "amount": 150000,
    "created_at": "2026-07-25T12:00:00Z",
    "updated_at": "2026-07-25T12:00:00Z"
  }
}
```

Errors: `422 Unprocessable Entity` (validation), `500 Internal Server Error`.

### Update Asset
- **URL**: `PUT /api/v1/wallets/:id/assets/:id`
- **Request Body**: same as Create Asset
- **Response**: `200 OK`

```json
{
  "data": {
    "id": "asset-id",
    "wallet_id": "wallet-id",
    "type": "stocks",
    "date": "2026-07-20T12:00:00Z",
    "description": "PETR4 - Petrobras (split adjusted)",
    "amount": 200000,
    "created_at": "2026-07-25T12:00:00Z",
    "updated_at": "2026-07-25T12:00:00Z"
  }
}
```

Behavior: updating an asset triggers an automatic patrimony recalculation within the same transaction. If the asset's `date` and/or `type` change in a way that crosses month or type boundaries, both the original and the new (wallet/type/year/month) patrimony records are recalculated.

Errors: `404 Not Found` (asset not found), `422 Unprocessable Entity` (validation), `500 Internal Server Error`.

### List Assets
- **URL**: `GET /api/v1/wallets/:id/assets?type=&start_date=&end_date=`
- **Query Parameters** (all optional):
  - `type`: asset type filter (one of: `stocks`, `fiis`, `fixed_income`, `emergency_reserve`, `liquid_cash`)
  - `start_date`: inclusive lower bound for the asset date, format `YYYY-MM-DD` (e.g. `2026-07-01`)
  - `end_date`: inclusive upper bound for the asset date, format `YYYY-MM-DD` (e.g. `2026-07-31`)
- **Response**: `200 OK` with assets ordered by date DESC

```json
{
  "data": [
    {
      "id": "asset-id",
      "wallet_id": "wallet-id",
      "type": "stocks",
      "date": "2026-07-15T12:00:00Z",
      "description": "PETR4 - Petrobras",
      "amount": 150000,
      "created_at": "2026-07-25T12:00:00Z",
      "updated_at": "2026-07-25T12:00:00Z"
    }
  ]
}
```

Errors: `422 Unprocessable Entity` (invalid date format, or `start_date > end_date`), `500 Internal Server Error`.

### Delete Asset
- **URL**: `DELETE /api/v1/wallets/:id/assets/:id`
- **Response**: `204 No Content` on success. The corresponding patrimony record (wallet/type/year/month) is recalculated inside the same transaction. If the sum becomes zero, the patrimony record amount is set to zero (kept as record).
- **Errors**: `404 Not Found` (asset not found), `500 Internal Server Error`.

---

## Stocks

Read-only catalog of B3 stocks. The catalog is populated and updated exclusively via the `make seed-stock` CLI. It is not mutable through the REST API.

### List Stocks
- **URL**: `GET /api/v1/stocks`
- **Response**: `200 OK`

```json
{
  "data": [
    {
      "id": "stock-id",
      "ticker": "PETR4",
      "name": "Petrobras PN",
      "sector": "Petróleo, Gás e Biocombustíveis",
      "rank": 10,
      "website": "https://petrobras.com.br",
      "created_at": "2026-08-20T12:00:00Z",
      "updated_at": "2026-08-20T12:00:00Z"
    }
  ]
}
```

### Find Stock by Ticker
- **URL**: `GET /api/v1/stocks/:ticker`
- **Response**: `200 OK` or `404 Not Found`

```json
{
  "data": {
    "id": "stock-id",
    "ticker": "PETR4",
    "name": "Petrobras PN",
    "sector": "Petróleo, Gás e Biocombustíveis",
    "rank": 10,
    "website": "https://petrobras.com.br",
    "created_at": "2026-08-20T12:00:00Z",
    "updated_at": "2026-08-20T12:00:00Z"
  }
}
```

Error response:
```json
{
  "error": {
    "code": "STOCK_NOT_FOUND",
    "message": "Stock not found"
  }
}
```

### Seed Stocks
- **Command**: `make seed-stock`
- **Description**: Idempotently seeds ~29 B3 blue-chip stocks. Repeating the command does not duplicate rows or overwrite manual edits.
- **Force overwrite**: `make seed-stock ARGS="--force"`

---

## Positions

Stock positions inside a wallet. All monetary values are integer cents (e.g. `5000` means R$ 50,00). Percentages are returned as `float64`.

When a position is created, `ConsolidateByWallet` recalculates every position of the wallet in the same transaction, deriving:

- `invested` = `quantity` * `average_price`
- `balance` = `quantity` * `current_price`
- `variation_value` = `balance` - `invested`
- `variation_percent` = (`current_price` - `average_price`) / `average_price` * 100
- `portfolio_percent` = `invested` / SUM(`invested` of wallet) * 100

If a current price is missing from `stocks_current_prices`, it is treated as `0` (balance = 0, variation = -100%) and a warning is logged.

### List Positions
- **URL**: `GET /api/v1/wallets/:id/positions?ticker=&sort=`
- **Query Parameters** (all optional):
  - `ticker`: partial case-insensitive match against the stock ticker (e.g. `petr` matches `PETR4`)
  - `sort`: order by `ticker`, `rank`, `invested`, `variation_percent` or `portfolio_percent`. Prefix `-` for descending. Default is `balance DESC` (largest balances first). Values outside the whitelist fall back to the default without raising an error.
- **Response**: `200 OK`

```json
{
  "data": [
    {
      "id": "position-id",
      "wallet_id": "wallet-id",
      "stock_id": "stock-uuid",
      "quantity": 100,
      "average_price": 5000,
      "current_price": 7500,
      "invested": 500000,
      "balance": 750000,
      "variation_value": 250000,
      "variation_percent": 50,
      "portfolio_percent": 100,
      "created_at": "2026-08-26T12:00:00Z",
      "updated_at": "2026-08-26T12:00:00Z"
    }
  ]
}
```

An empty wallet returns `200 OK` with `{"data":[]}`.

- **Errors**:
  - `404 Not Found` (`WALLET_NOT_FOUND`): wallet `:id` does not exist
  - `500 Internal Server Error` (`INTERNAL_ERROR`): unexpected error

### Find Position
- **URL**: `GET /api/v1/wallets/:id/positions/:positionId`
- **Response**: `200 OK` with the same position payload as List, or `404 Not Found` (`POSITION_NOT_FOUND`) if the position does not exist or does not belong to the wallet in the path.

- **Errors**:
  - `404 Not Found` (`WALLET_NOT_FOUND`): wallet `:id` does not exist
  - `404 Not Found` (`POSITION_NOT_FOUND`): position `:positionId` not found or belongs to another wallet
  - `500 Internal Server Error` (`INTERNAL_ERROR`): unexpected error

### Create Position
- **URL**: `POST /api/v1/wallets/:id/positions`
- **Request Body**:

```json
{
  "stock_id": "stock-uuid",
  "quantity": 100,
  "average_price": 5000
}
```

- **Response**: `201 Created`

```json
{
  "data": {
    "id": "position-id",
    "wallet_id": "wallet-id",
    "stock_id": "stock-uuid",
    "quantity": 100,
    "average_price": 5000,
    "current_price": 7500,
    "invested": 500000,
    "balance": 750000,
    "variation_value": 250000,
    "variation_percent": 50,
    "portfolio_percent": 100,
    "created_at": "2026-08-26T12:00:00Z",
    "updated_at": "2026-08-26T12:00:00Z"
  }
}
```

- **Errors**:
  - `404 Not Found` (`WALLET_NOT_FOUND`): wallet `:id` does not exist
  - `404 Not Found` (`STOCK_NOT_FOUND`): `stock_id` does not exist in the catalog
  - `409 Conflict` (`POSITION_ALREADY_EXISTS`): a position for the same wallet and stock already exists
  - `422 Unprocessable Entity` (`VALIDATION_ERROR`): missing/invalid fields (`quantity` and `average_price` must be >= 1)
  - `500 Internal Server Error` (`INTERNAL_ERROR`): unexpected error

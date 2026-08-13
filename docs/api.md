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
- **URL**: `GET /api/v1/wallets/:walletId/patrimonies?type=&year=&month=`
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
- **URL**: `POST /api/v1/wallets/:walletId/patrimonies`
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
- **URL**: `PUT /api/v1/wallets/:walletId/patrimonies/:id`
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

Creating or deleting an asset automatically recalculates the corresponding patrimony record (`SUM(amount)` of all assets for the same wallet/type/year/month) within the same transaction.

Valid asset types (`type`): same as patrimony.

### Create Asset
- **URL**: `POST /api/v1/wallets/:walletId/assets`
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

### Delete Asset
- **URL**: `DELETE /api/v1/wallets/:walletId/assets/:id`
- **Response**: `204 No Content` on success. The corresponding patrimony record (wallet/type/year/month) is recalculated inside the same transaction. If the sum becomes zero, the patrimony record amount is set to zero (kept as record).
- **Errors**: `404 Not Found` (asset not found), `500 Internal Server Error`.

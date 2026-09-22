# Micro Investing

Micro Investing is a **REST API Backend** built with **Go + Gin Gonic** with an embedded **Vue 3 SPA** dashboard.

## 🚀 Overview

The project provides:

- A REST API for investment portfolio management
- A visual Dashboard SPA (Vue 3 + PrimeVue + Pinia + Chart.js) embedded into the Go binary and served at `/`

## 🏗 Architecture

This project follows a **Feature-Based Clean Architecture** pattern. Each feature is implemented as an independent module within the `internal/` directory, ensuring high cohesion and low coupling.

### Key Architectural Pillars:
- **Dependency Injection**: All dependencies are managed and injected via a centralized DI container.
- **Interface Segregation**: Strict use of interfaces for service and repository layers.
- **Repository Pattern**: Abstraction of data access logic.
- **Structured Logging**: Using Go's native `slog` for consistent observability.

## 📂 Directory Structure

```text
├── cmd/
│   └── api/                # Application entry point
├── docs/                   # Documentation and samples
├── internal/
│   ├── di/                 # Dependency Injection container
│   ├── shared/             # Shared utilities (config, logger, middleware)
│   └── webui/              # Embedded frontend (SPA) serving (embed.go, spa.go)
├── migrations/             # Database schema migrations
├── web/                    # Vue 3 SPA (dashboard frontend, independent project)
│   ├── public/             # Static assets copied verbatim on build
│   └── src/
│       ├── assets/         # Global styles
│       ├── components/dashboard/ # Presentational dashboard components
│       ├── lib/            # API client and utilities (formatCurrencyBRL)
│       ├── pages/          # Page components (Dashboard.vue)
│       ├── router/         # Vue Router setup
│       └── stores/         # Pinia stores (dashboard)
├── Makefile                # Development automation commands
└── .env.sample             # Environment variables template
```

## 🛠 Prerequisites

Ensure you have the following installed:
- **Go**: 1.23+
- **Node.js**: 20.19+ (or 22.12+) with npm — required for the frontend Dashboard
- **Make**: For running automation commands
- **golang-migrate**: For database migrations — install with SQLite support:
  ```bash
  go install -tags 'sqlite3' github.com/golang-migrate/migrate/v4/cmd/migrate@latest
  ```
- **A database engine** (pick one):
  - **PostgreSQL** (default driver) — running locally or via Docker
  - **MySQL** 8.0+ — running locally or via Docker
  - **SQLite** — no server required; uses a local file (great for dev/testing)
- **reflex** (optional): For hot-reload during development
- **golangci-lint** (optional): For linting

Install development tools:
```bash
make install-tools
```

## ⚙️ Installation & Setup (Development)

### 1. Clone the repository
```bash
git clone <repository-url>
cd micro-investing
```

### 2. Configure Environment Variables
Copy the sample environment file and adjust the values:
```bash
cp .env.sample .env
```

Default development values:
```env
SERVER_PORT=3003
DB_DRIVER=sqlite
DB_NAME=data/micro_investing.db
DB_HOST=
DB_PORT=
DB_USER=
DB_PASSWORD=
```

Supported drivers: `postgres` | `mysql` | `sqlite`. When `DB_DRIVER=sqlite`, only `DB_NAME` is required and it is treated as the database file path (e.g. `DB_NAME=data/app.db`).

### 3. Install Dependencies
```bash
make install-deps
```

### 4. Database Setup
Ensure your database is running and accessible with the credentials in `.env`, then run migrations:
```bash
make migrate
```

> **SQLite (file mode):** set `DB_DRIVER=sqlite` and `DB_NAME=data/app.db` (or any path). Create the directory first (`mkdir -p data`) if needed. No server is required.

### 5. Run the Application

**Hot-reload mode (development):**
```bash
make run-dev
```
Server starts at `http://localhost:3003` with automatic reload on file changes.

**Compiled binary:**
```bash
make run
```

## 💻 Frontend (Dashboard SPA)

The Dashboard is a Vue 3 SPA (Vite + TypeScript + PrimeVue + Pinia + Chart.js) that lives in `web/` (an independent frontend project) and is **embedded into the Go binary** via `go:embed` (`internal/webui/`). In production the API server serves the built SPA at `/` with client-side routing fallback; unknown `/api/*` routes still return the standard JSON 404 envelope.

### 1. Development mode
Run the Go API (`make run-dev`, port 3003) and the Vite dev server in parallel:
```bash
make web-dev
```
The dev server runs at `http://localhost:5173` and proxies `/api` requests to the Go API (dependencies are installed automatically by `make web-build`; for a standalone install run `cd web && npm install`).

### 2. Build for embedding
```bash
make web-build
```
Builds the SPA into `web/dist` and copies it to `internal/webui/dist`, which is embedded by `internal/webui` at Go compile time. Only a `.gitkeep` is versioned in that directory (build output is never committed); `make build` runs this step automatically before compiling the Go binary.

### 3. Tests
```bash
make web-test
```
Runs the Vitest suites: `formatCurrencyBRL` utility, the `useDashboardStore` Pinia store (fetch actions mapping API payloads into state) and the `Dashboard.vue` orchestration (KPI cards rendering store values as BRL-formatted currency).

## 🗄 Database Migrations

Migrations are managed using `golang-migrate`. The `DB_URL` used by the `migrate`/`rollback` targets is selected automatically based on `DB_DRIVER` in your `.env`:

| `DB_DRIVER` | `DB_URL` |
|-------------|----------|
| `sqlite`    | `sqlite3://<DB_NAME>` (file path) |
| `mysql` (default) | `mysql://<DB_USER>:<DB_PASSWORD>@tcp(<DB_HOST>:<DB_PORT>)/<DB_NAME>?multiStatements=true` |

> **Note:** `golang-migrate` must be installed with the `sqlite3` build tag for SQLite migrations to work:
> ```bash
> go install -tags 'sqlite3' github.com/golang-migrate/migrate/v4/cmd/migrate@latest
> ```

Commands:
- **Run migrations**: `make migrate`
- **Rollback last migration**: `make rollback`
- **Create new migration**: `make migrate-create name=<migration_name>`

## 🧪 Testing

- **Run all tests**: `make test`
- **Run frontend tests**: `make web-test`
- **Run tests with verbose output**: `make test-v`
- **Check test coverage**: `make test-cover`

## 📡 API Endpoints

### Health Check
- **URL**: `GET /health`
- **Description**: Verifies if the API and its dependencies are healthy.

### Wallets
- **URL**: `GET /api/v1/wallets`
- **URL**: `POST /api/v1/wallets`
- **URL**: `GET /api/v1/wallets/:id`
- **URL**: `PUT /api/v1/wallets/:id`
- **URL**: `DELETE /api/v1/wallets/:id`
- **Description**: Manage investment wallets.

### Patrimonies
- **URL**: `GET /api/v1/wallets/:id/patrimonies?type=&year=&month=`
- **URL**: `POST /api/v1/wallets/:id/patrimonies`
- **URL**: `PUT /api/v1/wallets/:id/patrimonies/:id`
- **Description**: Manage manual patrimony records per wallet. All monetary values are stored as integer cents. The `type` must be one of: `stocks`, `fiis`, `fixed_income`, `emergency_reserve`, `liquid_cash`.

### Assets
- **URL**: `GET /api/v1/wallets/:id/assets?type=&start_date=&end_date=`
- **URL**: `POST /api/v1/wallets/:id/assets`
- **URL**: `PUT /api/v1/wallets/:id/assets/:id`
- **URL**: `DELETE /api/v1/wallets/:id/assets/:id`
- **Description**: Manage individual asset launches (investments) per wallet. Each asset records a date, description, type and amount (integer cents). Creating, updating or deleting an asset automatically recalculates the corresponding patrimony record via `SUM(amount)` within the same transaction. Listing supports optional filters: `type`, `start_date` (inclusive, `YYYY-MM-DD`) and `end_date` (inclusive, `YYYY-MM-DD`). When both dates are provided, `start_date <= end_date` is enforced.

### Stocks
- **URL**: `GET /api/v1/stocks`
- **URL**: `GET /api/v1/stocks/:ticker`
- **Description**: Read-only catalog of B3 stocks. The catalog is populated via `make seed-stock` and is not mutable through the API.

### Positions
- **URL**: `GET /api/v1/wallets/:id/positions?ticker=&sort=`
- **URL**: `GET /api/v1/wallets/:id/positions/:positionId`
- **URL**: `POST /api/v1/wallets/:id/positions`
- **URL**: `PUT /api/v1/wallets/:id/positions/:positionId`
- **Description**: Create and query stock positions inside a wallet. The create request requires only `stock_id`, `quantity` and `average_price` (all monetary values in integer cents). The update request accepts only `quantity` and `average_price`; `stock_id` is immutable, so changing the ticker requires deleting the position and creating a new one. The response includes the derived fields `current_price`, `invested`, `balance`, `variation_value`, `variation_percent` and `portfolio_percent`. Creating or updating a position triggers `ConsolidateByWallet`, which recalculates all positions of the wallet in the same transaction. List returns an empty array (`[]`) when there are no positions, ordered by `balance DESC` by default. Optional query parameters: `ticker` (partial case-insensitive match) and `sort` (`ticker`, `rank`, `invested`, `variation_percent`, `portfolio_percent`; prefix `-` for descending). Invalid `sort` values fall back to the default ordering. Returns `404 Not Found` if the wallet or position/stock does not exist, `409 Conflict` if a position for the same stock already exists in the wallet, and `422 Unprocessable Entity` for invalid payloads.

### Dashboard
- **URL**: `GET /api/v1/wallets/:id/dashboard/summary`
- **URL**: `GET /api/v1/wallets/:id/dashboard/allocation`
- **URL**: `GET /api/v1/wallets/:id/dashboard/risk`
- **URL**: `GET /api/v1/wallets/:id/dashboard/evolution?year=&quarter=`
- **URL**: `GET /api/v1/wallets/:id/dashboard/dividends`
- **Description**: Returns aggregated dashboard metrics for the wallet. `current_patrimony` is the sum of all Patrimony amounts for the latest month with data. `stocks_invested` is the sum of Position.Invested for the wallet. `yearly_dividends` is always `0` until the dividends epic is implemented. The allocation endpoint returns the percentage distribution by asset type for the latest month with data; categories with no balance are omitted. The risk endpoint returns the percentage distribution of invested capital by stock rank (1-5), omitting ranks with no invested amount. The evolution endpoint returns monthly patrimony evolution with total and breakdown by `fixed_income`, `stocks` and `emergency_reserve`; months without data are filled via carry forward, and missing categories within a partial month are also carried forward per category. Without filters it returns the last 12 rolling months; with `year` it returns the full year; with `year` and `quarter` it returns the 3 months of that quarter. `year` must be greater than zero, `quarter` must be between 1 and 4 and requires `year`. The dividends endpoint returns the yearly dividends history as a static empty list until the dividends epic is implemented. All monetary values are integer cents.

### Seed
- **Command**: `make seed-stock`
- **Description**: Idempotently seeds the B3 blue-chip catalog into the `stocks` table. Repeating the command does not duplicate or overwrite manual edits.
- **Force overwrite**: `make seed-stock ARGS="--force"`

For detailed request/response schemas see `docs/api.md`.

| Command | Description |
|---------|-------------|
| `make run` | Build and run the API locally |
| `make run-dev` | Run with hot-reload (reflex) |
| `make build` | Build optimized binary (`bin/stats-central-api`) |
| `make test` | Run tests |
| `make test-v` | Run tests with verbose output |
| `make test-cover` | Generate coverage report |
| `make lint` | Run linter (golangci-lint) |
| `make install-deps` | Install Go dependencies |
| `make web-build` | Build the SPA and copy it to `internal/webui/dist` (embedded in the binary) |
| `make web-dev` | Run the Vite dev server with `/api` proxy |
| `make web-test` | Run frontend tests (Vitest) |
| `make install-tools` | Install dev tools (reflex, golangci-lint) |
| `make migrate` | Run database migrations |
| `make rollback` | Rollback last migration |
| `make migrate-create name=<name>` | Create new migration |
| `make seed-stock` | Seed the B3 stocks catalog (idempotent) |
| `make seed-stock ARGS="--force"` | Force-overwrite existing stocks catalog |
| `make clean` | Remove binaries and coverage files |

---

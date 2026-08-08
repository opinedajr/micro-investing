# Micro Investing

Micro Investing is a **REST API Backend** built with **Go + Gin Gonic**.

## 🚀 Overview

The project provides a comprehensive API for:

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
│   └── shared/             # Shared utilities (config, logger, middleware)
├── install/                # Production deployment scripts (install.sh, update.sh)
├── migrations/             # Database schema migrations
├── Makefile                # Development automation commands
└── .env.sample             # Environment variables template
```

## 🛠 Prerequisites

Ensure you have the following installed:
- **Go**: 1.23+
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
- **URL**: `GET /api/v1/wallets/:walletId/patrimonies?type=&year=&month=`
- **URL**: `POST /api/v1/wallets/:walletId/patrimonies`
- **URL**: `PUT /api/v1/wallets/:walletId/patrimonies/:id`
- **Description**: Manage manual patrimony records per wallet. All monetary values are stored as integer cents. The `type` must be one of: `stocks`, `fiis`, `fixed_income`, `emergency_reserve`, `liquid_cash`.

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
| `make install-tools` | Install dev tools (reflex, golangci-lint) |
| `make migrate` | Run database migrations |
| `make rollback` | Rollback last migration |
| `make migrate-create name=<name>` | Create new migration |
| `make clean` | Remove binaries and coverage files |

---

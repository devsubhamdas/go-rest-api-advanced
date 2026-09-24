# Go REST API

A production-oriented REST API built with **Go**, **GORM**, and **PostgreSQL**, with a focus on clean architecture, configuration management, database migrations, testing, middleware, and maintainable project structure.

## 🚀 Tech Stack

- **Go** `1.26.3`
- **PostgreSQL**
- **GORM** — ORM and database access
- **Goose** — SQL database migrations
- **pgx** — PostgreSQL driver
- **UUID** — Unique resource identifiers
- **cleanenv** — Configuration management
- **godotenv** — `.env` loading
- **Testify** — Testing assertions and utilities
- **Testcontainers** — Integration testing with real PostgreSQL containers
- **golang.org/x/crypto** — Cryptographic utilities
- **golang.org/x/time** — Rate limiting utilities
- **Bluemonday** — HTML sanitization

## ⚙️ Prerequisites

Make sure the following are installed:

- Go `1.26+`
- PostgreSQL
- Make
- Goose

Verify Go:

```bash
go version
```

Verify PostgreSQL:

```bash
psql --version
```

Verify Goose:

```bash
goose -version
```

## 📥 Installation

Clone the repository:

```bash
git clone https://github.com/devsubhamdas/go-rest-api-advanced.git
```

Move into the project:

```bash
cd go-rest-api-advanced
```

Download dependencies:

```bash
go mod download
```

Or:

```bash
go mod tidy
```

## 🔐 Environment Configuration

Create a `.env` file in the project root:

```env
PORT=8080

# Application environment
# Allowed values: development, production
APP_ENV=development

# PostgreSQL configuration
DB_HOST=localhost
DB_PORT=5432
DB_USER=postgres
DB_PASSWORD=your_password
DB_NAME=your_database

# Allowed values: require, disable
DB_SSLMODE=disable
```

### Environment Variables

| Variable      | Description             | Example       |
| ------------- | ----------------------- | ------------- |
| `PORT`        | API server port         | `8080`        |
| `APP_ENV`     | Application environment | `development` |
| `DB_HOST`     | PostgreSQL host         | `localhost`   |
| `DB_PORT`     | PostgreSQL port         | `5432`        |
| `DB_USER`     | PostgreSQL user         | `postgres`    |
| `DB_PASSWORD` | PostgreSQL password     | `password`    |
| `DB_NAME`     | PostgreSQL database     | `app_db`      |
| `DB_SSLMODE`  | PostgreSQL SSL mode     | `disable`     |

> **Never commit your `.env` file or database credentials to Git.**

# 🗄️ Database Migrations

This project uses **Goose** for database schema migrations.

Migration files are stored in:

```text
migrations/
```

The project intentionally keeps database migrations separate from GORM's `AutoMigrate`, allowing schema changes to be explicitly version-controlled.

## Create a Migration

```bash
make migrate-create name=create_users_table
```

This creates a new migration inside:

```text
migrations/
```

For example:

```text
migrations/
├── 20260924120000_create_users_table.sql
└── ...
```

A migration contains:

```sql
-- +goose Up

CREATE TABLE users (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(100) NOT NULL,
    email VARCHAR(255) NOT NULL UNIQUE,
    password TEXT NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW()
);

-- +goose Down

DROP TABLE users;
```

## Apply Migrations

Apply all pending migrations:

```bash
make migrate-up
```

Apply only the next pending migration:

```bash
make migrate-up-by-one
```

## Rollback Migrations

Rollback the most recent migration:

```bash
make migrate-down
```

Redo the most recent migration:

```bash
make migrate-redo
```

Reset the database by rolling back all migrations:

```bash
make migrate-reset
```

> `migrate-reset` is intended primarily for development/testing. Avoid using it against production databases unless you explicitly intend to destroy the migrated schema.

## Check Migration Status

```bash
make migrate-status
```

Example:

```text
Applied At                  Migration
========================================
20260924120000_create_users_table.sql
20260924130000_create_products_table.sql
```

# ▶️ Running the Application

## Development

Run the API directly:

```bash
make run
```

Equivalent to:

```bash
go run ./cmd/api
```

## Live Reload

If [Air](https://github.com/air-verse/air) is installed:

```bash
make air
```

This automatically rebuilds and restarts the application when source files change.

## Build

Build the application:

```bash
make build
```

The binary will be generated at:

```text
bin/api/main
```

## Start the Built Application

```bash
make start
```

This performs:

```text
build
  ↓
start binary
```

# 🧪 Testing

Run all tests:

```bash
make test
```

Equivalent to:

```bash
go test ./...
```

Run tests with verbose output:

```bash
go test ./... -v
```

Run tests for a specific feature:

```bash
go test ./internal/user/... -v
```

Run a specific test unit:

```bash
go test ./internal/user -v -run "^TestRepository"
```

```bash
go test ./internal/user -v -run "^TestService"
```

```bash
go test ./internal/user -v -run "^TestHandler"
```

The project uses:

- Go's standard `testing` package
- `testify`
- Testcontainers for integration testing

Testcontainers can be used to start a real PostgreSQL container during integration tests, avoiding dependence on a manually configured local database.

---

## 📁 Project Structure

```text
go-rest-api-advanced
├── Makefile
├── README.md
├── cmd
│   └── api
│       └── main.go
├── go.mod
├── go.sum
├── internal
│   ├── application
│   │   └── ...
│   ├── platform
│   │   ├── config
│   │   │   └── ...
│   │   ├── errorsx
│   │   │   └── ...
│   │   ├── hash
│   │   │   └── ...
│   │   ├── logger
│   │   │   └── ...
│   │   ├── middleware
│   │   │   ├── header
│   │   │   │   └── ...
│   │   │   └── ...
│   │   ├── response
│   │   │   ├── ...
│   │   └── storage
│   │       └── ...
│   └── user
│       ├── dto
│       │   └── ...
│       └── ...
└── migrations
    └── ...
```

### Architecture

The application follows a layered architecture:

```text
HTTP Request
     │
     ▼
 Middleware
     │
     ▼
 Handler
     │
     ▼
 Service
     │
     ▼
 Repository
     │
     ▼
 PostgreSQL
```

### Responsibilities

| Layer        | Responsibility                         |
| ------------ | -------------------------------------- |
| `handler`    | HTTP request/response handling         |
| `service`    | Business logic                         |
| `repository` | Database operations                    |
| `models`     | Database/domain models                 |
| `storage`    | Database connection and infrastructure |
| `config`     | Application configuration              |
| `cmd/api`    | Application entry point                |
| `migrations` | Database schema changes                |

---

# 🔍 Static Analysis

Run `go vet`:

```bash
make vet
```

Run the project's linter:

```bash
make lint
```

The lint command uses:

```bash
golangci-lint run
```

---

# 🧹 Formatting

Format the entire project:

```bash
make fmt
```

Equivalent to:

```bash
go fmt ./...
```

It is recommended to run formatting before committing changes.

---

# 🛠️ Makefile Commands

| Command                           | Description                                |
| --------------------------------- | ------------------------------------------ |
| `make run`                        | Run the API using `go run`                 |
| `make build`                      | Build the API binary                       |
| `make start`                      | Build and start the API                    |
| `make air`                        | Run with Air live reload                   |
| `make test`                       | Run all tests                              |
| `make vet`                        | Run `go vet`                               |
| `make fmt`                        | Format Go source files                     |
| `make lint`                       | Run golangci-lint                          |
| `make migrate-create name=<name>` | Create a migration                         |
| `make migrate-up`                 | Apply all pending migrations               |
| `make migrate-up-by-one`          | Apply the next migration                   |
| `make migrate-down`               | Roll back the latest migration             |
| `make migrate-redo`               | Roll back and reapply the latest migration |
| `make migrate-reset`              | Roll back all migrations                   |
| `make migrate-status`             | Show migration status                      |

---

# 🔄 Development Workflow

A typical development workflow is:

```text
1. Create/update model
        ↓
2. Create migration
        ↓
3. Write migration SQL
        ↓
4. Apply migration
        ↓
5. Implement repository
        ↓
6. Implement service
        ↓
7. Implement handler
        ↓
8. Add tests
        ↓
9. Run fmt
        ↓
10. Run vet/lint
        ↓
11. Run tests
```

For example:

```bash
make migrate-create name=create_users_table

make migrate-up

make fmt

make vet

make test

make lint
```

---

# 🏗️ Application Flow

A typical request flows through the application like this:

```text
                     HTTP Request
                          │
                          ▼
                   ┌─────────────┐
                   │ Middleware  │
                   └──────┬──────┘
                          │
                          ▼
                   ┌─────────────┐
                   │   Handler   │
                   └──────┬──────┘
                          │
                          ▼
                   ┌─────────────┐
                   │   Service   │
                   └──────┬──────┘
                          │
                          ▼
                   ┌─────────────┐
                   │ Repository  │
                   └──────┬──────┘
                          │
                          ▼
                   ┌─────────────┐
                   │ PostgreSQL  │
                   └─────────────┘
```

This separation keeps HTTP concerns, business logic, and persistence logic independent from each other.

---

# 🔒 Security Considerations

The API is designed with several security-related concerns in mind, including:

- Request validation
- Secure password handling using cryptographic packages
- CORS configuration
- Security-related HTTP headers
- Request recovery middleware
- Rate limiting
- Request IDs for tracing
- HTML sanitization where applicable
- Environment-based configuration
- Parameterized database queries through GORM

Production deployments should additionally use:

- HTTPS/TLS
- Secure and appropriately scoped secrets
- Restricted CORS origins
- Strong database credentials
- Database connection limits
- Proper logging and monitoring
- Regular dependency updates

---

# 🌍 Production Configuration

For production, configure:

```env
APP_ENV=production

PORT=8080

DB_HOST=your-production-db-host
DB_PORT=5432
DB_USER=your-production-db-user
DB_PASSWORD=your-production-db-password
DB_NAME=your-production-db
DB_SSLMODE=require
```

Production database migrations should be applied deliberately and should be part of the deployment process.

---

# 📌 API Health Endpoints

The API can expose health endpoints such as:

```http
GET /healthz
GET /livez
GET /readyz
```

Typical responsibilities:

| Endpoint   | Purpose                                                     |
| ---------- | ----------------------------------------------------------- |
| `/healthz` | General application health                                  |
| `/livez`   | Indicates whether the process is alive                      |
| `/readyz`  | Indicates whether the application is ready to serve traffic |

A readiness check can include dependencies such as PostgreSQL, while a liveness check should generally remain lightweight.

---

# 📦 Dependency Management

Download dependencies:

```bash
go mod download
```

Clean and synchronize dependencies:

```bash
go mod tidy
```

Verify dependencies:

```bash
go mod verify
```

---

# 📄 License

This project is intended for learning and development purposes.

Add an appropriate license file if this project is distributed publicly.

---

## 👤 Author

**Subham Das**

GitHub: [@devsubhamdas](https://github.com/devsubhamdas)

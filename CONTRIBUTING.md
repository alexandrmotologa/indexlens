# Contributing to IndexLens

Thank you for your interest in contributing to IndexLens.

## Development Prerequisites

- Go 1.23 or newer
- Node.js 20+ and npm 10+
- Optional: A running PostgreSQL instance (12+) with `pg_stat_statements` enabled

## Repository Setup

1. Clone the repository:
   ```bash
   git clone https://github.com/alexandrmotologa/indexlens.git
   cd indexlens
   ```

2. Install backend dependencies and run tests:
   ```bash
   go mod tidy
   go test -v ./...
   ```

3. Install frontend dependencies:
   ```bash
   cd web
   npm install
   npm run build
   cd ..
   ```

4. Run IndexLens locally:
   ```bash
   # Offline demo mode (no PostgreSQL required)
   go run ./cmd/indexlens serve --demo --port 9920

   # Or with a live database
   go run ./cmd/indexlens serve --dsn "postgres://user:pass@localhost:5432/dbname" --port 9920
   ```

## Code Conventions

- **Language**: All code, comments, commit messages, and documentation must be written in English.
- **Go Style**: Adhere to `gofmt` and standard idiomatic Go practices.
- **Safety First**: Never suggest blocking DDL. All synthesized indexes must use `CREATE INDEX CONCURRENTLY`.
- **Non-Invasive**: IndexLens must only read PostgreSQL system catalog views and statistics counters. Mutating application tables is strictly prohibited.
- **Documentation**: Keep documentation factual, clear, and free from empty superlatives or AI clichés.

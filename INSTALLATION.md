# Installation

## Prerequisites

- [Docker](https://docs.docker.com/get-docker/) (for Postgres and Redis)
- [Go 1.22+](https://go.dev/dl/)
- [golang-migrate](https://github.com/golang-migrate/migrate) CLI

```bash
brew install golang-migrate
```

## 1. Start Postgres & Redis

From the project root:

```bash
docker compose up -d
```

## 2. Run Migrations

```bash
cd backend
DATABASE_URL="postgres://fantasy:fantasy@localhost:5432/fantasy?sslmode=disable" make migrate-up
```

## 3. Configure Environment

Create a `.env` file inside `backend/`:

```
DATABASE_URL=postgres://fantasy:fantasy@localhost:5432/fantasy?sslmode=disable
REDIS_URL=redis://localhost:6379
```

## 4. Start the Server

```bash
cd backend
make run
```

The API is now available at `http://localhost:8080`.

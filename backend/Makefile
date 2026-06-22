.PHONY: run build test migrate-up migrate-down sync-odds-wcf sync-odds-fpl

run:
	go run ./cmd/server

build:
	go build -o bin/server ./cmd/server

test:
	go test ./...

migrate-up:
	migrate -path ./migrations -database "$(DATABASE_URL)" up

migrate-down:
	migrate -path ./migrations -database "$(DATABASE_URL)" down

sync-odds-wcf:
	WCF_ODDS_ENABLED=true go run ./cmd/sync-odds -game wcf

sync-odds-fpl:
	FPL_ODDS_ENABLED=true go run ./cmd/sync-odds -game fpl

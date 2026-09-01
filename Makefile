.PHONY: run build test test-unit test-integration fmt lint docker-up docker-down migrate-up migrate-down migrate-version migrate-force migrate-create
run:
	go run ./cmd/api
build:
	go build ./cmd/api ./cmd/migrate
test: test-unit
test-unit:
	go test -race ./internal/...
test-integration:
	go test -tags=integration ./...
fmt:
	gofmt -w cmd internal
lint:
	go vet ./...
docker-up:
	docker compose up -d postgres
docker-down:
	docker compose down
migrate-up:
	go run ./cmd/migrate up
migrate-down:
	go run ./cmd/migrate down
migrate-version:
	go run ./cmd/migrate version
migrate-force:
	go run ./cmd/migrate force $(version)
migrate-create:
	migrate create -ext sql -dir migrations -seq $(name)

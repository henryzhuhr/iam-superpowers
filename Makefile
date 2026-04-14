.PHONY: up down migrate-up migrate-down migrate-create run run-web test test-e2e test-e2e-web build build-web lint

up:
	docker-compose up -d

down:
	docker-compose down

migrate-up:
	migrate -path migrations -database "postgresql://$${DB_USER}:$${DB_PASSWORD}@$${DB_HOST}:$${DB_PORT}/$${DB_NAME}?sslmode=$${DB_SSLMODE}" up

migrate-down:
	migrate -path migrations -database "postgresql://$${DB_USER}:$${DB_PASSWORD}@$${DB_HOST}:$${DB_PORT}/$${DB_NAME}?sslmode=$${DB_SSLMODE}" down 1

migrate-create:
	@test -n "$(name)" || (echo "Usage: make migrate-create name=migration_name" && exit 1)
	migrate create -ext sql -dir migrations -seq $(name)

run:
	air -c .air.toml

test:
	go test -v -race ./...

build:
	go build -o bin/server ./cmd/server

lint:
	golangci-lint run

run-web:
	cd web && npm run dev

build-web:
	cd web && npm run build

test-e2e:
	cd tests/e2e && uv run pytest -v

test-e2e-web:
	cd web && npx playwright test

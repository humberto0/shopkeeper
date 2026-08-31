.PHONY: migrate-create migrate-up migrate-down migrate-version swag

DATABASE_URL ?= postgres://shopkeeper:shopkeeper@localhost:5432/shopkeeper?sslmode=disable
TEST_DATABASE_URL ?= postgres://shopkeeper:shopkeeper@localhost:5433/shopkeeper_test?sslmode=disable

swag:
	go run github.com/swaggo/swag/cmd/swag init -g cmd/api/main.go -o docs --parseInternal

migrate-create:
	@test -n "$(name)" || (echo "usage: make migrate-create name=create_users_table" && exit 1)
	migrate create -ext sql -dir migrations -seq $(name)

migrate-up:
	migrate -path migrations -database "$(DATABASE_URL)" up

migrate-down:
	migrate -path migrations -database "$(DATABASE_URL)" down 1

migrate-version:
	migrate -path migrations -database "$(DATABASE_URL)" version

test-db-up:
	docker compose up -d --wait postgres-test
	migrate -path migrations -database "$(TEST_DATABASE_URL)" up

test-integration: test-db-up
	go test ./test/integration/... -v
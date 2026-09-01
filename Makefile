.PHONY: migrate-create migrate-up migrate-down migrate-version swag test-db-up test-integration test-e2e test test-coverage test-coverage-html test-coverage-check govulncheck

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

test-e2e: test-db-up
	go test ./test/e2e/... -v

test:
	go test ./... -short

COVERAGE_THRESHOLD ?= 40

test-coverage:
	go test ./... -coverpkg=./... -coverprofile=coverage.out
	go tool cover -func=coverage.out | tail -1

test-coverage-html: test-coverage
	go tool cover -html=coverage.out

# test-coverage-check is what CI runs: it fails the build if total coverage
# drops below COVERAGE_THRESHOLD, instead of just reporting the number.
test-coverage-check: test-coverage
	@total=$$(go tool cover -func=coverage.out | tail -1 | awk '{print $$3}' | tr -d '%'); \
	echo "total coverage: $$total% (threshold: $(COVERAGE_THRESHOLD)%)"; \
	awk -v t="$$total" -v th="$(COVERAGE_THRESHOLD)" 'BEGIN { exit !(t+0 >= th+0) }' || \
		(echo "coverage $$total% is below the $(COVERAGE_THRESHOLD)% threshold" && exit 1)

govulncheck:
	go run golang.org/x/vuln/cmd/govulncheck ./...
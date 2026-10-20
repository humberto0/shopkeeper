.PHONY: fmt fmt-check lint typecheck build check migrate-create migrate-up migrate-down migrate-version swag db-up test-db-up test-integration test-e2e test test-coverage test-coverage-html test-coverage-check govulncheck

DATABASE_URL ?= postgres://shopkeeper:shopkeeper@localhost:5432/shopkeeper?sslmode=disable
TEST_DATABASE_URL ?= postgres://shopkeeper:shopkeeper@localhost:5433/shopkeeper_test?sslmode=disable

# Canonical "definition of done" targets: CI, lefthook and local sessions
# all run these, so the commands live in one place.

# goimports runs through golangci-lint so it uses the local-prefixes from
# .golangci.yml instead of a separately installed binary with its own defaults.
fmt:
	gofmt -l -w .
	golangci-lint fmt

fmt-check:
	@unformatted=$$(gofmt -l .); \
	if [ -n "$$unformatted" ]; then \
		echo "files not formatted with gofmt (run make fmt):"; \
		echo "$$unformatted"; \
		exit 1; \
	fi

lint:
	golangci-lint run

# The compiler already type-checks during build; go vet adds the static
# checks the compiler does not do (printf args, struct tags, copied locks).
typecheck:
	go vet ./...

# Same flags as the Dockerfile, so the local binary matches the image.
build:
	CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o bin/api ./cmd/api

# check is the single local pre-push command.
check: fmt-check lint typecheck test

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

db-up:
	docker compose up -d --wait postgres

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
	# -p 1: test/integration and test/e2e share the same Postgres database
	# and fixture emails, so running those packages concurrently (the
	# default) races on the same rows and produces spurious failures.
	go test ./... -p 1 -coverpkg=./... -coverprofile=coverage.out
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
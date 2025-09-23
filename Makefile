# Breakout Globe Makefile

.PHONY: help test test-unit test-integration test-integration-setup test-integration-teardown dev dev-down build clean

# Default target
help:
	@echo "Available targets:"
	@echo "  test                 - Run all tests (unit + integration)"
	@echo "  test-unit           - Run unit tests only"
	@echo "  test-integration    - Run integration tests with Docker setup"
	@echo "  test-integration-setup - Start test infrastructure"
	@echo "  test-integration-teardown - Stop test infrastructure"
	@echo "  dev                 - Start development environment"
	@echo "  dev-down           - Stop development environment"
	@echo "  build              - Build all services"
	@echo "  clean              - Clean up containers and volumes"

# Development environment
dev:
	docker compose up -d

dev-down:
	docker compose down

# Build services
build:
	docker compose build

# Unit tests (no external dependencies)
test-unit:
	@echo "Running unit tests..."
	cd backend && go test ./... --run

# Integration test infrastructure setup
test-integration-setup:
	@echo "Starting test infrastructure..."
	docker compose -f compose.test.yml up -d
	@echo "Waiting for services to be ready..."
	docker compose -f compose.test.yml exec postgres-test pg_isready -U postgres || sleep 5
	docker compose -f compose.test.yml exec redis-test redis-cli ping || sleep 5

# Integration test infrastructure teardown
test-integration-teardown:
	@echo "Stopping test infrastructure..."
	docker compose -f compose.test.yml down -v

# Integration tests (requires Docker infrastructure)
test-integration: test-integration-setup
	@echo "Running integration tests..."
	cd backend && TEST_INTEGRATION=1 \
		TEST_DB_HOST=localhost \
		TEST_DB_PORT=5433 \
		TEST_DB_USER=postgres \
		TEST_DB_PASSWORD=postgres \
		REDIS_TEST_HOST=localhost \
		REDIS_TEST_PORT=6380 \
		go test ./... -tags=integration --run
	$(MAKE) test-integration-teardown

# All tests
test: test-unit test-integration

# Clean up everything
clean:
	docker compose down -v
	docker compose -f compose.test.yml down -v
	docker system prune -f
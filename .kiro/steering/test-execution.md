# Test Execution Rule

## CRITICAL REQUIREMENT: Always use `--run` parameter for test execution

**NEVER run tests in watch mode during development assistance**

This project requires tests to be executed in run-once mode to avoid blocking the development workflow. You MUST always use:

✅ **CORRECT:** `npm test -- --run`
❌ **WRONG:** `npm test`

## CRITICAL REQUIREMENT: Database Container for Backend Tests

**ALWAYS ensure PostgreSQL container is running before executing backend tests**

Backend tests that involve database operations require the PostgreSQL container to be running. You MUST ensure:

✅ **CORRECT:** `docker compose up -d postgres` before running backend tests
❌ **WRONG:** Running backend tests without database container

## Examples

### Frontend Test Commands:
- `npm test -- --run` (run all frontend tests once)
- `npm test -- --run stores` (run specific frontend tests once)
- `npm test -- --run --reporter=basic` (run with minimal output)

### Backend Test Commands (requires database):
- `docker compose up -d postgres` (start database first)
- `go test ./...` (run all backend tests)
- `go test ./internal/models` (run specific package tests)
- `go test ./internal/testdata ./internal/integration` (run database-dependent tests)

### In Development Workflow:
Always use the `--run` parameter for frontend tests when:
- Verifying test status during TDD cycles
- Running tests after code changes
- Checking test coverage
- Validating implementations

Always ensure database container is running for backend tests when:
- Running integration tests
- Testing database models
- Running full backend test suite
- Validating database-dependent functionality

### In Documentation:
When providing setup or troubleshooting instructions:
- Always reference the `--run` parameter for frontend test execution
- Always reference Docker setup requirements for backend test execution

## Rationale

### Frontend Tests
The user's development environment requires tests to exit automatically without user interaction. Using watch mode blocks the workflow and requires manual intervention to quit the test runner.

### Backend Tests
Backend tests require database connectivity for integration testing, model validation, and repository operations. The PostgreSQL container provides the necessary database instance for these tests to run successfully.

## Docker Setup Requirements

### Database Container Management:
- **Start database**: `docker compose up -d postgres`
- **Check status**: `docker compose ps`
- **Stop database**: `docker compose down postgres`
- **View logs**: `docker compose logs postgres`

### Test Execution Dependencies:
- **Unit tests**: No database required (models, utilities)
- **Integration tests**: Database required (repositories, services)
- **Full test suite**: Database required for complete coverage

This rule applies to:
- All test execution commands
- All development workflow instructions
- All troubleshooting guidance
- All TDD cycle implementations
- All database-dependent backend testing
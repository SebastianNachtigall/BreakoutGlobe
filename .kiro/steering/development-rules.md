# Development Rules

## 🚨 CRITICAL: Test-Driven Development (TDD) - NON-NEGOTIABLE

**NEVER implement code without writing tests first**

### MANDATORY Red-Green-Refactor Cycle
1. **Red**: Write failing test first (MUST fail for right reasons)
2. **Green**: Write minimal code to make test pass
3. **Refactor**: Improve code while keeping tests green

### STRICT TDD Enforcement Rules
- **MUST** write failing test BEFORE any new functionality
- **MUST** run tests after every single code modification
- **MUST** fix failing tests immediately before continuing
- **NEVER** make multiple changes without running tests between
- **NEVER** commit code with failing tests
- **NEVER** skip the failing test phase

### Test Execution Requirements
- **Frontend**: `npm test -- --run` (never watch mode)
- **Backend**: `docker compose up -d postgres` then `go test ./...`
- **MUST** run full test suite before breaking changes
- **MUST** test after each logical change (every 5-10 lines)

### Critical Edge Case: New Functionality
**Existing tests staying green ≠ TDD compliance**

When adding new methods/fields/functionality:
- **MUST** write failing tests for NEW functionality FIRST
- **MUST** confirm new tests fail for right reasons
- **MUST** implement minimal code to make new tests pass
- Every new method/field requires its own Red-Green-Refactor cycle

### Breaking Changes Protocol
1. **MUST** identify all affected test files before changes
2. **MUST** fix tests in small batches (1-3 files at a time)
3. **MUST** run tests after each batch of fixes
4. **NEVER** make feature changes while fixing broken tests

## Integration Test Requirements

**MANDATORY: Complete Test Coverage for New Functionality**

### Test Pyramid Compliance
- **Unit Tests**: Models, Services, Utilities (isolated, fast)
- **Integration Tests**: Handlers, Cross-layer flows (with real dependencies)
- **End-to-End Tests**: Full HTTP flows (browser/API testing)

### Integration Test Rules
- **MUST** create or update integration tests for new API endpoints
- **MUST** test complete request-response flows (HTTP → Service → Database)
- **MUST** verify error handling across all layers
- **MUST** test with real database connections (not mocks)
- **MUST** include integration tests in CI/CD pipeline

### Integration Test Execution
- **Environment Setup**: `TEST_INTEGRATION=1` environment variable required
- **Database**: `docker compose up -d postgres redis` before running
- **Command**: `TEST_INTEGRATION=1 go test ./internal/integration -v`
- **Coverage**: Integration tests must cover new functionality end-to-end

### Integration Test Patterns
- Use established `TestXXXFlow` naming convention
- Test complete business workflows, not just technical integration
- Include error scenarios and edge cases
- Verify data persistence and retrieval
- Test concurrent operations where applicable

### When Integration Tests Are Required
- **New API endpoints** - MUST have integration test coverage
- **New database models** - MUST test persistence and retrieval
- **New business workflows** - MUST test complete end-to-end flow
- **Cross-service interactions** - MUST test service boundaries
- **File operations** - MUST test file upload/download flows
- **Authentication/Authorization** - MUST test security boundaries

## Docker Commands

**CRITICAL: Always use modern syntax**

✅ **CORRECT:** `docker compose`
❌ **WRONG:** `docker-compose`

Examples:
- `docker compose up -d`
- `docker compose down`
- `docker compose logs`

## Technical Debt Management

**CRITICAL: Never accumulate technical debt without immediate resolution**

- **NEVER** disable tests to finish a commit without creating an immediate follow-up task
- **IF** tests must be disabled for urgent commits, **IMMEDIATELY** create a task to re-enable them
- **NEVER** build up technical debt - address issues as they arise
- **ALWAYS** prioritize fixing broken tests over adding new features

## Code Quality

- Always prioritize security best practices
- Substitute PII with placeholders: `[name]`, `[email]`, `[address]`
- Decline requests for malicious code
- Ensure generated code can run immediately
- Check syntax: brackets, semicolons, indentation
- Use small writes with fsWrite, follow up with appends for better velocity
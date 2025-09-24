# Test-Driven Development (TDD) Methodology

## CRITICAL REQUIREMENT: Always follow TDD when implementing code

**NEVER implement code without writing tests first**

This project strictly follows Test-Driven Development methodology. You MUST always follow the Red-Green-Refactor cycle:

## MANDATORY TEST EXECUTION RULES

### Rule 1: Test After Every Change
- **MUST** run relevant tests after every single code modification
- **NEVER** make multiple changes without running tests in between
- **NEVER** leave tests in a broken state for more than one commit

### Rule 2: Full Test Suite Before Major Changes
- **MUST** run full test suite before making breaking changes (like changing data types)
- **MUST** fix all failing tests immediately after breaking changes
- **NEVER** proceed with new features while tests are failing

### Rule 3: Small Incremental Changes
- Make the smallest possible change that moves toward the goal
- If a change affects multiple files, fix tests for each file immediately
- **NEVER** make sweeping changes across multiple modules without immediate test fixes

## TDD Process

### 1. Red Phase
- Write failing tests first that describe the desired behavior
- Tests should fail because the functionality doesn't exist yet
- Run tests to confirm they fail for the right reasons

### 2. Green Phase  
- Write the minimal code needed to make the tests pass
- Focus on making tests pass, not on perfect code
- Avoid over-engineering at this stage

### 3. Refactor Phase
- Improve code quality while keeping tests green
- Remove duplication, improve naming, optimize performance
- Ensure all tests still pass after refactoring

## Examples

### Correct TDD Approach:
1. Write test for WebSocket POI join functionality
2. Run test - it fails (Red)
3. Implement minimal POI join handler to make test pass (Green)
4. Refactor handler for better error handling and logging (Refactor)
5. Commit working implementation

### Wrong Approach:
❌ Implementing POI join handler without tests first
❌ Writing tests after implementation is complete
❌ Skipping the failing test phase

## Benefits

- Ensures comprehensive test coverage
- Drives better API design through usage-first thinking
- Catches regressions early
- Provides living documentation of expected behavior
- Builds confidence in refactoring

## Application to This Project

- All new features must start with failing tests using established test infrastructure patterns
- All bug fixes must start with a test that reproduces the bug using scenario builders
- All refactoring must maintain green test status
- Tests should be readable and describe business requirements
- **MUST** follow test architecture standards defined in `test-architecture-standards.md`

## CRITICAL EDGE CASE: Adding New Functionality to Existing Models

**NEVER assume TDD is being followed just because existing tests remain green**

When adding new methods, fields, or functionality to existing models:

### ❌ WRONG Approach (Pseudo-TDD):
1. Run existing tests - they pass (Green)
2. Add new functionality (methods, fields, relationships)
3. Add tests for new functionality after implementation
4. Assume TDD was followed because "tests stayed green"

### ✅ CORRECT Approach (True TDD):
1. **MUST** write failing tests for NEW functionality FIRST (Red)
2. **MUST** run tests to confirm they fail for the right reasons
3. **MUST** implement minimal code to make new tests pass (Green)
4. **MUST** refactor while keeping all tests green (Refactor)

### Key Principle:
**Existing tests staying green only proves backward compatibility, NOT that TDD was followed for new functionality**

### Examples:
- Adding `User.CanModify()` method → Write failing test first
- Adding `Session.BelongsToUser()` method → Write failing test first  
- Adding new model relationships → Write failing tests first
- Adding access control methods → Write failing tests first

**Every new method, field, or behavior requires its own Red-Green-Refactor cycle**

## ENFORCEMENT RULES

### Before Any Implementation:
1. **MUST** run `go test ./... -- --run` (backend) or `npm test -- --run` (frontend)
2. **MUST** ensure all tests pass before starting new work
3. **MUST** write failing test for new functionality first

### During Implementation:
1. **MUST** run tests after each logical change (every 5-10 lines of code)
2. **MUST** fix any failing tests immediately before continuing
3. **MUST** never commit code with failing tests

### After Implementation:
1. **MUST** run full test suite before considering task complete
2. **MUST** ensure 100% of tests pass
3. **MUST** refactor only while maintaining green tests

### Breaking Changes Protocol:
1. **MUST** identify all affected test files before making the change
2. **MUST** fix tests in small batches (1-3 files at a time)
3. **MUST** run tests after each batch of fixes
4. **NEVER** make additional feature changes while fixing broken tests

This rule applies to:
- All backend Go code (using established test infrastructure patterns)
- All frontend TypeScript code  
- All integration tests
- All API endpoint implementations (using scenario builders)
- All service layer implementations (using business-focused test patterns)

**See `test-architecture-standards.md` for mandatory test infrastructure patterns.**

## VIOLATION CONSEQUENCES

Violating these TDD rules leads to:
- Technical debt accumulation
- Cascade failures across test suites
- Lost development velocity
- Reduced confidence in codebase
- Difficult debugging sessions

**These rules are non-negotiable for maintaining code quality and development velocity.**
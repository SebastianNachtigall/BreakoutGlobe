# Test-Driven Development (TDD) Methodology

## CRITICAL REQUIREMENT: Always follow TDD when implementing code

**NEVER implement code without writing tests first**

This project strictly follows Test-Driven Development methodology. You MUST always follow the Red-Green-Refactor cycle:

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

- All new features must start with failing tests
- All bug fixes must start with a test that reproduces the bug
- All refactoring must maintain green test status
- Tests should be readable and describe business requirements

This rule applies to:
- All backend Go code
- All frontend TypeScript code  
- All integration tests
- All API endpoint implementations
- All service layer implementations
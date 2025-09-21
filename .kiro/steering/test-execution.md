# Test Execution Rule

## CRITICAL REQUIREMENT: Always use `--run` parameter for test execution

**NEVER run tests in watch mode during development assistance**

This project requires tests to be executed in run-once mode to avoid blocking the development workflow. You MUST always use:

✅ **CORRECT:** `npm test -- --run`
❌ **WRONG:** `npm test`

## Examples

### Correct Test Commands:
- `npm test -- --run` (run all tests once)
- `npm test -- --run stores` (run specific tests once)
- `npm test -- --run --reporter=basic` (run with minimal output)

### In Development Workflow:
Always use the `--run` parameter when:
- Verifying test status during TDD cycles
- Running tests after code changes
- Checking test coverage
- Validating implementations

### In Documentation:
When providing setup or troubleshooting instructions, always reference the `--run` parameter for test execution.

## Rationale

The user's development environment requires tests to exit automatically without user interaction. Using watch mode blocks the workflow and requires manual intervention to quit the test runner.

This rule applies to:
- All test execution commands
- All development workflow instructions
- All troubleshooting guidance
- All TDD cycle implementations
# Test Architecture Standards

## CRITICAL REQUIREMENT: Always use the established test infrastructure patterns

**NEVER write tests using legacy patterns or brittle mock setups**

This project has established a resilient, maintainable test infrastructure that MUST be used for all new tests and when refactoring existing tests.

## MANDATORY TEST INFRASTRUCTURE PATTERNS

### Rule 1: Use Scenario Builders for All Test Setup
- **MUST** use scenario builders (`newPOIScenario(t)`, `newSessionScenario(t)`) for test setup
- **NEVER** use manual mock setup with `mock.AnythingOfType("*gin.Context")`
- **MUST** use fluent expectation APIs (`expectRateLimitSuccess().expectCreationSuccess()`)
- **NEVER** write repetitive mock expectations in individual tests

### Rule 2: Business-Focused Test Structure
- **MUST** focus tests on business logic and domain rules, not implementation details
- **MUST** use descriptive test names that express business intent
- **MUST** use `defer scenario.cleanup()` for automatic resource management
- **NEVER** write tests that focus on HTTP mechanics over business operations

### Rule 3: Consistent Patterns Across All Layers
- **MUST** follow the same patterns at Handler, Service, and Repository layers
- **MUST** use test data builders (`NewPOI().WithName().WithPosition()`) for model creation
- **MUST** use fluent assertion helpers (`AssertPOIResponse(t, poi).HasName().HasPosition()`)
- **NEVER** create ad-hoc test data or manual field-by-field assertions

## REQUIRED TEST INFRASTRUCTURE COMPONENTS

### Scenario Builders (MANDATORY)
All tests MUST use scenario builders that provide:

```go
// Handler Layer Example
func TestCreatePOI_Success(t *testing.T) {
    scenario := newPOIScenario(t)
    defer scenario.cleanup()
    
    scenario.expectRateLimitSuccess().
        expectCreationSuccess()
    
    poi := scenario.createPOI(CreatePOIRequest{...})
    
    AssertPOIResponse(t, poi).
        HasName("Coffee Shop").
        HasCreator("user-123")
}

// Service Layer Example  
func TestCreatePOI_BusinessRule(t *testing.T) {
    scenario := newPOIServiceScenario(t)
    defer scenario.cleanup()
    
    scenario.expectNoDuplicateLocation().
        expectCreateSuccess()
    
    poi, err := scenario.createPOI(...)
    
    assert.NoError(t, err)
    assert.Equal(t, "Meeting Room", poi.Name)
}
```

### Test Data Builders (MANDATORY)
All test data MUST be created using fluent builders:

```go
// ✅ CORRECT: Use fluent builders
poi := NewPOI().
    WithName("Coffee Shop").
    WithPosition(40.7128, -74.0060).
    WithCreator("user-123").
    WithMaxParticipants(15).
    Build()

// ❌ WRONG: Manual model creation
poi := &models.POI{
    ID: "poi-123",
    Name: "Coffee Shop",
    Position: models.LatLng{Lat: 40.7128, Lng: -74.0060},
    // ... repetitive field setting
}
```

### Fluent Assertions (MANDATORY)
All assertions MUST use fluent assertion helpers:

```go
// ✅ CORRECT: Fluent assertions
AssertPOIResponse(t, poi).
    HasName("Coffee Shop").
    HasPosition(40.7128, -74.0060).
    HasCreator("user-123")

// ❌ WRONG: Manual assertions
assert.Equal(t, "Coffee Shop", poi.Name)
assert.Equal(t, 40.7128, poi.Position.Lat)
assert.Equal(t, -74.0060, poi.Position.Lng)
assert.Equal(t, "user-123", poi.CreatedBy)
```

## FORBIDDEN PATTERNS

### ❌ NEVER Use These Legacy Patterns:

1. **Brittle Context Mocking**:
   ```go
   // ❌ FORBIDDEN
   mock.On("Method", mock.AnythingOfType("*gin.Context"), ...)
   ```

2. **Manual Mock Setup**:
   ```go
   // ❌ FORBIDDEN
   mockService := new(MockPOIService)
   mockRateLimiter := new(MockRateLimiter)
   // ... 15+ lines of setup
   ```

3. **Test Suites for Simple Tests**:
   ```go
   // ❌ FORBIDDEN for new tests
   type POIHandlerTestSuite struct {
       suite.Suite
       // ... complex setup
   }
   ```

4. **Manual HTTP Request Construction**:
   ```go
   // ❌ FORBIDDEN
   body, _ := json.Marshal(request)
   req := httptest.NewRequest(http.MethodPost, "/api/pois", bytes.NewBuffer(body))
   // ... manual HTTP mechanics
   ```

## LAYER-SPECIFIC REQUIREMENTS

### Handler Layer Tests
- **MUST** use `newHandlerScenario(t)` pattern
- **MUST** focus on business operations, not HTTP status codes
- **MUST** use rate limiting expectation helpers
- **MUST** test error scenarios with business-focused assertions

### Service Layer Tests  
- **MUST** use `newServiceScenario(t)` pattern
- **MUST** focus on domain rules and business logic
- **MUST** test integration patterns (repository + cache + pubsub)
- **MUST** use business rule expectation helpers

### Repository Layer Tests
- **MUST** use `newRepositoryScenario(t)` pattern (when implemented)
- **MUST** focus on data persistence patterns
- **MUST** use database fixture builders
- **MUST** test transaction boundaries and error conditions

## MIGRATION REQUIREMENTS

### When Refactoring Existing Tests:
1. **MUST** migrate to new patterns when touching existing test files
2. **MUST** create migrated versions to demonstrate improvements
3. **MUST** maintain test coverage during migration
4. **NEVER** mix old and new patterns in the same file

### When Adding New Tests:
1. **MUST** use only the new test infrastructure patterns
2. **MUST** follow established naming conventions
3. **MUST** add to existing scenario builders when possible
4. **MUST** create new builders only when necessary

## ENFORCEMENT AND VALIDATION

### Code Review Requirements:
- **MUST** reject PRs that use legacy test patterns
- **MUST** require scenario builders for all new tests
- **MUST** ensure business-focused test names and assertions
- **MUST** verify proper cleanup patterns (`defer scenario.cleanup()`)

### Automated Checks:
- Tests using `mock.AnythingOfType("*gin.Context")` should be flagged
- Tests without scenario builders should be flagged
- Tests with manual HTTP construction should be flagged
- Tests without fluent assertions should be flagged

## BENEFITS ENFORCEMENT

These patterns are mandatory because they provide:

✅ **Resilience**: Interface changes affect <5 files instead of 100+  
✅ **Readability**: Business intent is immediately clear  
✅ **Maintainability**: New tests take minutes, not hours  
✅ **Consistency**: Same patterns work across all layers  
✅ **Developer Experience**: Testing becomes enjoyable and productive  

## VIOLATION CONSEQUENCES

Violating these test architecture standards leads to:
- **Technical Debt**: Brittle tests that break with interface changes
- **Maintenance Burden**: 10x more files to update for simple changes  
- **Developer Frustration**: Difficult test writing and debugging
- **Reduced Velocity**: Slow test development and maintenance
- **Inconsistent Patterns**: Confusion and knowledge silos

## EXAMPLES AND REFERENCES

### Reference Implementations:
- `backend/internal/handlers/poi_handler_migrated_test.go` - Handler patterns
- `backend/internal/handlers/session_handler_migrated_test.go` - Session patterns  
- `backend/internal/services/poi_service_migrated_test.go` - Service patterns
- `backend/internal/testdata/MIGRATION_DEMO.md` - Before/after comparison

### Infrastructure Components:
- `backend/internal/testdata/builders.go` - Test data builders
- `backend/internal/testdata/scenarios.go` - Scenario builders
- `backend/internal/testdata/assertions.go` - Fluent assertions
- `backend/internal/testdata/mocks.go` - Mock abstractions

**These test architecture standards are non-negotiable for maintaining code quality, developer productivity, and sustainable development velocity.**

## INTEGRATION WITH TDD

These patterns MUST be used in conjunction with the TDD methodology defined in `tdd-methodology.md`:

1. **Red Phase**: Write failing tests using scenario builders
2. **Green Phase**: Implement minimal code to pass business-focused tests  
3. **Refactor Phase**: Improve code while maintaining green scenario-based tests

The test infrastructure patterns enhance TDD by making the Red-Green-Refactor cycle faster, more reliable, and more focused on business value.
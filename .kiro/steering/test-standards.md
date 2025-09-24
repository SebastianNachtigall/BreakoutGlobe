# Test Standards

## CRITICAL: Always use established test infrastructure patterns

**NEVER write tests using legacy patterns or brittle mock setups**

## Mandatory Patterns

### 1. Scenario Builders (REQUIRED)
```go
// ✅ CORRECT: Use scenario builders
func TestCreatePOI_Success(t *testing.T) {
    scenario := newPOIScenario(t)
    defer scenario.cleanup()
    
    scenario.expectRateLimitSuccess().expectCreationSuccess()
    poi := scenario.createPOI(CreatePOIRequest{...})
    
    AssertPOIResponse(t, poi).HasName("Coffee Shop").HasCreator("user-123")
}
```

### 2. Test Data Builders (REQUIRED)
```go
// ✅ CORRECT: Use fluent builders
poi := NewPOI().WithName("Coffee Shop").WithPosition(40.7128, -74.0060).Build()

// ❌ WRONG: Manual model creation
poi := &models.POI{ID: "poi-123", Name: "Coffee Shop", ...}
```

### 3. Fluent Assertions (REQUIRED)
```go
// ✅ CORRECT: Fluent assertions
AssertPOIResponse(t, poi).HasName("Coffee Shop").HasPosition(40.7128, -74.0060)

// ❌ WRONG: Manual assertions
assert.Equal(t, "Coffee Shop", poi.Name)
assert.Equal(t, 40.7128, poi.Position.Lat)
```

## Forbidden Patterns

❌ **NEVER use these legacy patterns:**
- `mock.AnythingOfType("*gin.Context")` - Use scenario builders instead
- Manual HTTP request construction - Use scenario execution methods
- `suite.Suite` for simple tests - Use scenario builders
- Manual model creation - Use fluent builders
- Field-by-field assertions - Use fluent assertion helpers
- Missing `defer scenario.cleanup()` - Always clean up resources

## Layer-Specific Requirements

### Handler Tests
- Use `newHandlerScenario(t)` pattern
- Focus on business operations, not HTTP mechanics
- Use rate limiting expectation helpers

### Service Tests  
- Use `newServiceScenario(t)` pattern
- Focus on domain rules and business logic
- Test integration patterns (repository + cache + pubsub)

### Repository Tests
- Use `newRepositoryScenario(t)` pattern
- Focus on data persistence patterns
- Use database fixture builders

## TDD Integration with Test Standards

**These patterns MUST be used with strict TDD methodology:**

1. **Red Phase**: Write failing tests using scenario builders
2. **Green Phase**: Implement minimal code to pass business-focused tests
3. **Refactor Phase**: Improve code while maintaining green scenario-based tests

## Reference Files
- **Examples**: `backend/internal/handlers/*_migrated_test.go`
- **Infrastructure**: `backend/internal/testdata/*.go`
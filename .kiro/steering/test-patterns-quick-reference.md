# Test Patterns Quick Reference

## 🚀 **Quick Start: Writing Tests the Right Way**

### Handler Tests Template
```go
func TestFeature_Scenario(t *testing.T) {
    // 1. Setup scenario
    scenario := newFeatureScenario(t)
    defer scenario.cleanup()
    
    // 2. Set expectations
    scenario.expectRateLimitSuccess().
        expectBusinessRuleSuccess()
    
    // 3. Execute operation
    result := scenario.executeOperation(request)
    
    // 4. Assert business outcomes
    AssertResponse(t, result).
        HasExpectedField("value").
        MeetsBusinessRule()
}
```

### Service Tests Template
```go
func TestBusinessRule_Scenario(t *testing.T) {
    // 1. Setup scenario
    scenario := newBusinessScenario(t)
    defer scenario.cleanup()
    
    // 2. Set business expectations
    scenario.expectDomainRuleEnforced().
        expectIntegrationSuccess()
    
    // 3. Execute business operation
    result, err := scenario.executeBusinessOperation(params)
    
    // 4. Assert domain outcomes
    assert.NoError(t, err)
    assert.Equal(t, expectedBusinessValue, result.BusinessField)
}
```

## ✅ **DO: Use These Patterns**

### Test Data Creation
```go
// ✅ Use fluent builders
poi := NewPOI().
    WithName("Coffee Shop").
    WithPosition(40.7128, -74.0060).
    WithCreator("user-123").
    Build()

session := NewSession().
    WithUser("user-456").
    WithMap("map-789").
    WithPosition(41.0, -75.0).
    Build()
```

### Mock Expectations
```go
// ✅ Use fluent expectations
scenario.expectRateLimitSuccess().
    expectCreationSuccess().
    expectNotificationSent()

// ✅ Use business-focused expectations
scenario.expectNoDuplicateLocation().
    expectUserCanJoin().
    expectCapacityAvailable()
```

### Assertions
```go
// ✅ Use fluent assertions
AssertPOIResponse(t, poi).
    HasName("Coffee Shop").
    HasCreator("user-123").
    IsActive()

AssertErrorResponse(t, err).
    HasCode("RATE_LIMIT_EXCEEDED").
    HasRetryAfter("3600")
```

## ❌ **DON'T: Avoid These Legacy Patterns**

### Brittle Mock Setup
```go
// ❌ NEVER use brittle context mocking
mockService.On("Method", 
    mock.AnythingOfType("*gin.Context"), // BRITTLE!
    param1, param2).Return(result, nil)

// ❌ NEVER use manual mock setup
mockPOIService := new(MockPOIService)
mockRateLimiter := new(MockRateLimiter)
handler := NewPOIHandler(mockPOIService, mockRateLimiter)
// ... 15+ lines of setup
```

### Manual Test Data
```go
// ❌ NEVER create models manually
poi := &models.POI{
    ID: "poi-123",
    Name: "Coffee Shop",
    Position: models.LatLng{Lat: 40.7128, Lng: -74.0060},
    CreatedBy: "user-123",
    MaxParticipants: 15,
    CreatedAt: time.Now(),
    // ... repetitive field setting
}
```

### Manual HTTP Construction
```go
// ❌ NEVER construct HTTP requests manually
body, _ := json.Marshal(request)
req := httptest.NewRequest(http.MethodPost, "/api/pois", bytes.NewBuffer(body))
req.Header.Set("Content-Type", "application/json")
recorder := httptest.NewRecorder()
router.ServeHTTP(recorder, req)
// ... manual response parsing
```

## 🎯 **Layer-Specific Patterns**

### Handler Layer
- **Focus**: Business operations, not HTTP mechanics
- **Pattern**: `newHandlerScenario(t)` → expectations → execute → assert
- **Example**: Rate limiting, request validation, response formatting

### Service Layer  
- **Focus**: Domain rules and business logic
- **Pattern**: `newServiceScenario(t)` → business expectations → execute → assert
- **Example**: Duplicate prevention, participation rules, integration coordination

### Repository Layer
- **Focus**: Data persistence patterns
- **Pattern**: `newRepositoryScenario(t)` → data expectations → execute → assert  
- **Example**: Transaction boundaries, constraint enforcement, query optimization

## 🔧 **Common Scenarios**

### Rate Limiting Tests
```go
scenario.expectRateLimit(ActionCreatePOI, 5, "1h")
err := scenario.executeExpectingError(request)
AssertErrorResponse(t, err).HasCode("RATE_LIMIT_EXCEEDED")
```

### Business Rule Violations
```go
scenario.expectDuplicateLocation(existingPOI)
poi, err := scenario.createPOI(request)
assert.Error(t, err)
assert.Contains(t, err.Error(), "already exists at this location")
```

### Success Scenarios
```go
scenario.expectRateLimitSuccess().expectCreationSuccess()
poi := scenario.createPOI(request)
AssertPOIResponse(t, poi).HasName("Coffee Shop").IsActive()
```

## 📚 **Reference Files**

- **Handler Examples**: `backend/internal/handlers/*_migrated_test.go`
- **Service Examples**: `backend/internal/services/*_migrated_test.go`
- **Infrastructure**: `backend/internal/testdata/*.go`
- **Migration Guide**: `backend/internal/testdata/MIGRATION_DEMO.md`

## 🚨 **Red Flags in Code Review**

Watch for these patterns that indicate legacy test code:

- `mock.AnythingOfType("*gin.Context")` - Use scenario builders instead
- Manual HTTP request construction - Use scenario execution methods
- `suite.Suite` for simple tests - Use scenario builders
- Manual model creation - Use fluent builders
- Field-by-field assertions - Use fluent assertion helpers
- Missing `defer scenario.cleanup()` - Always clean up resources

**When you see these patterns, require migration to the established test infrastructure!**
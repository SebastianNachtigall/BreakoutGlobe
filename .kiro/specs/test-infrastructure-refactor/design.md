# Test Infrastructure Refactoring Design

## Architecture Overview

This design document outlines the technical approach for refactoring our test infrastructure to be more resilient, maintainable, and aligned with TDD best practices.

## Current State Analysis

### Problems with Current Architecture
```go
// Current: Brittle and verbose
func TestCreatePOI(t *testing.T) {
    mockService := &MockPOIService{}
    mockRateLimiter := &MockRateLimiter{}
    
    // 15+ lines of mock setup with context type issues
    mockRateLimiter.On("CheckRateLimit", mock.AnythingOfType("context.backgroundCtx"), userID, action).Return(nil)
    mockService.On("CreatePOI", mock.AnythingOfType("context.backgroundCtx"), mapID, name, desc, pos, userID, max).Return(poi, nil)
    // ... more mock setup
    
    // Test execution
    handler := NewPOIHandler(mockService, mockRateLimiter)
    // ... HTTP request setup and assertions
}
```

### Root Causes
1. **No Abstraction**: Direct mock usage in every test
2. **Context Confusion**: Manual context type handling
3. **Repetitive Setup**: Same patterns copied across files
4. **Tight Coupling**: Tests know too much about implementation

## Target Architecture

### 1. Test Builder Pattern

```go
// New: Expressive and resilient
func TestCreatePOI(t *testing.T) {
    scenario := NewPOITestScenario().
        WithValidUser().
        WithValidMap().
        ExpectRateLimitSuccess().
        ExpectCreationSuccess()
    
    response := scenario.CreatePOI(t, CreatePOIRequest{
        Name: "Coffee Shop",
        Position: LatLng{40.7128, -74.0060},
    })
    
    scenario.AssertPOICreated(response, "Coffee Shop")
}
```

### 2. Layered Test Architecture

```
┌─────────────────────────────────────┐
│           Test Layer                │
├─────────────────────────────────────┤
│         Builder Layer               │  ← New abstraction layer
├─────────────────────────────────────┤
│         Mock Layer                  │  ← Simplified and hidden
├─────────────────────────────────────┤
│       Service Layer                 │
└─────────────────────────────────────┘
```

## Component Design

### 1. Test Scenario Builders

#### POI Test Scenario Builder
```go
type POITestScenario struct {
    mockSetup    *MockSetup
    userID       uuid.UUID
    mapID        uuid.UUID
    expectations []Expectation
}

func NewPOITestScenario() *POITestScenario {
    return &POITestScenario{
        mockSetup: NewMockSetup(),
        userID:    testdata.GenerateUUID(),
        mapID:     testdata.GenerateUUID(),
    }
}

func (s *POITestScenario) WithValidUser() *POITestScenario {
    s.userID = testdata.GenerateUUID()
    return s
}

func (s *POITestScenario) WithMap(mapID string) *POITestScenario {
    s.mapID = uuid.MustParse(mapID)
    return s
}

func (s *POITestScenario) ExpectRateLimitSuccess() *POITestScenario {
    s.expectations = append(s.expectations, RateLimitExpectation{
        UserID: s.userID,
        Action: services.ActionCreatePOI,
        Result: Success,
    })
    return s
}

func (s *POITestScenario) ExpectCreationSuccess() *POITestScenario {
    s.expectations = append(s.expectations, POICreationExpectation{
        MapID:  s.mapID,
        UserID: s.userID,
        Result: Success,
    })
    return s
}

func (s *POITestScenario) CreatePOI(t *testing.T, req CreatePOIRequest) *CreatePOIResponse {
    // Setup mocks based on expectations
    s.setupMocks()
    
    // Create handler with mocks
    handler := handlers.NewPOIHandler(s.mockSetup.POIService, s.mockSetup.RateLimiter)
    
    // Execute request
    return s.executeRequest(t, handler, req)
}
```

#### Session Test Scenario Builder
```go
type SessionTestScenario struct {
    mockSetup *MockSetup
    userID    uuid.UUID
    mapID     uuid.UUID
    sessionID uuid.UUID
}

func NewSessionTestScenario() *SessionTestScenario {
    return &SessionTestScenario{
        mockSetup: NewMockSetup(),
        userID:    testdata.GenerateUUID(),
        mapID:     testdata.GenerateUUID(),
        sessionID: testdata.GenerateUUID(),
    }
}

func (s *SessionTestScenario) WithExistingSession() *SessionTestScenario {
    s.mockSetup.SessionService.ExpectGetSession(s.sessionID).
        Returns(testdata.NewSession().
            WithID(s.sessionID).
            WithUser(s.userID).
            WithMap(s.mapID).
            Build())
    return s
}
```

### 2. Mock Setup Abstraction

```go
type MockSetup struct {
    POIService     *MockPOIServiceBuilder
    SessionService *MockSessionServiceBuilder
    RateLimiter    *MockRateLimiterBuilder
}

func NewMockSetup() *MockSetup {
    return &MockSetup{
        POIService:     NewMockPOIServiceBuilder(),
        SessionService: NewMockSessionServiceBuilder(),
        RateLimiter:    NewMockRateLimiterBuilder(),
    }
}

type MockPOIServiceBuilder struct {
    mock *services.MockPOIService
}

func (m *MockPOIServiceBuilder) ExpectCreate(mapID uuid.UUID, userID uuid.UUID) *POICreationExpectation {
    return &POICreationExpectation{
        mock:   m.mock,
        mapID:  mapID,
        userID: userID,
    }
}

type POICreationExpectation struct {
    mock   *services.MockPOIService
    mapID  uuid.UUID
    userID uuid.UUID
}

func (e *POICreationExpectation) Returns(poi *models.POI) *POICreationExpectation {
    // Handle context automatically - no more manual context.backgroundCtx!
    e.mock.On("CreatePOI", 
        mock.AnythingOfType("context.backgroundCtx"), 
        e.mapID.String(), 
        mock.AnythingOfType("string"), // name
        mock.AnythingOfType("string"), // description
        mock.AnythingOfType("models.LatLng"), 
        e.userID.String(), 
        mock.AnythingOfType("int")).Return(poi, nil)
    return e
}

func (e *POICreationExpectation) ReturnsError(err error) *POICreationExpectation {
    e.mock.On("CreatePOI", 
        mock.AnythingOfType("context.backgroundCtx"), 
        e.mapID.String(), 
        mock.AnythingOfType("string"),
        mock.AnythingOfType("string"),
        mock.AnythingOfType("models.LatLng"), 
        e.userID.String(), 
        mock.AnythingOfType("int")).Return(nil, err)
    return e
}
```

### 3. Test Data Factories

```go
package testdata

type UserBuilder struct {
    user *models.User
}

func NewUser() *UserBuilder {
    return &UserBuilder{
        user: &models.User{
            ID:          GenerateUUID(),
            DisplayName: "Test User",
            Email:       "test@example.com",
            AccountType: models.AccountTypeGuest,
            Role:        models.RoleUser,
            CreatedAt:   time.Now(),
            UpdatedAt:   time.Now(),
        },
    }
}

func (b *UserBuilder) WithID(id uuid.UUID) *UserBuilder {
    b.user.ID = id
    return b
}

func (b *UserBuilder) WithEmail(email string) *UserBuilder {
    b.user.Email = email
    return b
}

func (b *UserBuilder) WithRole(role models.UserRole) *UserBuilder {
    b.user.Role = role
    return b
}

func (b *UserBuilder) Build() *models.User {
    return b.user
}

type POIBuilder struct {
    poi *models.POI
}

func NewPOI() *POIBuilder {
    return &POIBuilder{
        poi: &models.POI{
            ID:              GenerateUUID(),
            MapID:           GenerateUUID(),
            Name:            "Test POI",
            Description:     "Test Description",
            Position:        LatLng{Lat: 40.7128, Lng: -74.0060},
            CreatedBy:       GenerateUUID(),
            MaxParticipants: 10,
            CreatedAt:       time.Now(),
            UpdatedAt:       time.Now(),
        },
    }
}

func (b *POIBuilder) WithCreator(userID uuid.UUID) *POIBuilder {
    b.poi.CreatedBy = userID
    return b
}

func (b *POIBuilder) WithMap(mapID uuid.UUID) *POIBuilder {
    b.poi.MapID = mapID
    return b
}

func (b *POIBuilder) Build() *models.POI {
    return b.poi
}

// UUID utilities
func GenerateUUID() uuid.UUID {
    return uuid.New()
}

func ParseUUID(s string) uuid.UUID {
    return uuid.MustParse(s)
}
```

### 4. Integration Test Support

```go
package testdb

func Setup(t *testing.T) *gorm.DB {
    // Create isolated test database
    dbName := fmt.Sprintf("test_%s_%d", t.Name(), time.Now().UnixNano())
    
    db, err := gorm.Open(postgres.Open(testDSN(dbName)), &gorm.Config{})
    require.NoError(t, err)
    
    // Run migrations
    err = db.AutoMigrate(&models.User{}, &models.Map{}, &models.POI{}, &models.Session{})
    require.NoError(t, err)
    
    // Cleanup on test completion
    t.Cleanup(func() {
        dropTestDB(dbName)
    })
    
    return db
}

func SeedFixtures(db *gorm.DB, fixtures ...interface{}) {
    for _, fixture := range fixtures {
        db.Create(fixture)
    }
}

// Integration test example
func TestPOICreation_Integration(t *testing.T) {
    db := testdb.Setup(t)
    
    // Seed test data
    user := testdata.NewUser().Build()
    testMap := testdata.NewMap().Build()
    testdb.SeedFixtures(db, user, testMap)
    
    // Use real services
    poiService := services.NewPOIService(db, realRedisClient)
    rateLimiter := services.NewRateLimiter(realRedisClient)
    handler := handlers.NewPOIHandler(poiService, rateLimiter)
    
    // Test real behavior
    response := makeCreatePOIRequest(handler, CreatePOIRequest{
        MapID:    testMap.ID.String(),
        Name:     "Real Coffee Shop",
        CreatedBy: user.ID.String(),
    })
    
    // Verify in database
    var createdPOI models.POI
    err := db.First(&createdPOI, "name = ?", "Real Coffee Shop").Error
    require.NoError(t, err)
    assert.Equal(t, "Real Coffee Shop", createdPOI.Name)
}
```

### 5. Assertion Helpers

```go
package testassert

func POIResponse(t *testing.T, response *CreatePOIResponse, expected *models.POI) {
    assert.Equal(t, expected.ID.String(), response.ID)
    assert.Equal(t, expected.Name, response.Name)
    assert.Equal(t, expected.Description, response.Description)
    assert.Equal(t, expected.Position.Lat, response.Position.Lat)
    assert.Equal(t, expected.Position.Lng, response.Position.Lng)
    // Ignore timestamps and other irrelevant fields
}

func HTTPStatus(t *testing.T, response *httptest.ResponseRecorder, expectedStatus int) {
    if response.Code != expectedStatus {
        t.Errorf("Expected status %d, got %d. Response body: %s", 
            expectedStatus, response.Code, response.Body.String())
    }
}

func ErrorResponse(t *testing.T, response *httptest.ResponseRecorder, expectedCode string) {
    var errorResp handlers.ErrorResponse
    err := json.Unmarshal(response.Body.Bytes(), &errorResp)
    require.NoError(t, err)
    assert.Equal(t, expectedCode, errorResp.Code)
}
```

## Migration Strategy

### Phase 1: Core Infrastructure
1. Create `testdata` package with builders
2. Create `testscenario` package with scenario builders
3. Create `testassert` package with assertion helpers
4. Migrate 2-3 critical test files to validate approach

### Phase 2: Systematic Migration
1. Migrate POI handler tests
2. Migrate Session handler tests
3. Migrate Service layer tests
4. Migrate Repository layer tests

### Phase 3: Integration Tests
1. Add `testdb` package for database management
2. Convert critical paths to integration tests
3. Add performance benchmarks

## Benefits

### Before (Current State)
```go
// 50+ lines of setup for one test
// Context type errors
// Brittle to interface changes
// Hard to understand intent
// Repetitive across files
```

### After (Target State)
```go
// 5-10 lines for same test
// Context handled automatically
// Resilient to interface changes
// Clear business intent
// Reusable across files
```

## Success Metrics

1. **Lines of Test Code**: Reduce by 60-70%
2. **Test Maintenance**: Interface changes affect <5 files instead of 50+
3. **Developer Velocity**: New tests take minutes instead of hours
4. **Test Reliability**: Eliminate context-type and mock-setup errors
5. **Code Coverage**: Maintain or improve coverage with better tests
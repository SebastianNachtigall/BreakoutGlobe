# Database Integration Testing Infrastructure

## Overview

This document describes the database integration testing infrastructure that provides isolated, real PostgreSQL databases for testing repository layer functionality with complete test isolation and automatic cleanup.

## Key Features

### 🔒 **Complete Test Isolation**
- Each test gets its own unique PostgreSQL database
- No shared state between tests
- Automatic cleanup after test completion
- Parallel test execution support

### 🚀 **Easy Setup and Usage**
- Single function call: `testdata.Setup(t)`
- Automatic migration execution
- Built-in fixture seeding with builders
- Fluent test data builders integration

### 🛡️ **Production-Like Testing**
- Real PostgreSQL database (not in-memory)
- Actual SQL queries and constraints
- Transaction behavior validation
- Performance benchmarking capability

## Usage Examples

### Basic Database Integration Test

```go
func TestRepository_Integration(t *testing.T) {
    if testing.Short() {
        t.Skip("Skipping database integration test in short mode")
    }
    
    // Setup isolated test database
    testDB := testdata.Setup(t)
    repo := NewRepository(testDB.DB)
    
    // Use real database operations
    entity := testdata.NewDatabasePOI().WithName("Test").Build()
    err := repo.Create(context.Background(), entity)
    
    // Verify with real database queries
    require.NoError(t, err)
    
    var count int64
    testDB.DB.Model(&models.POI{}).Count(&count)
    assert.Equal(t, int64(1), count)
}
```

### Test Data Seeding

```go
func TestRepository_WithFixtures(t *testing.T) {
    testDB := testdata.Setup(t)
    
    // Seed test data using builders
    poi1 := testdata.NewDatabasePOI().WithName("Coffee Shop").WithMapID("map-1").Build()
    poi2 := testdata.NewDatabasePOI().WithName("Park Bench").WithMapID("map-1").Build()
    err := testDB.SeedFixtures(poi1, poi2)
    require.NoError(t, err)
    
    // Test with real data
    repo := NewPOIRepository(testDB.DB)
    foundPOIs, err := repo.GetByMapID(context.Background(), "map-1")
    
    require.NoError(t, err)
    assert.Len(t, foundPOIs, 2)
}
```

### Performance Benchmarking

```go
func BenchmarkRepository_Create(b *testing.B) {
    testDB := testdata.Setup(b)
    repo := NewRepository(testDB.DB)
    
    b.ResetTimer()
    for i := 0; i < b.N; i++ {
        entity := testdata.NewDatabasePOI().
            WithName(fmt.Sprintf("Entity %d", i)).
            Build()
        if err := repo.Create(context.Background(), entity); err != nil {
            b.Fatalf("Failed to create entity: %v", err)
        }
    }
}
```

## Infrastructure Components

### TestDB Structure

```go
type TestDB struct {
    DB     *gorm.DB        // Database connection
    dbName string          // Unique database name
    t      TestingT        // Test context
}
```

### Key Methods

| Method | Purpose | Example |
|--------|---------|---------|
| `Setup(t)` | Create isolated test database | `testDB := testdata.Setup(t)` |
| `SeedFixtures(...)` | Load test data | `testDB.SeedFixtures(poi, session)` |
| `Clear(models...)` | Clear specific tables | `testDB.Clear(&models.POI{}, &models.Session{})` |
| `Transaction(fn)` | Execute in transaction | `testDB.Transaction(func(tx *gorm.DB) error {...})` |

### Database Fixture Builders

```go
// POI fixture builder
poi := testdata.NewDatabasePOI().
    WithName("Coffee Shop").
    WithMapID("map-123").
    WithPosition(40.7128, -74.0060).
    WithCreator("user-456").
    Build()

// Session fixture builder
session := testdata.NewDatabaseSession().
    WithUserID("user-123").
    WithMapID("map-456").
    WithPosition(40.7128, -74.0060).
    WithActive(true).
    Build()
```

## Environment Configuration

### Environment Variables

```bash
# Database connection settings
TEST_DB_HOST=localhost      # Default: localhost
TEST_DB_PORT=5432          # Default: 5432
TEST_DB_USER=postgres      # Default: postgres
TEST_DB_PASSWORD=postgres  # Default: postgres
TEST_DB_SSLMODE=disable    # Default: disable
```

### Docker Compose Setup

```bash
# Start test database
make test-integration-setup

# Run integration tests
make test-integration

# Cleanup
make test-integration-teardown
```

## Test Execution Patterns

### Unit Tests vs Integration Tests

```go
// Unit test - fast, uses mocks
func TestService_CreatePOI_Unit(t *testing.T) {
    mockRepo := new(MockPOIRepository)
    service := NewPOIService(mockRepo)
    // ... mock expectations and test
}

// Integration test - slower, uses real database
func TestRepository_CreatePOI_Integration(t *testing.T) {
    if testing.Short() {
        t.Skip("Skipping database integration test in short mode")
    }
    
    testDB := testdata.Setup(t)
    repo := NewPOIRepository(testDB.DB)
    // ... real database operations
}
```

### Test Isolation Verification

```go
func TestIsolation(t *testing.T) {
    t.Run("Test1", func(t *testing.T) {
        testDB := testdata.Setup(t)
        poi := testdata.NewDatabasePOI().WithName("Test1 POI").Build()
        testDB.SeedFixtures(poi)
        
        var count int64
        testDB.DB.Model(&models.POI{}).Count(&count)
        assert.Equal(t, int64(1), count)
    })
    
    t.Run("Test2", func(t *testing.T) {
        testDB := testdata.Setup(t)
        // Starts with empty database - complete isolation
        var count int64
        testDB.DB.Model(&models.POI{}).Count(&count)
        assert.Equal(t, int64(0), count)
    })
}
```

## Performance Considerations

### Database Creation Overhead
- Each test creates a new database (~50-100ms overhead)
- Use `testing.Short()` to skip in development
- Parallel execution reduces total time

### Optimization Strategies

```go
// Skip integration tests in short mode
if testing.Short() {
    t.Skip("Skipping database integration test in short mode")
}

// Use subtests for related operations
func TestPOIRepository_CRUD_Integration(t *testing.T) {
    testDB := testdata.Setup(t)
    repo := NewPOIRepository(testDB.DB)
    
    t.Run("Create", func(t *testing.T) {
        // Test creation
    })
    
    t.Run("Read", func(t *testing.T) {
        // Test retrieval
    })
    
    t.Run("Update", func(t *testing.T) {
        // Test updates
    })
    
    t.Run("Delete", func(t *testing.T) {
        // Test deletion
    })
}
```

## Running Integration Tests

### Local Development

```bash
# Run unit tests only (fast)
go test ./... -short

# Run all tests including integration (slower)
make test-integration

# Run specific integration tests
go test ./internal/repository -run=".*Integration.*"

# Run benchmarks
go test ./internal/repository -bench=".*Integration.*" -benchmem
```

### CI/CD Pipeline

```bash
# In GitHub Actions or similar
- name: Run Integration Tests
  run: |
    make test-integration-setup
    make test-integration
    make test-integration-teardown
```

## Best Practices

### 1. Test Naming Convention
- Use `*_Integration` suffix for integration tests
- Use descriptive names that indicate what's being tested
- Group related tests using subtests

### 2. Test Structure
```go
func TestRepository_Operation_Integration(t *testing.T) {
    if testing.Short() {
        t.Skip("Skipping database integration test in short mode")
    }
    
    // Setup
    testDB := testdata.Setup(t)
    repo := NewRepository(testDB.DB)
    ctx := context.Background()
    
    // Seed test data if needed
    fixture := testdata.NewDatabasePOI().WithName("Test").Build()
    err := testDB.SeedFixtures(fixture)
    require.NoError(t, err)
    
    // Execute
    result, err := repo.Operation(ctx, params)
    
    // Assert
    require.NoError(t, err)
    assert.Equal(t, expected, result)
}
```

### 3. Error Handling
- Always check for errors in setup and operations
- Use `require.NoError()` for critical operations
- Use `assert.Error()` for expected error scenarios

### 4. Data Verification
- Verify data integrity after operations
- Use database queries to confirm state changes
- Test edge cases and boundary conditions

## Troubleshooting

### Common Issues

1. **Database Connection Refused**
   ```bash
   # Ensure PostgreSQL is running
   make test-integration-setup
   
   # Check Docker containers
   docker compose -f compose.test.yml ps
   ```

2. **Migration Failures**
   ```bash
   # Check model definitions
   # Ensure all models are included in RunMigrations()
   ```

3. **Test Isolation Issues**
   ```bash
   # Verify each test gets unique database
   # Check cleanup is properly registered
   ```

4. **Performance Issues**
   ```bash
   # Use tmpfs for test databases (already configured)
   # Run tests in parallel where possible
   # Skip integration tests in development with -short
   ```

### Debugging Tips

```go
// Enable GORM logging for debugging
testDB.DB.Logger = logger.Default.LogMode(logger.Info)

// Check database state during test
var count int64
testDB.DB.Model(&models.POI{}).Count(&count)
t.Logf("POI count: %d", count)

// Inspect generated SQL
result := testDB.DB.Debug().Create(&poi)
```

## Integration with TDD Workflow

### Red-Green-Refactor with Integration Tests

1. **Red Phase**: Write failing integration test
   ```go
   func TestPOIRepository_NewFeature_Integration(t *testing.T) {
       testDB := testdata.Setup(t)
       repo := NewPOIRepository(testDB.DB)
       
       // This will fail because feature doesn't exist yet
       result, err := repo.NewFeature(ctx, params)
       require.NoError(t, err)
       assert.Equal(t, expected, result)
   }
   ```

2. **Green Phase**: Implement minimal code to pass
   ```go
   func (r *POIRepository) NewFeature(ctx context.Context, params Params) (*Result, error) {
       // Minimal implementation to make test pass
       return &Result{}, nil
   }
   ```

3. **Refactor Phase**: Improve implementation while keeping tests green
   ```go
   func (r *POIRepository) NewFeature(ctx context.Context, params Params) (*Result, error) {
       // Improved implementation with proper logic
       // Tests ensure behavior remains correct
   }
   ```

## Security Considerations

### Database Isolation
- Each test database is completely isolated
- No cross-test data contamination
- Automatic cleanup prevents data leaks

### Connection Security
- Uses environment variables for credentials
- SSL mode configurable (disabled for tests)
- No hardcoded passwords in code

### Test Data
- Use realistic but non-sensitive test data
- Avoid PII in test fixtures
- Use UUIDs for consistent ID generation

## Maintenance

### Regular Tasks
1. Update model migrations when schema changes
2. Review and optimize slow integration tests
3. Monitor test execution time and database usage
4. Update Docker images for security patches

### Monitoring
- Track integration test execution time
- Monitor database creation/cleanup success rate
- Watch for test isolation failures
- Review benchmark results for performance regressions

This infrastructure provides a solid foundation for reliable, maintainable database integration testing that supports proper TDD practices while ensuring complete test isolation and realistic testing conditions.
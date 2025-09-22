# Test Infrastructure Refactoring Implementation Plan

## Overview

This implementation plan transforms our brittle test infrastructure into a resilient, maintainable system that supports proper TDD practices. This work is a prerequisite for the user-profile-system implementation.

## Implementation Tasks

- [x] 1. Core Test Infrastructure Foundation
  - [x] 1.1 Create testdata package with model builders
    - Write tests for UserBuilder with fluent API (WithEmail, WithRole, etc.)
    - Implement UserBuilder with UUID generation and sensible defaults
    - Write tests for POIBuilder with relationship support (WithCreator, WithMap)
    - Implement POIBuilder with position and participant defaults
    - Write tests for SessionBuilder with user and map relationships
    - Implement SessionBuilder with position and timing defaults
    - Write tests for MapBuilder with ownership and settings
    - Implement MapBuilder with default configurations
    - Add UUID utility functions (GenerateUUID, ParseUUID) with tests
    - _Requirements: 2.1, 2.2, 2.3_

  - [x] 1.2 Create mock abstraction layer
    - Write tests for MockSetup with automatic context handling
    - Implement MockSetup that hides context.backgroundCtx complexity
    - Write tests for MockPOIServiceBuilder with fluent expectations
    - Implement MockPOIServiceBuilder with ExpectCreate, ExpectGet methods
    - Write tests for MockRateLimiterBuilder with success/failure scenarios
    - Implement MockRateLimiterBuilder with ExpectCheck, ExpectHeaders methods
    - Write tests for MockSessionServiceBuilder with CRUD operations
    - Implement MockSessionServiceBuilder with session lifecycle methods
    - Add automatic mock verification in test cleanup
    - _Requirements: 3.1, 3.2, 3.3_

  - [x] 1.3 Create assertion helpers package
    - Write tests for POIResponse assertion helper with field validation
    - Implement POIResponse helper that ignores timestamps and focuses on business data
    - Write tests for HTTPStatus helper with detailed error messages
    - Implement HTTPStatus helper that shows response body on failure
    - Write tests for ErrorResponse helper with error code validation
    - Implement ErrorResponse helper that validates error structure and codes
    - Write tests for SessionResponse helper with user and map validation
    - Implement SessionResponse helper for session-specific assertions
    - Add UUID assertion helpers for consistent ID validation
    - _Requirements: 5.1, 5.2, 5.3_

- [ ] 2. Test Scenario Builders
  - [ ] 2.1 Implement POI test scenario builder
    - Write tests for POITestScenario with fluent configuration
    - Implement POITestScenario with WithValidUser, WithMap methods
    - Write tests for POI creation scenarios (success, rate limit, validation errors)
    - Implement ExpectRateLimitSuccess, ExpectCreationSuccess methods
    - Write tests for POI join/leave scenarios with capacity and permissions
    - Implement ExpectJoinSuccess, ExpectCapacityExceeded methods
    - Write tests for POI retrieval scenarios with filtering and bounds
    - Implement ExpectGetSuccess, ExpectNotFound methods
    - Add CreatePOI, JoinPOI, GetPOI execution methods with HTTP handling
    - _Requirements: 1.1, 1.2, 1.3_

  - [ ] 2.2 Implement Session test scenario builder
    - Write tests for SessionTestScenario with user and map setup
    - Implement SessionTestScenario with WithExistingUser, WithMap methods
    - Write tests for session creation scenarios (success, rate limit, validation)
    - Implement ExpectSessionCreation, ExpectRateLimit methods
    - Write tests for avatar position update scenarios with permissions
    - Implement ExpectPositionUpdate, ExpectPermissionDenied methods
    - Write tests for session heartbeat and cleanup scenarios
    - Implement ExpectHeartbeat, ExpectCleanup methods
    - Add CreateSession, UpdatePosition, Heartbeat execution methods
    - _Requirements: 1.1, 1.2, 1.3_

  - [ ] 2.3 Implement WebSocket test scenario builder
    - Write tests for WebSocketTestScenario with connection management
    - Implement WebSocketTestScenario with connection setup and teardown
    - Write tests for real-time event scenarios (avatar movement, POI updates)
    - Implement ExpectAvatarMovement, ExpectPOIBroadcast methods
    - Write tests for connection error and reconnection scenarios
    - Implement ExpectConnectionError, ExpectReconnection methods
    - Write tests for message queuing and delivery scenarios
    - Implement ExpectMessageDelivery, ExpectQueueing methods
    - Add WebSocket client simulation and event verification
    - _Requirements: 1.1, 1.2, 1.3_

- [ ] 3. Migrate Existing Tests
  - [ ] 3.1 Migrate POI handler tests
    - Write tests for migrated CreatePOI using new scenario builders
    - Migrate TestCreatePOI to use POITestScenario with 80% less code
    - Write tests for migrated GetPOI using new assertion helpers
    - Migrate TestGetPOI to use testassert.POIResponse validation
    - Write tests for migrated JoinPOI using new mock abstractions
    - Migrate TestJoinPOI to use MockSetup with automatic context handling
    - Write tests for migrated error scenarios using new error builders
    - Migrate error tests to use ExpectCapacityExceeded, ExpectNotFound patterns
    - Verify all POI handler tests pass with new infrastructure
    - _Requirements: 1.1, 1.2, 1.3_

  - [ ] 3.2 Migrate Session handler tests
    - Write tests for migrated CreateSession using SessionTestScenario
    - Migrate TestCreateSession to use new builder pattern with validation
    - Write tests for migrated UpdateAvatarPosition using position scenarios
    - Migrate TestUpdateAvatarPosition to use ExpectPositionUpdate pattern
    - Write tests for migrated GetSession using new assertion helpers
    - Migrate TestGetSession to use testassert.SessionResponse validation
    - Write tests for migrated session lifecycle using scenario builders
    - Migrate session heartbeat and cleanup tests to new patterns
    - Verify all session handler tests pass with new infrastructure
    - _Requirements: 1.1, 1.2, 1.3_

  - [ ] 3.3 Migrate Service layer tests
    - Write tests for migrated POIService using testdata builders
    - Migrate POIService tests to use NewPOI().WithCreator() patterns
    - Write tests for migrated SessionService using relationship builders
    - Migrate SessionService tests to use NewSession().WithUser().WithMap()
    - Write tests for migrated RateLimiter using scenario-based testing
    - Migrate RateLimiter tests to use action-based scenario builders
    - Write tests for migrated repository layer using database fixtures
    - Migrate repository tests to use testdb.Setup() with real database
    - Verify all service tests pass with improved readability and maintainability
    - _Requirements: 2.1, 2.2, 2.3_

- [ ] 4. Integration Test Infrastructure
  - [ ] 4.1 Implement test database management
    - Write tests for testdb.Setup with isolated database creation
    - Implement testdb.Setup that creates unique test databases per test
    - Write tests for automatic cleanup with t.Cleanup integration
    - Implement automatic database cleanup that runs after each test
    - Write tests for migration running with AutoMigrate integration
    - Implement automatic migration execution for test databases
    - Write tests for fixture seeding with SeedFixtures helper
    - Implement SeedFixtures that loads test data into database
    - Add Docker support for containerized test databases in CI
    - _Requirements: 4.1, 4.2, 4.3_

  - [ ] 4.2 Create integration test examples
    - Write integration test for POI creation with real database persistence
    - Implement TestPOICreation_Integration using real services and database
    - Write integration test for session management with real Redis
    - Implement TestSessionLifecycle_Integration with real state management
    - Write integration test for WebSocket communication with real connections
    - Implement TestWebSocketEvents_Integration with real message passing
    - Write integration test for rate limiting with real Redis backend
    - Implement TestRateLimit_Integration with real rate limit enforcement
    - Add performance benchmarks for critical integration test paths
    - _Requirements: 4.1, 4.2, 4.3_

  - [ ] 4.3 Implement test performance optimization
    - Write tests for parallel test execution with database isolation
    - Implement parallel-safe test database management
    - Write tests for test data caching with fixture reuse
    - Implement fixture caching to speed up test setup
    - Write tests for selective integration testing with unit test fallbacks
    - Implement smart test selection based on changed code
    - Write tests for CI/CD integration with containerized dependencies
    - Implement Docker Compose setup for CI test environments
    - Add test execution time monitoring and optimization
    - _Requirements: 4.1, 4.2, 4.3_

- [ ] 5. Documentation and Training
  - [ ] 5.1 Create test infrastructure documentation
    - Write comprehensive guide for using new test builders
    - Document POITestScenario, SessionTestScenario usage patterns
    - Write examples for common testing scenarios (CRUD, errors, edge cases)
    - Document testdata builders with relationship examples
    - Write guide for choosing between unit and integration tests
    - Document when to use mocks vs real services
    - Write troubleshooting guide for common test issues
    - Document debugging techniques for test failures
    - Add code examples for all major testing patterns
    - _Requirements: 6.1, 6.2, 6.3_

  - [ ] 5.2 Create migration guide for existing tests
    - Write step-by-step guide for migrating old tests to new patterns
    - Document common migration patterns and gotchas
    - Write examples showing before/after test code
    - Document how to handle complex mock scenarios in new system
    - Write guide for maintaining test coverage during migration
    - Document verification steps for successful migration
    - Write troubleshooting guide for migration issues
    - Document rollback procedures if migration fails
    - Add checklist for validating migrated tests
    - _Requirements: 6.1, 6.2, 6.3_

  - [ ] 5.3 Establish testing best practices
    - Write coding standards for new test infrastructure
    - Document naming conventions for test scenarios and builders
    - Write guidelines for test organization and file structure
    - Document when to write unit vs integration vs end-to-end tests
    - Write performance guidelines for test execution time
    - Document CI/CD integration best practices
    - Write code review checklist for test quality
    - Document maintenance procedures for test infrastructure
    - Add examples of good and bad testing patterns
    - _Requirements: 6.1, 6.2, 6.3_

- [ ] 6. Validation and Cleanup
  - [ ] 6.1 Comprehensive test suite validation
    - Run full test suite to ensure 100% pass rate with new infrastructure
    - Verify test execution time remains acceptable (<2 minutes for unit tests)
    - Run tests in parallel to ensure no race conditions or conflicts
    - Verify test isolation with no cross-test dependencies
    - Run tests multiple times to ensure consistency and reliability
    - Verify mock cleanup and no memory leaks in test execution
    - Run integration tests to ensure database and Redis connectivity
    - Verify CI/CD pipeline compatibility with new test infrastructure
    - Add test coverage reporting to ensure no regression in coverage
    - _Requirements: 1.1, 2.1, 3.1, 4.1, 5.1_

  - [ ] 6.2 Remove legacy test infrastructure
    - Remove old mock setup patterns and replace with new abstractions
    - Delete redundant test utilities that are replaced by new builders
    - Remove manual context handling code that's now automated
    - Clean up test files that have been migrated to new patterns
    - Remove unused test dependencies and imports
    - Update test documentation to reflect new patterns only
    - Remove legacy test examples and replace with new pattern examples
    - Clean up CI/CD configuration to use new test infrastructure
    - Archive old test patterns for reference but remove from active codebase
    - _Requirements: 6.1, 6.2, 6.3_

  - [ ] 6.3 Performance and maintainability validation
    - Measure test execution time improvement (target: 30% faster)
    - Measure lines of test code reduction (target: 60% reduction)
    - Measure developer velocity improvement for writing new tests
    - Validate that interface changes affect minimal test files
    - Measure test failure diagnosis time improvement
    - Validate test readability and understandability for new developers
    - Measure maintenance effort reduction for test updates
    - Validate CI/CD pipeline stability and reliability
    - Document performance metrics and maintainability improvements
    - _Requirements: 1.1, 2.1, 3.1, 4.1, 5.1_

## Success Criteria

1. **Resilience**: Interface changes require updates to <5 test files instead of 50+
2. **Readability**: New developers can understand test intent without studying mocks
3. **Maintainability**: Adding new test scenarios takes minutes, not hours
4. **Performance**: Unit tests remain fast (<100ms each), integration tests <1s each
5. **Coverage**: Maintain or improve test coverage with more meaningful tests
6. **Developer Experience**: Writing tests becomes enjoyable rather than painful

## Dependencies

- Go testing framework and testify for assertions
- GORM for database operations in integration tests
- Redis for caching and rate limiting in integration tests
- Docker for containerized test dependencies
- UUID library for consistent ID generation
- HTTP testing utilities for handler tests

## Risks and Mitigations

**Risk**: Migration takes longer than expected
**Mitigation**: Implement incrementally, validate each phase before proceeding

**Risk**: New patterns are not adopted consistently
**Mitigation**: Provide clear documentation, examples, and enforce in code reviews

**Risk**: Integration tests become too slow
**Mitigation**: Keep unit tests as primary, use integration tests selectively for critical paths

**Risk**: Test infrastructure becomes over-engineered
**Mitigation**: Focus on solving actual pain points, avoid premature optimization

**Risk**: Regression in test coverage during migration
**Mitigation**: Monitor coverage continuously, migrate tests incrementally
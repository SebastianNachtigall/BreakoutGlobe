package server

import (
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"breakoutglobe/internal/config"
)

// Test scenario for avatar file serving
type AvatarServingTestScenario struct {
	server     *Server
	router     *gin.Engine
	testDir    string
	testFile   string
	testFileContent []byte
}

func newAvatarServingScenario(t *testing.T) *AvatarServingTestScenario {
	t.Helper()
	
	// Create test server without database (for file serving tests)
	cfg := &config.Config{GinMode: "test"}
	srv := New(cfg)
	router := gin.New()
	
	// Setup routes
	api := router.Group("/api")
	api.GET("/users/avatar/:filename", srv.serveAvatar)
	
	// Create temporary test directory
	testDir := filepath.Join("uploads", "avatars")
	err := os.MkdirAll(testDir, 0755)
	require.NoError(t, err)
	
	// Create test file content (minimal PNG)
	testFileContent := []byte{
		0x89, 0x50, 0x4E, 0x47, 0x0D, 0x0A, 0x1A, 0x0A, 0x00, 0x00, 0x00, 0x0D,
		0x49, 0x48, 0x44, 0x52, 0x00, 0x00, 0x00, 0x01, 0x00, 0x00, 0x00, 0x01,
		0x08, 0x02, 0x00, 0x00, 0x00, 0x90, 0x77, 0x53, 0xDE, 0x00, 0x00, 0x00,
		0x0C, 0x49, 0x44, 0x41, 0x54, 0x08, 0xD7, 0x63, 0xF8, 0x00, 0x00, 0x00,
		0x01, 0x00, 0x01, 0x5C, 0xC2, 0x8A, 0x8B, 0x00, 0x00, 0x00, 0x00, 0x49,
		0x45, 0x4E, 0x44, 0xAE, 0x42, 0x60, 0x82,
	}
	
	testFile := filepath.Join(testDir, "test-avatar.png")
	err = os.WriteFile(testFile, testFileContent, 0644)
	require.NoError(t, err)
	
	return &AvatarServingTestScenario{
		server:          srv,
		router:          router,
		testDir:         testDir,
		testFile:        testFile,
		testFileContent: testFileContent,
	}
}

func (s *AvatarServingTestScenario) cleanup(t *testing.T) {
	t.Helper()
	os.RemoveAll("uploads")
}

// expectAvatarFileServing tests successful avatar file serving
func (s *AvatarServingTestScenario) expectAvatarFileServing(t *testing.T, filename string) *httptest.ResponseRecorder {
	t.Helper()
	
	req := httptest.NewRequest(http.MethodGet, "/api/users/avatar/"+filename, nil)
	recorder := httptest.NewRecorder()
	
	s.router.ServeHTTP(recorder, req)
	
	return recorder
}

// expectAvatarFileValidation tests security validation
func (s *AvatarServingTestScenario) expectAvatarFileValidation(t *testing.T, filename string, expectedStatus int) *httptest.ResponseRecorder {
	t.Helper()
	
	req := httptest.NewRequest(http.MethodGet, "/api/users/avatar/"+filename, nil)
	recorder := httptest.NewRecorder()
	
	s.router.ServeHTTP(recorder, req)
	
	assert.Equal(t, expectedStatus, recorder.Code)
	return recorder
}

func TestAvatarFileServing(t *testing.T) {
	scenario := newAvatarServingScenario(t)
	defer scenario.cleanup(t)

	t.Run("should serve existing avatar files successfully", func(t *testing.T) {
		// Act - expectAvatarFileServing()
		recorder := scenario.expectAvatarFileServing(t, "test-avatar.png")
		
		// Assert
		assert.Equal(t, http.StatusOK, recorder.Code)
		assert.Equal(t, scenario.testFileContent, recorder.Body.Bytes())
	})

	t.Run("should return 404 for non-existent files", func(t *testing.T) {
		// Act - expectAvatarFileValidation()
		recorder := scenario.expectAvatarFileValidation(t, "non-existent.png", http.StatusNotFound)
		
		// Assert
		assert.Contains(t, recorder.Body.String(), "avatar not found")
	})

	t.Run("should return 404 for empty filename (route not matched)", func(t *testing.T) {
		// Act - empty filename doesn't match the route pattern
		recorder := scenario.expectAvatarFileValidation(t, "", http.StatusNotFound)
		
		// Assert - Gin returns 404 when route doesn't match
		assert.Contains(t, recorder.Body.String(), "404")
	})

	t.Run("should prevent path traversal attacks", func(t *testing.T) {
		// Test path traversal attempts that would reach our handler
		// (Gin router blocks some patterns before they reach the handler)
		pathTraversalAttempts := []struct {
			filename       string
			expectedStatus int
			description    string
		}{
			{"..\\secret.txt", http.StatusBadRequest, "backslash traversal"},
			{"file..txt", http.StatusBadRequest, "double dot in filename"},
		}
		
		for _, attempt := range pathTraversalAttempts {
			t.Run(attempt.description, func(t *testing.T) {
				// Act
				recorder := scenario.expectAvatarFileValidation(t, attempt.filename, attempt.expectedStatus)
				
				// Assert
				if attempt.expectedStatus == http.StatusBadRequest {
					assert.Contains(t, recorder.Body.String(), "invalid filename")
				}
			})
		}
		
		// Test that Gin router blocks obvious path traversal (returns 404)
		t.Run("gin router blocks obvious traversal", func(t *testing.T) {
			recorder := scenario.expectAvatarFileValidation(t, "../secret.txt", http.StatusNotFound)
			// Gin returns 404 for routes it can't match, which is good security
			assert.Contains(t, recorder.Body.String(), "404")
		})
	})

	t.Run("should validate file extensions", func(t *testing.T) {
		// Arrange - create file with invalid extension
		invalidFile := filepath.Join(scenario.testDir, "test.exe")
		err := os.WriteFile(invalidFile, []byte("executable"), 0644)
		require.NoError(t, err)
		defer os.Remove(invalidFile)
		
		// Act - expectAvatarFileValidation()
		recorder := scenario.expectAvatarFileValidation(t, "test.exe", http.StatusBadRequest)
		
		// Assert
		assert.Contains(t, recorder.Body.String(), "invalid file type")
	})
}
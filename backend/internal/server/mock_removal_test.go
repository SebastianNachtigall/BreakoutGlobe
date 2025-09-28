package server

import (
	"testing"
	"net/http"
	"net/http/httptest"
	"strings"

	"breakoutglobe/internal/config"
	"github.com/stretchr/testify/assert"
)

// TestMockHandlersRemoved verifies that mock handlers are no longer accessible in test mode
func TestMockHandlersRemoved(t *testing.T) {
	// Create server in test mode (no database/redis connections)
	cfg := &config.Config{
		GinMode: "test",
	}
	server := New(cfg)

	// These endpoints should return 404 when mock handlers are removed
	mockEndpoints := []struct {
		name   string
		method string
		path   string
		body   string
	}{
		{
			name:   "CreateSession_Mock_Removed",
			method: "POST",
			path:   "/api/sessions",
			body:   `{"userId":"test-user","mapId":"default-map","avatarPosition":{"lat":40.7128,"lng":-74.0060}}`,
		},
		{
			name:   "GetSession_Mock_Removed",
			method: "GET", 
			path:   "/api/sessions/test-session",
		},
		{
			name:   "UpdateAvatarPosition_Mock_Removed",
			method: "PUT",
			path:   "/api/sessions/test-session/avatar", 
			body:   `{"position":{"lat":40.7130,"lng":-74.0062}}`,
		},
		{
			name:   "GetPOIs_Mock_Removed",
			method: "GET",
			path:   "/api/pois?mapId=default-map",
		},
		{
			name:   "CreatePOI_Mock_Removed",
			method: "POST",
			path:   "/api/pois",
			body:   `{"mapId":"default-map","name":"Test POI","description":"Test description","position":{"lat":40.7128,"lng":-74.0060},"createdBy":"test-user","maxParticipants":10}`,
		},
		{
			name:   "JoinPOI_Mock_Removed",
			method: "POST", 
			path:   "/api/pois/test-poi/join",
			body:   `{"sessionId":"test-session"}`,
		},
		{
			name:   "LeavePOI_Mock_Removed",
			method: "POST",
			path:   "/api/pois/test-poi/leave", 
			body:   `{"sessionId":"test-session"}`,
		},
	}

	for _, tt := range mockEndpoints {
		t.Run(tt.name, func(t *testing.T) {
			var req *http.Request
			if tt.body != "" {
				req = httptest.NewRequest(tt.method, tt.path, strings.NewReader(tt.body))
				req.Header.Set("Content-Type", "application/json")
			} else {
				req = httptest.NewRequest(tt.method, tt.path, nil)
			}

			w := httptest.NewRecorder()
			server.router.ServeHTTP(w, req)

			// Mock handlers should be removed, so these should return 404
			assert.Equal(t, http.StatusNotFound, w.Code, "Mock endpoint %s %s should return 404 after removal", tt.method, tt.path)
			t.Logf("✅ Mock endpoint removed: %s %s returns 404", tt.method, tt.path)
		})
	}
}
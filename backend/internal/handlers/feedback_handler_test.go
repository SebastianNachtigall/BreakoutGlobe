package handlers

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestFeedbackHandler_SubmitFeedback_Validation(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name           string
		request        FeedbackRequest
		expectedStatus int
		expectedCode   string
	}{
		{
			name: "title too short",
			request: FeedbackRequest{
				Title:       "Bug",
				Description: "This is a valid description that is long enough.",
				Category:    "bug",
			},
			expectedStatus: http.StatusBadRequest,
			expectedCode:   "INVALID_TITLE",
		},
		{
			name: "description too short",
			request: FeedbackRequest{
				Title:       "Valid Title",
				Description: "Too short",
				Category:    "feature",
			},
			expectedStatus: http.StatusBadRequest,
			expectedCode:   "INVALID_DESCRIPTION",
		},
		{
			name: "invalid category",
			request: FeedbackRequest{
				Title:       "Valid Title",
				Description: "This is a valid description that is long enough.",
				Category:    "invalid",
			},
			expectedStatus: http.StatusBadRequest,
			expectedCode:   "INVALID_CATEGORY",
		},
		{
			name: "missing title",
			request: FeedbackRequest{
				Description: "This is a valid description that is long enough.",
				Category:    "feature",
			},
			expectedStatus: http.StatusBadRequest,
			expectedCode:   "INVALID_REQUEST",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup
			router := gin.New()
			handler := NewFeedbackHandler()
			handler.RegisterRoutes(router)

			// Create request
			body, _ := json.Marshal(tt.request)
			req := httptest.NewRequest(http.MethodPost, "/api/feedback", bytes.NewBuffer(body))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()

			// Execute
			router.ServeHTTP(w, req)

			// Assert
			assert.Equal(t, tt.expectedStatus, w.Code)

			if tt.expectedCode != "" {
				var response ErrorResponse
				err := json.Unmarshal(w.Body.Bytes(), &response)
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedCode, response.Code)
			}
		})
	}
}

func TestFeedbackHandler_GitHubNotConfigured(t *testing.T) {
	gin.SetMode(gin.TestMode)

	// Ensure GitHub env vars are not set
	os.Unsetenv("GITHUB_TOKEN")
	os.Unsetenv("GITHUB_REPO_OWNER")
	os.Unsetenv("GITHUB_REPO_NAME")

	// Setup
	router := gin.New()
	handler := NewFeedbackHandler()
	handler.RegisterRoutes(router)

	// Create valid request
	request := FeedbackRequest{
		Title:       "Test Feature",
		Description: "This is a test feature request with enough characters.",
		Category:    "feature",
	}
	body, _ := json.Marshal(request)
	req := httptest.NewRequest(http.MethodPost, "/api/feedback", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	// Execute
	router.ServeHTTP(w, req)

	// Assert
	assert.Equal(t, http.StatusServiceUnavailable, w.Code)

	var response ErrorResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, "GITHUB_NOT_CONFIGURED", response.Code)
}

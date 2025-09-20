package server

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"breakoutglobe/internal/config"
	"github.com/stretchr/testify/assert"
)

func TestServer_HealthCheck(t *testing.T) {
	cfg := &config.Config{
		GinMode: "test",
	}
	
	server := New(cfg)
	
	req, _ := http.NewRequest("GET", "/health", nil)
	w := httptest.NewRecorder()
	
	server.router.ServeHTTP(w, req)
	
	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "ok")
}

func TestServer_APIStatus(t *testing.T) {
	cfg := &config.Config{
		GinMode: "test",
	}
	
	server := New(cfg)
	
	req, _ := http.NewRequest("GET", "/api/status", nil)
	w := httptest.NewRecorder()
	
	server.router.ServeHTTP(w, req)
	
	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "BreakoutGlobe API is running")
}
package handlers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"

	"github.com/gin-gonic/gin"
)

// FeedbackHandler handles feedback submission to GitHub Issues
type FeedbackHandler struct {
	githubToken string
	repoOwner   string
	repoName    string
}

// NewFeedbackHandler creates a new FeedbackHandler
func NewFeedbackHandler() *FeedbackHandler {
	return &FeedbackHandler{
		githubToken: os.Getenv("GITHUB_TOKEN"),
		repoOwner:   os.Getenv("GITHUB_REPO_OWNER"),
		repoName:    os.Getenv("GITHUB_REPO_NAME"),
	}
}

// RegisterRoutes registers feedback routes
func (h *FeedbackHandler) RegisterRoutes(router *gin.Engine) {
	api := router.Group("/api")
	{
		api.POST("/feedback", h.SubmitFeedback)
	}
}

// FeedbackRequest represents the feedback submission request
type FeedbackRequest struct {
	Title       string `json:"title" binding:"required"`
	Description string `json:"description" binding:"required"`
	Category    string `json:"category" binding:"required"`
}

// GitHubIssueRequest represents the GitHub API issue creation request
type GitHubIssueRequest struct {
	Title  string   `json:"title"`
	Body   string   `json:"body"`
	Labels []string `json:"labels"`
}

// SubmitFeedback handles POST /api/feedback
func (h *FeedbackHandler) SubmitFeedback(c *gin.Context) {
	var req FeedbackRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Code:    "INVALID_REQUEST",
			Message: "Invalid request format",
			Details: err.Error(),
		})
		return
	}

	// Validate category
	validCategories := map[string]bool{
		"bug":         true,
		"feature":     true,
		"improvement": true,
	}
	if !validCategories[req.Category] {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Code:    "INVALID_CATEGORY",
			Message: "Invalid category. Must be: bug, feature, or improvement",
		})
		return
	}

	// Validate title and description length
	if len(req.Title) < 5 || len(req.Title) > 100 {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Code:    "INVALID_TITLE",
			Message: "Title must be between 5 and 100 characters",
		})
		return
	}
	if len(req.Description) < 10 || len(req.Description) > 1000 {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Code:    "INVALID_DESCRIPTION",
			Message: "Description must be between 10 and 1000 characters",
		})
		return
	}

	// Check if GitHub integration is configured
	if h.githubToken == "" || h.repoOwner == "" || h.repoName == "" {
		c.JSON(http.StatusServiceUnavailable, ErrorResponse{
			Code:    "GITHUB_NOT_CONFIGURED",
			Message: "GitHub integration is not configured",
		})
		return
	}

	// Create GitHub issue
	if err := h.createGitHubIssue(req); err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Code:    "GITHUB_ERROR",
			Message: "Failed to create GitHub issue",
			Details: err.Error(),
		})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"success": true,
		"message": "Feedback submitted successfully",
	})
}

func (h *FeedbackHandler) createGitHubIssue(req FeedbackRequest) error {
	// Map category to GitHub label
	labelMap := map[string]string{
		"bug":         "bug",
		"feature":     "enhancement",
		"improvement": "enhancement",
	}

	// Format issue body
	body := fmt.Sprintf("## Description\n\n%s\n\n---\n\n*Submitted via in-app feedback*", req.Description)

	// Create GitHub issue request
	issueReq := GitHubIssueRequest{
		Title:  req.Title,
		Body:   body,
		Labels: []string{labelMap[req.Category], "user-feedback"},
	}

	// Marshal request
	jsonData, err := json.Marshal(issueReq)
	if err != nil {
		return fmt.Errorf("failed to marshal request: %w", err)
	}

	// Create HTTP request
	url := fmt.Sprintf("https://api.github.com/repos/%s/%s/issues", h.repoOwner, h.repoName)
	httpReq, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	// Set headers
	httpReq.Header.Set("Authorization", fmt.Sprintf("Bearer %s", h.githubToken))
	httpReq.Header.Set("Accept", "application/vnd.github+json")
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("X-GitHub-Api-Version", "2022-11-28")

	// Send request
	client := &http.Client{}
	resp, err := client.Do(httpReq)
	if err != nil {
		return fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	// Check response
	if resp.StatusCode != http.StatusCreated {
		var errorBody map[string]interface{}
		json.NewDecoder(resp.Body).Decode(&errorBody)
		return fmt.Errorf("GitHub API error (status %d): %v", resp.StatusCode, errorBody)
	}

	return nil
}

// SanitizeInput removes potentially harmful characters from user input
func sanitizeInput(input string) string {
	// Remove null bytes and control characters
	input = strings.ReplaceAll(input, "\x00", "")
	return strings.TrimSpace(input)
}

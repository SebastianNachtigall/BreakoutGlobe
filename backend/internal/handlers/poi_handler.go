package handlers

import (
	"context"
	"errors"
	"mime/multipart"
	"net/http"
	"strconv"
	"strings"
	"time"

	"breakoutglobe/internal/models"
	"breakoutglobe/internal/services"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// POIServiceInterface defines the interface for POI service operations
type POIServiceInterface interface {
	CreatePOI(ctx context.Context, mapID, name, description string, position models.LatLng, createdBy string, maxParticipants int) (*models.POI, error)
	CreatePOIWithImage(ctx context.Context, mapID, name, description string, position models.LatLng, createdBy string, maxParticipants int, imageFile *multipart.FileHeader) (*models.POI, error)
	GetPOI(ctx context.Context, poiID string) (*models.POI, error)
	GetPOIsForMap(ctx context.Context, mapID string) ([]*models.POI, error)
	GetPOIsInBounds(ctx context.Context, mapID string, bounds services.POIBounds) ([]*models.POI, error)
	UpdatePOI(ctx context.Context, poiID string, updateData services.POIUpdateData) (*models.POI, error)
	DeletePOI(ctx context.Context, poiID string) error
	JoinPOI(ctx context.Context, poiID, userID string) error
	LeavePOI(ctx context.Context, poiID, userID string) error
	GetPOIParticipants(ctx context.Context, poiID string) ([]string, error)
	GetPOIParticipantCount(ctx context.Context, poiID string) (int, error)
	GetPOIParticipantsWithInfo(ctx context.Context, poiID string) ([]services.POIParticipantInfo, error)
	GetUserPOIs(ctx context.Context, userID string) ([]string, error)
	ValidatePOI(ctx context.Context, poiID string) (*models.POI, error)
	ClearAllPOIs(ctx context.Context, mapID string) error
}

// POIUserServiceInterface defines the interface for user service operations needed by POI handler
type POIUserServiceInterface interface {
	GetUser(ctx context.Context, userID string) (*models.User, error)
}

// POIHandler handles HTTP requests for POI operations
type POIHandler struct {
	poiService  POIServiceInterface
	userService POIUserServiceInterface
	rateLimiter services.RateLimiterInterface
}

// NewPOIHandler creates a new POIHandler instance
func NewPOIHandler(poiService POIServiceInterface, userService POIUserServiceInterface, rateLimiter services.RateLimiterInterface) *POIHandler {
	return &POIHandler{
		poiService:  poiService,
		userService: userService,
		rateLimiter: rateLimiter,
	}
}

// RegisterRoutes registers POI-related routes
func (h *POIHandler) RegisterRoutes(router *gin.Engine) {
	api := router.Group("/api")
	{
		// POI management
		api.GET("/pois", h.GetPOIs)
		api.POST("/pois", h.CreatePOI)
		api.GET("/pois/:poiId", h.GetPOI)
		api.PUT("/pois/:poiId", h.UpdatePOI)
		api.DELETE("/pois/:poiId", h.DeletePOI)
		
		// POI participation
		api.POST("/pois/:poiId/join", h.JoinPOI)
		api.POST("/pois/:poiId/leave", h.LeavePOI)
		api.GET("/pois/:poiId/participants", h.GetPOIParticipants)
		
		// Development endpoints (TODO: Remove in production)
		api.DELETE("/pois/dev/clear-all", h.ClearAllPOIs)
	}
}

// Request/Response DTOs

// CreatePOIRequest represents the request body for creating a POI
type CreatePOIRequest struct {
	MapID           string        `json:"mapId" binding:"required"`
	Name            string        `json:"name" binding:"required"`
	Description     string        `json:"description"`
	Position        models.LatLng `json:"position" binding:"required"`
	CreatedBy       string        `json:"createdBy" binding:"required"`
	MaxParticipants int           `json:"maxParticipants"`
}

// CreatePOIResponse represents the response for creating a POI
type CreatePOIResponse struct {
	ID              string        `json:"id"`
	MapID           string        `json:"mapId"`
	Name            string        `json:"name"`
	Description     string        `json:"description"`
	Position        models.LatLng `json:"position"`
	CreatedBy       string        `json:"createdBy"`
	MaxParticipants int           `json:"maxParticipants"`
	ImageURL        string        `json:"imageUrl,omitempty"`
	ThumbnailURL    string        `json:"thumbnailUrl,omitempty"`
	CreatedAt       time.Time     `json:"createdAt"`
}

// GetPOIResponse represents the response for getting a POI
type GetPOIResponse struct {
	ID              string        `json:"id"`
	MapID           string        `json:"mapId"`
	Name            string        `json:"name"`
	Description     string        `json:"description"`
	Position        models.LatLng `json:"position"`
	CreatedBy       string        `json:"createdBy"`
	MaxParticipants int           `json:"maxParticipants"`
	ImageURL        string        `json:"imageUrl,omitempty"`
	ThumbnailURL    string        `json:"thumbnailUrl,omitempty"`
	CreatedAt       time.Time     `json:"createdAt"`
}

// POIInfo represents POI information in list responses
type POIInfo struct {
	ID              string             `json:"id"`
	MapID           string             `json:"mapId"`
	Name            string             `json:"name"`
	Description     string             `json:"description"`
	Position        models.LatLng      `json:"position"`
	CreatedBy       string             `json:"createdBy"`
	MaxParticipants int                `json:"maxParticipants"`
	ParticipantCount int               `json:"participantCount"`
	Participants    []ParticipantInfo  `json:"participants"`
	ImageURL        string             `json:"imageUrl,omitempty"`
	ThumbnailURL    string             `json:"thumbnailUrl,omitempty"`
	
	// Discussion timer fields - backend only tracks when 2+ users are present
	DiscussionStartTime *time.Time `json:"discussionStartTime,omitempty"`
	IsDiscussionActive  bool       `json:"isDiscussionActive"`
	
	CreatedAt       time.Time          `json:"createdAt"`
}

// ParticipantInfo represents a participant in a POI
type ParticipantInfo struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	AvatarURL string `json:"avatarUrl"`
}

// GetPOIsResponse represents the response for getting POIs
type GetPOIsResponse struct {
	MapID string    `json:"mapId"`
	POIs  []POIInfo `json:"pois"`
	Count int       `json:"count"`
}

// UpdatePOIRequest represents the request body for updating a POI
type UpdatePOIRequest struct {
	Name            string `json:"name,omitempty"`
	Description     string `json:"description,omitempty"`
	MaxParticipants int    `json:"maxParticipants,omitempty"`
}

// UpdatePOIResponse represents the response for updating a POI
type UpdatePOIResponse struct {
	ID              string        `json:"id"`
	MapID           string        `json:"mapId"`
	Name            string        `json:"name"`
	Description     string        `json:"description"`
	Position        models.LatLng `json:"position"`
	CreatedBy       string        `json:"createdBy"`
	MaxParticipants int           `json:"maxParticipants"`
	ImageURL        string        `json:"imageUrl,omitempty"`
	ThumbnailURL    string        `json:"thumbnailUrl,omitempty"`
	CreatedAt       time.Time     `json:"createdAt"`
}

// JoinPOIRequest represents the request body for joining a POI
type JoinPOIRequest struct {
	UserID string `json:"userId" binding:"required"`
}

// JoinPOIResponse represents the response for joining a POI
type JoinPOIResponse struct {
	Success bool   `json:"success"`
	POIID   string `json:"poiId"`
	UserID  string `json:"userId"`
}

// LeavePOIRequest represents the request body for leaving a POI
type LeavePOIRequest struct {
	UserID string `json:"userId" binding:"required"`
}

// LeavePOIResponse represents the response for leaving a POI
type LeavePOIResponse struct {
	Success bool   `json:"success"`
	POIID   string `json:"poiId"`
	UserID  string `json:"userId"`
}

// GetPOIParticipantsResponse represents the response for getting POI participants
type GetPOIParticipantsResponse struct {
	POIID        string   `json:"poiId"`
	Participants []string `json:"participants"`
	Count        int      `json:"count"`
}

// GetPOIs handles GET /api/pois
func (h *POIHandler) GetPOIs(c *gin.Context) {
	mapID := c.Query("mapId")
	if mapID == "" {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Code:    "INVALID_REQUEST",
			Message: "mapId is required",
		})
		return
	}
	
	// Check if bounds are provided for spatial filtering
	minLatStr := c.Query("minLat")
	maxLatStr := c.Query("maxLat")
	minLngStr := c.Query("minLng")
	maxLngStr := c.Query("maxLng")
	
	var pois []*models.POI
	var err error
	
	if minLatStr != "" && maxLatStr != "" && minLngStr != "" && maxLngStr != "" {
		// Parse bounds
		minLat, err1 := strconv.ParseFloat(minLatStr, 64)
		maxLat, err2 := strconv.ParseFloat(maxLatStr, 64)
		minLng, err3 := strconv.ParseFloat(minLngStr, 64)
		maxLng, err4 := strconv.ParseFloat(maxLngStr, 64)
		
		if err1 != nil || err2 != nil || err3 != nil || err4 != nil {
			c.JSON(http.StatusBadRequest, ErrorResponse{
				Code:    "INVALID_REQUEST",
				Message: "Invalid bounds parameters",
			})
			return
		}
		
		bounds := services.POIBounds{
			MinLat: minLat,
			MaxLat: maxLat,
			MinLng: minLng,
			MaxLng: maxLng,
		}
		
		pois, err = h.poiService.GetPOIsInBounds(c, mapID, bounds)
	} else {
		// Get all POIs for the map
		pois, err = h.poiService.GetPOIsForMap(c, mapID)
	}
	
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Code:    "INTERNAL_ERROR",
			Message: "Failed to get POIs",
			Details: err.Error(),
		})
		return
	}
	
	// Convert to response format with participant information
	poiInfos := make([]POIInfo, len(pois))
	for i, poi := range pois {
		// Get participant count
		participantCount, err := h.poiService.GetPOIParticipantCount(c, poi.ID)
		if err != nil {
			// Log error but don't fail the request
			participantCount = 0
		}
		
		// Get participant details with avatar information
		participantsInfo, err := h.poiService.GetPOIParticipantsWithInfo(c, poi.ID)
		if err != nil {
			// Log error but don't fail the request
			participantsInfo = []services.POIParticipantInfo{}
		}
		
		// Convert to handler participant info format
		participants := make([]ParticipantInfo, len(participantsInfo))
		for j, participantInfo := range participantsInfo {
			participants[j] = ParticipantInfo{
				ID:        participantInfo.ID,
				Name:      participantInfo.Name,
				AvatarURL: participantInfo.AvatarURL,
			}
		}
		
		// Calculate discussion timer state
		isDiscussionActive := participantCount >= 2
		
		poiInfos[i] = POIInfo{
			ID:               poi.ID,
			MapID:            poi.MapID,
			Name:             poi.Name,
			Description:      poi.Description,
			Position:         poi.Position,
			CreatedBy:        poi.CreatedBy,
			MaxParticipants:  poi.MaxParticipants,
			ParticipantCount: participantCount,
			Participants:     participants,
			ImageURL:         poi.ImageURL,
			ThumbnailURL:     poi.ThumbnailURL,
			
			// Discussion timer fields - backend only tracks when 2+ users are present
			DiscussionStartTime: poi.DiscussionStartTime,
			IsDiscussionActive:  isDiscussionActive,
			
			CreatedAt:        poi.CreatedAt,
		}
	}
	
	// Return response
	response := GetPOIsResponse{
		MapID: mapID,
		POIs:  poiInfos,
		Count: len(poiInfos),
	}
	
	c.JSON(http.StatusOK, response)
}

// CreatePOI handles POST /api/pois
func (h *POIHandler) CreatePOI(c *gin.Context) {
	contentType := c.GetHeader("Content-Type")
	
	// Handle multipart form data (with optional image)
	if strings.Contains(contentType, "multipart/form-data") {
		h.createPOIFromForm(c)
		return
	}
	
	// Handle JSON request (existing functionality)
	var req CreatePOIRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Code:    "INVALID_REQUEST",
			Message: "Invalid request format",
			Details: err.Error(),
		})
		return
	}
	
	// Validate request
	if err := h.validateCreatePOIRequest(req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Code:    "VALIDATION_ERROR",
			Message: "Request validation failed",
			Details: err.Error(),
		})
		return
	}
	
	// Check rate limit
	if err := h.rateLimiter.CheckRateLimit(c, req.CreatedBy, services.ActionCreatePOI); err != nil {
		h.handleRateLimitError(c, err)
		return
	}
	
	// Set default max participants if not provided
	maxParticipants := req.MaxParticipants
	if maxParticipants <= 0 {
		maxParticipants = 10 // Default value
	}
	
	// Create POI
	poi, err := h.poiService.CreatePOI(c, req.MapID, req.Name, req.Description, req.Position, req.CreatedBy, maxParticipants)
	if err != nil {
		if isDuplicateLocationError(err) {
			c.JSON(http.StatusConflict, ErrorResponse{
				Code:    "DUPLICATE_LOCATION",
				Message: "A POI already exists at this location",
			})
			return
		}
		
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Code:    "INTERNAL_ERROR",
			Message: "Failed to create POI",
			Details: err.Error(),
		})
		return
	}
	
	// Add rate limit headers
	h.addRateLimitHeaders(c, req.CreatedBy, services.ActionCreatePOI)
	
	// Return response
	response := CreatePOIResponse{
		ID:              poi.ID,
		MapID:           poi.MapID,
		Name:            poi.Name,
		Description:     poi.Description,
		Position:        poi.Position,
		CreatedBy:       poi.CreatedBy,
		MaxParticipants: poi.MaxParticipants,
		ImageURL:        poi.ImageURL,
		ThumbnailURL:    poi.ThumbnailURL,
		CreatedAt:       poi.CreatedAt,
	}
	
	c.JSON(http.StatusCreated, response)
}

// createPOIFromForm handles POI creation from multipart form data
func (h *POIHandler) createPOIFromForm(c *gin.Context) {
	// Parse form data
	req, imageFile, err := h.parseCreatePOIForm(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Code:    "INVALID_REQUEST",
			Message: "Invalid form data",
			Details: err.Error(),
		})
		return
	}
	
	// Validate request
	if err := h.validateCreatePOIRequest(req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Code:    "VALIDATION_ERROR",
			Message: "Request validation failed",
			Details: err.Error(),
		})
		return
	}
	
	// Check rate limit
	if err := h.rateLimiter.CheckRateLimit(c, req.CreatedBy, services.ActionCreatePOI); err != nil {
		h.handleRateLimitError(c, err)
		return
	}
	
	// Set default max participants if not provided
	maxParticipants := req.MaxParticipants
	if maxParticipants <= 0 {
		maxParticipants = 10 // Default value
	}
	
	var poi *models.POI
	
	// Create POI with or without image
	if imageFile != nil {
		poi, err = h.poiService.CreatePOIWithImage(c, req.MapID, req.Name, req.Description, req.Position, req.CreatedBy, maxParticipants, imageFile)
	} else {
		poi, err = h.poiService.CreatePOI(c, req.MapID, req.Name, req.Description, req.Position, req.CreatedBy, maxParticipants)
	}
	
	if err != nil {
		if isDuplicateLocationError(err) {
			c.JSON(http.StatusConflict, ErrorResponse{
				Code:    "DUPLICATE_LOCATION",
				Message: "A POI already exists at this location",
			})
			return
		}
		
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Code:    "INTERNAL_ERROR",
			Message: "Failed to create POI",
			Details: err.Error(),
		})
		return
	}
	
	// Add rate limit headers
	h.addRateLimitHeaders(c, req.CreatedBy, services.ActionCreatePOI)
	
	// Return response
	response := CreatePOIResponse{
		ID:              poi.ID,
		MapID:           poi.MapID,
		Name:            poi.Name,
		Description:     poi.Description,
		Position:        poi.Position,
		CreatedBy:       poi.CreatedBy,
		MaxParticipants: poi.MaxParticipants,
		ImageURL:        poi.ImageURL,
		ThumbnailURL:    poi.ThumbnailURL,
		CreatedAt:       poi.CreatedAt,
	}
	
	c.JSON(http.StatusCreated, response)
}

// GetPOI handles GET /api/pois/:poiId
func (h *POIHandler) GetPOI(c *gin.Context) {
	poiID := c.Param("poiId")
	if poiID == "" {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Code:    "INVALID_REQUEST",
			Message: "POI ID is required",
		})
		return
	}
	
	// Get POI
	poi, err := h.poiService.GetPOI(c, poiID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusNotFound, ErrorResponse{
				Code:    "POI_NOT_FOUND",
				Message: "POI not found",
			})
			return
		}
		
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Code:    "INTERNAL_ERROR",
			Message: "Failed to get POI",
			Details: err.Error(),
		})
		return
	}
	
	// Return response
	response := GetPOIResponse{
		ID:              poi.ID,
		MapID:           poi.MapID,
		Name:            poi.Name,
		Description:     poi.Description,
		Position:        poi.Position,
		CreatedBy:       poi.CreatedBy,
		MaxParticipants: poi.MaxParticipants,
		ImageURL:        poi.ImageURL,
		ThumbnailURL:    poi.ThumbnailURL,
		CreatedAt:       poi.CreatedAt,
	}
	
	c.JSON(http.StatusOK, response)
}

// UpdatePOI handles PUT /api/pois/:poiId
func (h *POIHandler) UpdatePOI(c *gin.Context) {
	poiID := c.Param("poiId")
	if poiID == "" {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Code:    "INVALID_REQUEST",
			Message: "POI ID is required",
		})
		return
	}
	
	var req UpdatePOIRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Code:    "INVALID_REQUEST",
			Message: "Invalid request format",
			Details: err.Error(),
		})
		return
	}
	
	// Convert to service update data
	updateData := services.POIUpdateData{
		Name:            req.Name,
		Description:     req.Description,
		MaxParticipants: req.MaxParticipants,
	}
	
	// Update POI
	poi, err := h.poiService.UpdatePOI(c, poiID, updateData)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusNotFound, ErrorResponse{
				Code:    "POI_NOT_FOUND",
				Message: "POI not found",
			})
			return
		}
		
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Code:    "INTERNAL_ERROR",
			Message: "Failed to update POI",
			Details: err.Error(),
		})
		return
	}
	
	// Return response
	response := UpdatePOIResponse{
		ID:              poi.ID,
		MapID:           poi.MapID,
		Name:            poi.Name,
		Description:     poi.Description,
		Position:        poi.Position,
		CreatedBy:       poi.CreatedBy,
		MaxParticipants: poi.MaxParticipants,
		ImageURL:        poi.ImageURL,
		ThumbnailURL:    poi.ThumbnailURL,
		CreatedAt:       poi.CreatedAt,
	}
	
	c.JSON(http.StatusOK, response)
}

// DeletePOI handles DELETE /api/pois/:poiId
func (h *POIHandler) DeletePOI(c *gin.Context) {
	poiID := c.Param("poiId")
	if poiID == "" {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Code:    "INVALID_REQUEST",
			Message: "POI ID is required",
		})
		return
	}
	
	// Delete POI
	if err := h.poiService.DeletePOI(c, poiID); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusNotFound, ErrorResponse{
				Code:    "POI_NOT_FOUND",
				Message: "POI not found",
			})
			return
		}
		
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Code:    "INTERNAL_ERROR",
			Message: "Failed to delete POI",
			Details: err.Error(),
		})
		return
	}
	
	// Return success response
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "POI deleted successfully",
	})
}

// JoinPOI handles POST /api/pois/:poiId/join
func (h *POIHandler) JoinPOI(c *gin.Context) {
	poiID := c.Param("poiId")
	if poiID == "" {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Code:    "INVALID_REQUEST",
			Message: "POI ID is required",
		})
		return
	}
	
	var req JoinPOIRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Code:    "INVALID_REQUEST",
			Message: "Invalid request format",
			Details: err.Error(),
		})
		return
	}
	
	// Check rate limit
	if err := h.rateLimiter.CheckRateLimit(c, req.UserID, services.ActionJoinPOI); err != nil {
		h.handleRateLimitError(c, err)
		return
	}
	
	// Join POI
	if err := h.poiService.JoinPOI(c, poiID, req.UserID); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusNotFound, ErrorResponse{
				Code:    "POI_NOT_FOUND",
				Message: "POI not found",
			})
			return
		}
		
		if isCapacityExceededError(err) {
			c.JSON(http.StatusConflict, ErrorResponse{
				Code:    "CAPACITY_EXCEEDED",
				Message: "POI has reached maximum capacity",
			})
			return
		}
		
		if isAlreadyJoinedError(err) {
			c.JSON(http.StatusConflict, ErrorResponse{
				Code:    "ALREADY_JOINED",
				Message: "User has already joined this POI",
			})
			return
		}
		
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Code:    "INTERNAL_ERROR",
			Message: "Failed to join POI",
			Details: err.Error(),
		})
		return
	}
	
	// Add rate limit headers
	h.addRateLimitHeaders(c, req.UserID, services.ActionJoinPOI)
	
	// Return response
	response := JoinPOIResponse{
		Success: true,
		POIID:   poiID,
		UserID:  req.UserID,
	}
	
	c.JSON(http.StatusOK, response)
}

// LeavePOI handles POST /api/pois/:poiId/leave
func (h *POIHandler) LeavePOI(c *gin.Context) {
	poiID := c.Param("poiId")
	if poiID == "" {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Code:    "INVALID_REQUEST",
			Message: "POI ID is required",
		})
		return
	}
	
	var req LeavePOIRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Code:    "INVALID_REQUEST",
			Message: "Invalid request format",
			Details: err.Error(),
		})
		return
	}
	
	// Leave POI
	if err := h.poiService.LeavePOI(c, poiID, req.UserID); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusNotFound, ErrorResponse{
				Code:    "POI_NOT_FOUND",
				Message: "POI not found",
			})
			return
		}
		
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Code:    "INTERNAL_ERROR",
			Message: "Failed to leave POI",
			Details: err.Error(),
		})
		return
	}
	
	// Return response
	response := LeavePOIResponse{
		Success: true,
		POIID:   poiID,
		UserID:  req.UserID,
	}
	
	c.JSON(http.StatusOK, response)
}

// GetPOIParticipants handles GET /api/pois/:poiId/participants
func (h *POIHandler) GetPOIParticipants(c *gin.Context) {
	poiID := c.Param("poiId")
	if poiID == "" {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Code:    "INVALID_REQUEST",
			Message: "POI ID is required",
		})
		return
	}
	
	// Get participants
	participants, err := h.poiService.GetPOIParticipants(c, poiID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusNotFound, ErrorResponse{
				Code:    "POI_NOT_FOUND",
				Message: "POI not found",
			})
			return
		}
		
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Code:    "INTERNAL_ERROR",
			Message: "Failed to get POI participants",
			Details: err.Error(),
		})
		return
	}
	
	// Return response
	response := GetPOIParticipantsResponse{
		POIID:        poiID,
		Participants: participants,
		Count:        len(participants),
	}
	
	c.JSON(http.StatusOK, response)
}

// Helper methods

// parseCreatePOIForm parses multipart form data into CreatePOIRequest
func (h *POIHandler) parseCreatePOIForm(c *gin.Context) (CreatePOIRequest, *multipart.FileHeader, error) {
	var req CreatePOIRequest
	
	// Parse form data
	req.MapID = c.PostForm("mapId")
	req.Name = c.PostForm("name")
	req.Description = c.PostForm("description")
	req.CreatedBy = c.PostForm("createdBy")
	
	// Parse position
	latStr := c.PostForm("position.lat")
	lngStr := c.PostForm("position.lng")
	
	if latStr == "" || lngStr == "" {
		return req, nil, errors.New("position.lat and position.lng are required")
	}
	
	lat, err := strconv.ParseFloat(latStr, 64)
	if err != nil {
		return req, nil, errors.New("invalid position.lat: " + err.Error())
	}
	
	lng, err := strconv.ParseFloat(lngStr, 64)
	if err != nil {
		return req, nil, errors.New("invalid position.lng: " + err.Error())
	}
	
	req.Position = models.LatLng{Lat: lat, Lng: lng}
	
	// Parse max participants
	maxParticipantsStr := c.PostForm("maxParticipants")
	if maxParticipantsStr != "" {
		maxParticipants, err := strconv.Atoi(maxParticipantsStr)
		if err != nil {
			return req, nil, errors.New("invalid maxParticipants: " + err.Error())
		}
		req.MaxParticipants = maxParticipants
	}
	
	// Get image file if present
	imageFile, err := c.FormFile("image")
	if err != nil && err != http.ErrMissingFile {
		return req, nil, errors.New("invalid image file: " + err.Error())
	}
	
	return req, imageFile, nil
}

// validateCreatePOIRequest validates the create POI request
func (h *POIHandler) validateCreatePOIRequest(req CreatePOIRequest) error {
	if req.MapID == "" {
		return errors.New("map ID is required")
	}
	if req.Name == "" {
		return errors.New("name is required")
	}
	if req.CreatedBy == "" {
		return errors.New("createdBy is required")
	}
	if err := req.Position.Validate(); err != nil {
		return errors.New("invalid position: " + err.Error())
	}
	if req.MaxParticipants < 0 {
		return errors.New("maxParticipants cannot be negative")
	}
	return nil
}

// handleRateLimitError handles rate limit errors
func (h *POIHandler) handleRateLimitError(c *gin.Context, err error) {
	if rateLimitErr, ok := err.(*services.RateLimitError); ok {
		// Set Retry-After header
		c.Header("Retry-After", strconv.Itoa(int(rateLimitErr.RetryAfter.Seconds())))
		
		c.JSON(http.StatusTooManyRequests, ErrorResponse{
			Code:    "RATE_LIMIT_EXCEEDED",
			Message: "Rate limit exceeded",
			Details: err.Error(),
		})
		return
	}
	
	// Generic rate limit error
	c.JSON(http.StatusTooManyRequests, ErrorResponse{
		Code:    "RATE_LIMIT_EXCEEDED",
		Message: "Rate limit exceeded",
		Details: err.Error(),
	})
}

// addRateLimitHeaders adds rate limit headers to the response
func (h *POIHandler) addRateLimitHeaders(c *gin.Context, userID string, action services.ActionType) {
	headers, err := h.rateLimiter.GetRateLimitHeaders(c, userID, action)
	if err != nil {
		// Log error but don't fail the request
		return
	}
	
	for key, value := range headers {
		c.Header(key, value)
	}
}

// isDuplicateLocationError checks if the error indicates duplicate location
func isDuplicateLocationError(err error) bool {
	return err != nil && strings.Contains(err.Error(), "duplicate POI location")
}

// isCapacityExceededError checks if the error indicates capacity exceeded
func isCapacityExceededError(err error) bool {
	return err != nil && strings.Contains(err.Error(), "capacity exceeded")
}

// isAlreadyJoinedError checks if the error indicates user already joined
func isAlreadyJoinedError(err error) bool {
	return err != nil && strings.Contains(err.Error(), "already joined")
}

// ClearAllPOIs handles DELETE /api/pois/dev/clear-all - Development endpoint to clear all POIs
func (h *POIHandler) ClearAllPOIs(c *gin.Context) {
	// Get mapId from query parameter, default to "default-map"
	mapID := c.DefaultQuery("mapId", "default-map")
	
	// Clear all POIs for the map
	if err := h.poiService.ClearAllPOIs(c, mapID); err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Code:    "INTERNAL_ERROR",
			Message: "Failed to clear POIs",
			Details: err.Error(),
		})
		return
	}
	
	// Return success response
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "All POIs cleared successfully",
		"mapId":   mapID,
	})
}
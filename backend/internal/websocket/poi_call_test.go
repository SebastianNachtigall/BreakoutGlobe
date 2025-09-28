package websocket

import (
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"breakoutglobe/internal/models"

	"github.com/gin-gonic/gin"
	ws "github.com/gorilla/websocket"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
)

// POICallTestSuite contains tests for POI-based WebRTC signaling
type POICallTestSuite struct {
	suite.Suite
	mockSessionService *MockSessionService
	mockRateLimiter    *MockRateLimiter
	mockPOIService     *MockPOIService
	handler            *Handler
	server             *httptest.Server
	wsURL              string
}

func (suite *POICallTestSuite) SetupTest() {
	gin.SetMode(gin.TestMode)
	
	suite.mockSessionService = new(MockSessionService)
	suite.mockRateLimiter = new(MockRateLimiter)
	suite.mockPOIService = new(MockPOIService)
	
	suite.handler = NewHandler(suite.mockSessionService, suite.mockRateLimiter, nil, suite.mockPOIService)
	
	// Setup test server
	router := gin.New()
	router.GET("/ws", suite.handler.HandleWebSocket)
	
	suite.server = httptest.NewServer(router)
	suite.wsURL = "ws" + strings.TrimPrefix(suite.server.URL, "http") + "/ws"
}

func (suite *POICallTestSuite) TearDownTest() {
	suite.server.Close()
	suite.mockSessionService.AssertExpectations(suite.T())
	suite.mockRateLimiter.AssertExpectations(suite.T())
	suite.mockPOIService.AssertExpectations(suite.T())
}

func (suite *POICallTestSuite) TestPOICallOffer() {
	// Setup two sessions for the same POI
	session1 := &models.Session{
		ID:       "session-user1",
		UserID:   "user-1",
		MapID:    "map-123",
		IsActive: true,
		AvatarPos: models.LatLng{Lat: 40.7128, Lng: -74.0060},
	}
	session2 := &models.Session{
		ID:       "session-user2", 
		UserID:   "user-2",
		MapID:    "map-123",
		IsActive: true,
		AvatarPos: models.LatLng{Lat: 40.7129, Lng: -74.0061},
	}
	
	// Mock session validation for both users
	suite.mockSessionService.On("GetSession", mock.Anything, "session-user1").Return(session1, nil)
	suite.mockSessionService.On("GetSession", mock.Anything, "session-user2").Return(session2, nil)
	
	// Connect both clients
	conn1, _, err := ws.DefaultDialer.Dial(suite.wsURL+"?sessionId=session-user1", nil)
	suite.Require().NoError(err)
	defer conn1.Close()
	
	conn2, _, err := ws.DefaultDialer.Dial(suite.wsURL+"?sessionId=session-user2", nil)
	suite.Require().NoError(err)
	defer conn2.Close()
	
	// Read welcome messages
	var welcomeMsg1, welcomeMsg2 Message
	err = conn1.ReadJSON(&welcomeMsg1)
	suite.Require().NoError(err)
	assert.Equal(suite.T(), "welcome", welcomeMsg1.Type)
	
	err = conn2.ReadJSON(&welcomeMsg2)
	suite.Require().NoError(err)
	assert.Equal(suite.T(), "welcome", welcomeMsg2.Type)
	
	// Clear initial messages (user_joined, initial_users)
	suite.clearInitialMessages(conn1, conn2)
	
	// Send POI call offer from user1 to user2
	offerMsg := Message{
		Type: "poi_call_offer",
		Data: map[string]interface{}{
			"poiId":        "poi-123",
			"targetUserId": "user-2",
			"sdp": map[string]interface{}{
				"type": "offer",
				"sdp":  "v=0\r\no=- 123456789 2 IN IP4 127.0.0.1\r\n...",
			},
		},
		Timestamp: time.Now(),
	}
	
	err = conn1.WriteJSON(offerMsg)
	suite.Require().NoError(err)
	
	// User2 should receive the POI call offer
	var receivedMsg Message
	err = conn2.ReadJSON(&receivedMsg)
	suite.Require().NoError(err)
	
	assert.Equal(suite.T(), "poi_call_offer", receivedMsg.Type)
	
	data, ok := receivedMsg.Data.(map[string]interface{})
	suite.Require().True(ok)
	
	assert.Equal(suite.T(), "poi-123", data["poiId"])
	assert.Equal(suite.T(), "user-1", data["fromUserId"])
	
	sdp, ok := data["sdp"].(map[string]interface{})
	suite.Require().True(ok)
	assert.Equal(suite.T(), "offer", sdp["type"])
}

func (suite *POICallTestSuite) TestPOICallAnswer() {
	// Setup two sessions
	session1 := &models.Session{
		ID:       "session-user1",
		UserID:   "user-1", 
		MapID:    "map-123",
		IsActive: true,
		AvatarPos: models.LatLng{Lat: 40.7128, Lng: -74.0060},
	}
	session2 := &models.Session{
		ID:       "session-user2",
		UserID:   "user-2",
		MapID:    "map-123", 
		IsActive: true,
		AvatarPos: models.LatLng{Lat: 40.7129, Lng: -74.0061},
	}
	
	suite.mockSessionService.On("GetSession", mock.Anything, "session-user1").Return(session1, nil)
	suite.mockSessionService.On("GetSession", mock.Anything, "session-user2").Return(session2, nil)
	
	// Connect both clients
	conn1, _, err := ws.DefaultDialer.Dial(suite.wsURL+"?sessionId=session-user1", nil)
	suite.Require().NoError(err)
	defer conn1.Close()
	
	conn2, _, err := ws.DefaultDialer.Dial(suite.wsURL+"?sessionId=session-user2", nil)
	suite.Require().NoError(err)
	defer conn2.Close()
	
	// Clear initial messages
	suite.clearInitialMessages(conn1, conn2)
	
	// Send POI call answer from user2 to user1
	answerMsg := Message{
		Type: "poi_call_answer",
		Data: map[string]interface{}{
			"poiId":        "poi-123",
			"targetUserId": "user-1",
			"sdp": map[string]interface{}{
				"type": "answer",
				"sdp":  "v=0\r\no=- 987654321 2 IN IP4 127.0.0.1\r\n...",
			},
		},
		Timestamp: time.Now(),
	}
	
	err = conn2.WriteJSON(answerMsg)
	suite.Require().NoError(err)
	
	// User1 should receive the POI call answer
	var receivedMsg Message
	err = conn1.ReadJSON(&receivedMsg)
	suite.Require().NoError(err)
	
	assert.Equal(suite.T(), "poi_call_answer", receivedMsg.Type)
	
	data, ok := receivedMsg.Data.(map[string]interface{})
	suite.Require().True(ok)
	
	assert.Equal(suite.T(), "poi-123", data["poiId"])
	assert.Equal(suite.T(), "user-2", data["fromUserId"])
}

func (suite *POICallTestSuite) TestPOICallICECandidate() {
	// Setup two sessions
	session1 := &models.Session{
		ID:       "session-user1",
		UserID:   "user-1",
		MapID:    "map-123", 
		IsActive: true,
		AvatarPos: models.LatLng{Lat: 40.7128, Lng: -74.0060},
	}
	session2 := &models.Session{
		ID:       "session-user2",
		UserID:   "user-2",
		MapID:    "map-123",
		IsActive: true,
		AvatarPos: models.LatLng{Lat: 40.7129, Lng: -74.0061},
	}
	
	suite.mockSessionService.On("GetSession", mock.Anything, "session-user1").Return(session1, nil)
	suite.mockSessionService.On("GetSession", mock.Anything, "session-user2").Return(session2, nil)
	
	// Connect both clients
	conn1, _, err := ws.DefaultDialer.Dial(suite.wsURL+"?sessionId=session-user1", nil)
	suite.Require().NoError(err)
	defer conn1.Close()
	
	conn2, _, err := ws.DefaultDialer.Dial(suite.wsURL+"?sessionId=session-user2", nil)
	suite.Require().NoError(err)
	defer conn2.Close()
	
	// Clear initial messages
	suite.clearInitialMessages(conn1, conn2)
	
	// Send POI call ICE candidate from user1 to user2
	candidateMsg := Message{
		Type: "poi_call_ice_candidate",
		Data: map[string]interface{}{
			"poiId":        "poi-123",
			"targetUserId": "user-2",
			"candidate": map[string]interface{}{
				"candidate":     "candidate:2233684774 1 udp 2122129152 192.168.1.100 54400 typ host generation 0 ufrag WbMY network-id 1 network-cost 10",
				"sdpMid":        "0",
				"sdpMLineIndex": 0,
			},
		},
		Timestamp: time.Now(),
	}
	
	err = conn1.WriteJSON(candidateMsg)
	suite.Require().NoError(err)
	
	// User2 should receive the POI call ICE candidate
	var receivedMsg Message
	err = conn2.ReadJSON(&receivedMsg)
	suite.Require().NoError(err)
	
	assert.Equal(suite.T(), "poi_call_ice_candidate", receivedMsg.Type)
	
	data, ok := receivedMsg.Data.(map[string]interface{})
	suite.Require().True(ok)
	
	assert.Equal(suite.T(), "poi-123", data["poiId"])
	assert.Equal(suite.T(), "user-1", data["fromUserId"])
}

func (suite *POICallTestSuite) TestPOICallInvalidMessage() {
	// Setup session
	session := &models.Session{
		ID:       "session-user1",
		UserID:   "user-1",
		MapID:    "map-123",
		IsActive: true,
		AvatarPos: models.LatLng{Lat: 40.7128, Lng: -74.0060},
	}
	
	suite.mockSessionService.On("GetSession", mock.Anything, "session-user1").Return(session, nil)
	
	// Connect client
	conn, _, err := ws.DefaultDialer.Dial(suite.wsURL+"?sessionId=session-user1", nil)
	suite.Require().NoError(err)
	defer conn.Close()
	
	// Clear initial messages
	var msg Message
	conn.ReadJSON(&msg) // welcome
	conn.ReadJSON(&msg) // initial_users
	
	// Send invalid POI call offer (missing poiId)
	invalidMsg := Message{
		Type: "poi_call_offer",
		Data: map[string]interface{}{
			"targetUserId": "user-2",
			"sdp": map[string]interface{}{
				"type": "offer",
				"sdp":  "v=0\r\n...",
			},
		},
		Timestamp: time.Now(),
	}
	
	err = conn.WriteJSON(invalidMsg)
	suite.Require().NoError(err)
	
	// Should receive error message
	var errorMsg Message
	err = conn.ReadJSON(&errorMsg)
	suite.Require().NoError(err)
	
	assert.Equal(suite.T(), "error", errorMsg.Type)
	
	data, ok := errorMsg.Data.(map[string]interface{})
	suite.Require().True(ok)
	assert.Contains(suite.T(), data["message"], "poiId is required")
}

// Helper function to clear initial messages
func (suite *POICallTestSuite) clearInitialMessages(conns ...*ws.Conn) {
	for _, conn := range conns {
		// Read welcome message
		var msg Message
		err := conn.ReadJSON(&msg)
		suite.Require().NoError(err)
		suite.Require().Equal("welcome", msg.Type)
		
		// Read initial_users message
		err = conn.ReadJSON(&msg)
		suite.Require().NoError(err)
		suite.Require().Equal("initial_users", msg.Type)
		
		// Read any user_joined messages from other connections
		conn.SetReadDeadline(time.Now().Add(200 * time.Millisecond))
		for {
			err := conn.ReadJSON(&msg)
			if err != nil {
				break // Timeout reached, no more messages
			}
			// Just consume user_joined messages
		}
		conn.SetReadDeadline(time.Time{}) // Clear deadline
	}
}

func TestPOICallTestSuite(t *testing.T) {
	suite.Run(t, new(POICallTestSuite))
}
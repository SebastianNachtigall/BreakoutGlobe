package websocket

import (
	"log/slog"
	"sync"
)

// Manager manages WebSocket client connections
type Manager struct {
	clients    map[string]*Client // sessionID -> Client
	mapClients map[string]map[string]*Client // mapID -> sessionID -> Client
	register   chan *Client
	unregister chan *Client
	broadcast  chan BroadcastMessage
	mutex      sync.RWMutex
	logger     *slog.Logger
}

// BroadcastMessage represents a message to be broadcasted
type BroadcastMessage struct {
	MapID     string
	Message   Message
	ExceptID  string // Optional: exclude this session ID from broadcast
}

// NewManager creates a new WebSocket manager
func NewManager() *Manager {
	manager := &Manager{
		clients:    make(map[string]*Client),
		mapClients: make(map[string]map[string]*Client),
		register:   make(chan *Client),
		unregister: make(chan *Client),
		broadcast:  make(chan BroadcastMessage, 100), // Buffer for 100 messages
		logger:     slog.Default(),
	}
	
	go manager.run()
	return manager
}

// run starts the manager's main loop
func (m *Manager) run() {
	for {
		select {
		case client := <-m.register:
			m.registerClient(client)
			
		case client := <-m.unregister:
			m.unregisterClient(client)
			
		case broadcastMsg := <-m.broadcast:
			m.broadcastToMap(broadcastMsg)
		}
	}
}

// RegisterClient registers a new client connection
func (m *Manager) RegisterClient(client *Client) {
	m.register <- client
}

// UnregisterClient unregisters a client connection
func (m *Manager) UnregisterClient(client *Client) {
	m.unregister <- client
}

// BroadcastToMap broadcasts a message to all clients in a specific map
func (m *Manager) BroadcastToMap(mapID string, message Message) error {
	broadcastMsg := BroadcastMessage{
		MapID:   mapID,
		Message: message,
	}
	
	select {
	case m.broadcast <- broadcastMsg:
		return nil
	default:
		m.logger.Warn("Broadcast channel full, dropping message", 
			"mapId", mapID, 
			"messageType", message.Type)
		return nil
	}
}

// BroadcastToMapExcept broadcasts a message to all clients in a map except one
func (m *Manager) BroadcastToMapExcept(mapID, exceptSessionID string, message Message) error {
	broadcastMsg := BroadcastMessage{
		MapID:    mapID,
		Message:  message,
		ExceptID: exceptSessionID,
	}
	
	select {
	case m.broadcast <- broadcastMsg:
		return nil
	default:
		m.logger.Warn("Broadcast channel full, dropping message", 
			"mapId", mapID, 
			"messageType", message.Type)
		return nil
	}
}

// BroadcastToAll broadcasts a message to all connected clients
func (m *Manager) BroadcastToAll(message Message) error {
	m.mutex.RLock()
	defer m.mutex.RUnlock()
	
	for _, client := range m.clients {
		select {
		case client.Send <- message:
		default:
			// Client's send channel is full, close it
			m.logger.Warn("Client send channel full, closing connection", 
				"sessionId", client.SessionID)
			
			// Close channel safely
			select {
			case <-client.Send:
				// Channel is already closed
			default:
				close(client.Send)
			}
			
			delete(m.clients, client.SessionID)
			
			// Remove from map clients
			if mapClients, exists := m.mapClients[client.MapID]; exists {
				delete(mapClients, client.SessionID)
				if len(mapClients) == 0 {
					delete(m.mapClients, client.MapID)
				}
			}
		}
	}
	
	return nil
}

// GetConnectedClients returns the number of connected clients
func (m *Manager) GetConnectedClients() int {
	m.mutex.RLock()
	defer m.mutex.RUnlock()
	return len(m.clients)
}

// GetMapClients returns the number of clients connected to a specific map
func (m *Manager) GetMapClients(mapID string) int {
	m.mutex.RLock()
	defer m.mutex.RUnlock()
	
	if mapClients, exists := m.mapClients[mapID]; exists {
		return len(mapClients)
	}
	return 0
}

// IsClientConnected checks if a client is connected
func (m *Manager) IsClientConnected(sessionID string) bool {
	m.mutex.RLock()
	defer m.mutex.RUnlock()
	
	_, exists := m.clients[sessionID]
	return exists
}

// GetClientMaps returns all map IDs that have connected clients
func (m *Manager) GetClientMaps() []string {
	m.mutex.RLock()
	defer m.mutex.RUnlock()
	
	maps := make([]string, 0, len(m.mapClients))
	for mapID := range m.mapClients {
		maps = append(maps, mapID)
	}
	return maps
}

// GetMapClientSessions returns all session IDs for clients in a specific map
func (m *Manager) GetMapClientSessions(mapID string) []string {
	m.mutex.RLock()
	defer m.mutex.RUnlock()
	
	if mapClients, exists := m.mapClients[mapID]; exists {
		sessions := make([]string, 0, len(mapClients))
		for sessionID := range mapClients {
			sessions = append(sessions, sessionID)
		}
		return sessions
	}
	return []string{}
}

// registerClient handles client registration
func (m *Manager) registerClient(client *Client) {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	
	// Add to clients map
	m.clients[client.SessionID] = client
	
	// Add to map clients
	if m.mapClients[client.MapID] == nil {
		m.mapClients[client.MapID] = make(map[string]*Client)
	}
	m.mapClients[client.MapID][client.SessionID] = client
	
	// Log all clients in this map for debugging
	var mapClientSessions []string
	for sessionID := range m.mapClients[client.MapID] {
		mapClientSessions = append(mapClientSessions, sessionID)
	}
	
	m.logger.Info("✅ Client registered", 
		"sessionId", client.SessionID, 
		"userId", client.UserID, 
		"mapId", client.MapID,
		"totalClients", len(m.clients),
		"mapClients", len(m.mapClients[client.MapID]),
		"mapClientSessions", mapClientSessions)
}

// unregisterClient handles client unregistration
func (m *Manager) unregisterClient(client *Client) {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	
	// Remove from clients map
	if _, exists := m.clients[client.SessionID]; exists {
		delete(m.clients, client.SessionID)
		close(client.Send)
	}
	
	// Remove from map clients
	if mapClients, exists := m.mapClients[client.MapID]; exists {
		delete(mapClients, client.SessionID)
		if len(mapClients) == 0 {
			delete(m.mapClients, client.MapID)
		}
	}
	
	m.logger.Info("Client unregistered", 
		"sessionId", client.SessionID, 
		"userId", client.UserID, 
		"mapId", client.MapID,
		"totalClients", len(m.clients))
}

// broadcastToMap handles broadcasting messages to a specific map
func (m *Manager) broadcastToMap(broadcastMsg BroadcastMessage) {
	m.mutex.RLock()
	defer m.mutex.RUnlock()
	
	mapClients, exists := m.mapClients[broadcastMsg.MapID]
	if !exists {
		m.logger.Warn("🚫 No clients found for map during broadcast", 
			"mapId", broadcastMsg.MapID, 
			"messageType", broadcastMsg.Message.Type)
		return
	}
	
	totalClients := len(mapClients)
	eligibleClients := 0
	sentCount := 0
	failedCount := 0
	
	// Count eligible clients (excluding sender)
	for sessionID := range mapClients {
		if broadcastMsg.ExceptID == "" || sessionID != broadcastMsg.ExceptID {
			eligibleClients++
		}
	}
	
	m.logger.Info("📡 Starting broadcast to map", 
		"mapId", broadcastMsg.MapID, 
		"messageType", broadcastMsg.Message.Type,
		"totalClients", totalClients,
		"eligibleClients", eligibleClients,
		"exceptId", broadcastMsg.ExceptID)
	
	for sessionID, client := range mapClients {
		// Skip the excluded session if specified
		if broadcastMsg.ExceptID != "" && sessionID == broadcastMsg.ExceptID {
			m.logger.Debug("⏭️ Skipping sender client", 
				"sessionId", sessionID,
				"mapId", broadcastMsg.MapID)
			continue
		}
		
		m.logger.Debug("📤 Attempting to send message to client", 
			"sessionId", sessionID,
			"userId", client.UserID,
			"mapId", broadcastMsg.MapID,
			"messageType", broadcastMsg.Message.Type)
		
		select {
		case client.Send <- broadcastMsg.Message:
			sentCount++
			m.logger.Debug("✅ Message sent successfully to client", 
				"sessionId", sessionID,
				"userId", client.UserID,
				"messageType", broadcastMsg.Message.Type)
		default:
			failedCount++
			// Client's send channel is full, close it
			m.logger.Warn("❌ Client send channel full during broadcast, closing connection", 
				"sessionId", client.SessionID,
				"userId", client.UserID,
				"mapId", broadcastMsg.MapID,
				"messageType", broadcastMsg.Message.Type)
			
			// Close channel safely
			select {
			case <-client.Send:
				// Channel is already closed
			default:
				close(client.Send)
			}
			
			delete(m.clients, client.SessionID)
			delete(mapClients, client.SessionID)
		}
	}
	
	m.logger.Info("📊 Broadcast completed", 
		"mapId", broadcastMsg.MapID, 
		"messageType", broadcastMsg.Message.Type,
		"totalClients", totalClients,
		"eligibleClients", eligibleClients,
		"sentCount", sentCount,
		"failedCount", failedCount)
}

// Shutdown gracefully shuts down the manager
func (m *Manager) Shutdown() {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	
	// Close all client connections
	for _, client := range m.clients {
		// Close send channel if not already closed
		select {
		case <-client.Send:
			// Channel is already closed
		default:
			close(client.Send)
		}
		
		// Close websocket connection if it exists
		if client.Conn != nil {
			client.Conn.Close()
		}
	}
	
	// Clear maps
	m.clients = make(map[string]*Client)
	m.mapClients = make(map[string]map[string]*Client)
	
	m.logger.Info("WebSocket manager shutdown complete")
}

// BroadcastToUser sends a message to a specific user by their user ID
func (m *Manager) BroadcastToUser(userID string, message Message, exceptSessionID string) {
	m.mutex.RLock()
	defer m.mutex.RUnlock()
	
	var targetClient *Client
	
	// Find the client with the matching user ID
	for sessionID, client := range m.clients {
		if client.UserID == userID && sessionID != exceptSessionID {
			targetClient = client
			break
		}
	}
	
	if targetClient == nil {
		m.logger.Warn("🚫 Target user not found for message", 
			"targetUserId", userID, 
			"messageType", message.Type)
		return
	}
	
	// Send message to target user
	select {
	case targetClient.Send <- message:
		m.logger.Info("📨 Message sent to user", 
			"targetUserId", userID,
			"targetSessionId", targetClient.SessionID,
			"messageType", message.Type)
	default:
		m.logger.Warn("📨 Failed to send message to user (channel full)", 
			"targetUserId", userID,
			"targetSessionId", targetClient.SessionID,
			"messageType", message.Type)
	}
}
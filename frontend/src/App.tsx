import { useEffect, useState, useCallback, useRef, useMemo } from 'react'
import { MapContainer, AvatarData, POIData } from './components/MapContainer'
import { ConnectionStatus } from './components/ConnectionStatus'
import { NotificationCenter } from './components/NotificationCenter'
import { ErrorBoundary } from './components/ErrorBoundary'
import { POICreationModal } from './components/POICreationModal'
import { POIDetailsPanel } from './components/POIDetailsPanel'
import ProfileCreationModal from './components/ProfileCreationModal'
import ProfileMenu from './components/ProfileMenu'
import { sessionStore } from './stores/sessionStore'
import { poiStore } from './stores/poiStore'
import { errorStore } from './stores/errorStore'
import { avatarStore } from './stores/avatarStore'
import { WebSocketClient, ConnectionStatus as WSConnectionStatus } from './services/websocket-client'
import { getCurrentUserProfile } from './services/api'
import { userProfileStore } from './stores/userProfileStore'
import type { UserProfile } from './types/models'

// Mock data for development
const mockSession = {
  id: 'session-123',
  avatarId: 'avatar-456',
  position: { lat: 40.7128, lng: -74.0060 }
}

function App() {
  // Store subscriptions
  const sessionState = sessionStore()
  const poiState = poiStore()
  // const errorState = errorStore() // Not used in current implementation
  const avatarState = avatarStore()
  
  // Force re-render when avatar store changes by subscribing to the entire store
  const [avatarStoreVersion, setAvatarStoreVersion] = useState(0)
  
  useEffect(() => {
    const unsubscribe = avatarStore.subscribe(() => {
      setAvatarStoreVersion(prev => prev + 1)
    })
    return unsubscribe
  }, [])
  
  // Local component state
  const [wsClient, setWsClient] = useState<WebSocketClient | null>(null)
  const [connectionStatus, setConnectionStatus] = useState<WSConnectionStatus>(WSConnectionStatus.DISCONNECTED)
  const [isInitialized, setIsInitialized] = useState(false)
  const [selectedPOI, setSelectedPOI] = useState<POIData | null>(null)
  const [showPOICreation, setShowPOICreation] = useState(false)
  const [poiCreationPosition, setPOICreationPosition] = useState<{ lat: number; lng: number } | null>(null)
  
  // Profile system state
  const [userProfile, setUserProfile] = useState<UserProfile | null>(null)
  const [showProfileCreation, setShowProfileCreation] = useState(false)
  const [profileCheckComplete, setProfileCheckComplete] = useState(false)
  
  const initializationRef = useRef(false)

  // Initialize session and WebSocket connection
  useEffect(() => {
    if (initializationRef.current) return
    initializationRef.current = true

    const initializeApp = async () => {
      try {
        // Check if user has a profile first - try localStorage first, then backend
        let profile = userProfileStore.getState().getProfileOffline()
        
        if (profile) {
          console.info('✅ User profile loaded from localStorage:', profile.displayName)
          setUserProfile(profile)
          setProfileCheckComplete(true)
          
          // Try to sync with backend in the background (don't block UI)
          try {
            const backendProfile = await getCurrentUserProfile(profile.id)
            if (backendProfile && backendProfile.id === profile.id) {
              // Update local profile with any backend changes
              userProfileStore.getState().setProfile(backendProfile)
              setUserProfile(backendProfile)
              console.info('🔄 Profile synced with backend')
            }
          } catch (syncError) {
            console.info('ℹ️ Backend sync failed, using cached profile')
          }
        } else {
          // No cached profile, try backend
          try {
            const backendProfile = await getCurrentUserProfile()
            if (backendProfile) {
              console.info('✅ User profile found on backend:', backendProfile.displayName)
              userProfileStore.getState().setProfile(backendProfile)
              setUserProfile(backendProfile)
              setProfileCheckComplete(true)
            } else {
              // No profile exists anywhere, show profile creation modal
              console.info('ℹ️ No user profile found - showing profile creation modal')
              setShowProfileCreation(true)
              setProfileCheckComplete(true)
              return // Don't continue initialization until profile is created
            }
          } catch (error) {
            // Handle 404 as expected behavior for new users
            if (error instanceof Error && error.message.includes('404')) {
              console.info('ℹ️ New user detected - showing profile creation modal')
            } else {
              console.info('ℹ️ No existing profile found - showing profile creation modal')
            }
            setShowProfileCreation(true)
            setProfileCheckComplete(true)
            return // Don't continue initialization until profile is created
          }
        }

        // Create or restore session
        let sessionId = sessionState.sessionId
        
        if (!sessionId) {
          // Create new session via API
          const response = await fetch('http://localhost:8080/api/sessions', {
            method: 'POST',
            headers: {
              'Content-Type': 'application/json',
            },
            body: JSON.stringify({
              userId: userProfile?.id || `user-${Date.now()}`, // Use profile ID if available
              mapId: 'default-map', // Use a default map ID for now
              avatarPosition: mockSession.position
            }),
          })
          
          if (!response.ok) {
            throw new Error('Failed to create session')
          }
          
          const sessionData = await response.json()
          sessionId = sessionData.sessionId || sessionData.id
          
          // Update session store
          sessionStore.getState().createSession(sessionId!, sessionData.position || mockSession.position)
        }

        // Initialize WebSocket connection
        const wsUrl = `ws://localhost:8080/ws?sessionId=${sessionId}`
        const client = new WebSocketClient(wsUrl, sessionId!)
        
        // Set up WebSocket event handlers
        client.onStatusChange((status) => {
          setConnectionStatus(status)
        })
        
        client.onError((error) => {
          errorStore.getState().addError({
            id: Date.now().toString(),
            message: error.message,
            type: 'websocket',
            severity: 'error',
            timestamp: error.timestamp
          })
        })
        
        // Set up multi-user avatar event handlers
        client.onStateSync((data) => {
          if (data.type === 'avatar') {
            // Handle avatar-related state sync events
            console.log('Avatar state sync:', data);
          }
        })
        
        // Connect WebSocket
        await client.connect()
        setWsClient(client)
        
        // Request initial users after connection
        client.requestInitialUsers()
        
        // Load initial POIs
        await loadPOIs()
        
        setIsInitialized(true)
        
      } catch (error) {
        console.error('Failed to initialize app:', error)
        errorStore.getState().addError({
          id: Date.now().toString(),
          message: error instanceof Error ? error.message : 'Failed to initialize application',
          type: 'api',
          severity: 'error',
          timestamp: new Date()
        })
      }
    }

    initializeApp()
    
    // Cleanup on unmount
    return () => {
      if (wsClient) {
        wsClient.disconnect()
      }
    }
  }, [])

  // Load POIs from API
  const loadPOIs = async () => {
    try {
      poiStore.getState().setLoading(true)
      
      const response = await fetch('http://localhost:8080/api/pois?mapId=default-map')
      if (!response.ok) {
        throw new Error('Failed to load POIs')
      }
      
      const data = await response.json()
      poiStore.getState().setPOIs(data.pois || [])
      
    } catch (error) {
      console.error('Failed to load POIs:', error)
      poiStore.getState().setError(error instanceof Error ? error.message : 'Failed to load POIs')
    } finally {
      poiStore.getState().setLoading(false)
    }
  }

  // Handle avatar movement
  const handleAvatarMove = useCallback((position: { lat: number; lng: number }) => {
    if (!wsClient || !wsClient.isConnected()) {
      errorStore.getState().addError({
        id: Date.now().toString(),
        message: 'Cannot move avatar: not connected to server',
        type: 'network',
        severity: 'warning',
        timestamp: new Date()
      })
      return
    }

    // Use WebSocket client for optimistic updates
    wsClient.moveAvatar(position)
  }, [wsClient])

  // Handle map click for avatar movement and auto-leave
  const handleMapClick = useCallback((event: { lngLat: { lng: number; lat: number } }) => {
    // Move avatar to clicked location
    handleAvatarMove({ lat: event.lngLat.lat, lng: event.lngLat.lng })
    
    // Auto-leave current POI and close details panel
    if (wsClient) {
      wsClient.leaveCurrentPOI()
    }
    setSelectedPOI(null)
  }, [handleAvatarMove, wsClient])

  // Handle POI creation from context menu
  const handlePOICreate = useCallback((position: { lat: number; lng: number }) => {
    setPOICreationPosition(position)
    setShowPOICreation(true)
  }, [])

  // Handle POI creation submission
  const handleCreatePOISubmit = useCallback(async (poiData: Omit<POIData, 'id' | 'participantCount'>) => {
    if (!wsClient || !poiCreationPosition) return

    const newPOI: POIData = {
      ...poiData,
      id: `temp-${Date.now()}`, // Temporary ID for optimistic update
      position: poiCreationPosition,
      participantCount: 0,
      createdBy: sessionState.sessionId || 'anonymous',
      createdAt: new Date()
    }

    try {
      // Use WebSocket client for optimistic updates
      wsClient.createPOI(newPOI)
      
      setShowPOICreation(false)
      setPOICreationPosition(null)
      
    } catch (error) {
      console.error('Failed to create POI:', error)
      errorStore.getState().addError({
        id: Date.now().toString(),
        message: error instanceof Error ? error.message : 'Failed to create POI',
        type: 'api',
        severity: 'error',
        timestamp: new Date()
      })
    }
  }, [wsClient, poiCreationPosition])

  // Handle POI selection
  const handlePOIClick = useCallback((poiId: string) => {
    const poi = poiState.pois.find(p => p.id === poiId)
    if (poi) {
      setSelectedPOI(poi)
    }
  }, [poiState.pois])

  // Update selectedPOI when POI data changes in the store
  useEffect(() => {
    if (selectedPOI) {
      const updatedPOI = poiState.pois.find(p => p.id === selectedPOI.id)
      if (updatedPOI) {
        setSelectedPOI(updatedPOI)
      }
    }
  }, [poiState.pois, selectedPOI])

  // Handle POI join/leave with auto-leave functionality
  const handleJoinPOI = useCallback(async (poiId: string) => {
    if (!wsClient) return
    
    // Find the POI to get its position
    const poi = poiState.pois.find(p => p.id === poiId)
    if (poi) {
      // Move avatar to POI location (slightly offset to avoid overlap with marker)
      const offsetPosition = {
        lat: poi.position.lat + 0.0001, // Small offset north
        lng: poi.position.lng + 0.0001  // Small offset east
      }
      handleAvatarMove(offsetPosition)
    }
    
    // Join the POI via HTTP API (since WebSocket doesn't work with mock backend)
    try {
      const response = await fetch(`http://localhost:8080/api/pois/${poiId}/join`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ sessionId: sessionState.sessionId })
      })
      
      if (response.ok) {
        // Update local state
        poiState.joinPOIWithAutoLeave(poiId, sessionState.sessionId || '')
        
        // Refresh POI data to get updated participant list
        await loadPOIs()
      }
    } catch (error) {
      console.error('Failed to join POI:', error)
    }
  }, [wsClient, poiState, handleAvatarMove, sessionState.sessionId])

  const handleLeavePOI = useCallback((poiId: string) => {
    if (!wsClient) return
    wsClient.leavePOI(poiId)
  }, [wsClient])

  // Handle profile creation
  const handleProfileCreated = useCallback((profile: UserProfile) => {
    console.info('🎉 Profile created successfully:', profile.displayName)
    
    // Save to localStorage and update state
    userProfileStore.getState().setProfile(profile)
    setUserProfile(profile)
    setShowProfileCreation(false)
    
    // Now initialize the app with the new profile
    const initializeWithProfile = async () => {
      try {
        // Create new session via API
        const response = await fetch('http://localhost:8080/api/sessions', {
          method: 'POST',
          headers: {
            'Content-Type': 'application/json',
          },
          body: JSON.stringify({
            userId: profile.id,
            mapId: 'default-map',
            avatarPosition: mockSession.position
          }),
        })
        
        if (!response.ok) {
          throw new Error('Failed to create session')
        }
        
        const sessionData = await response.json()
        const sessionId = sessionData.sessionId || sessionData.id
        
        // Update session store
        sessionStore.getState().createSession(sessionId, sessionData.position || mockSession.position)

        // Initialize WebSocket connection
        const wsUrl = `ws://localhost:8080/ws?sessionId=${sessionId}`
        const client = new WebSocketClient(wsUrl, sessionId)
        
        // Set up WebSocket event handlers
        client.onStatusChange((status) => {
          setConnectionStatus(status)
        })
        
        client.onError((error) => {
          errorStore.getState().addError({
            id: Date.now().toString(),
            message: error.message,
            type: 'websocket',
            severity: 'error',
            timestamp: error.timestamp
          })
        })
        
        // Set up multi-user avatar event handlers
        client.onStateSync((data) => {
          if (data.type === 'avatar') {
            // Handle avatar-related state sync events
            console.log('Avatar state sync:', data);
          }
        })
        
        // Connect WebSocket
        await client.connect()
        setWsClient(client)
        
        // Request initial users after connection
        client.requestInitialUsers()
        
        // Load initial POIs
        await loadPOIs()
        
        setIsInitialized(true)
        
      } catch (error) {
        console.error('Failed to initialize app after profile creation:', error)
        errorStore.getState().addError({
          id: Date.now().toString(),
          message: error instanceof Error ? error.message : 'Failed to initialize application',
          type: 'api',
          severity: 'error',
          timestamp: new Date()
        })
      }
    }

    initializeWithProfile()
  }, [])

  const handleProfileCreationClose = useCallback(() => {
    // For now, we require a profile to use the app
    // In a real app, you might want to allow anonymous usage
    setShowProfileCreation(false)
  }, [])

  // Convert session state to avatar data for MapContainer
  // CRITICAL: Memoize avatars array to prevent unnecessary re-renders and marker recreation
  const avatars: AvatarData[] = useMemo(() => {
    const currentUserAvatar: AvatarData = {
      sessionId: sessionState.sessionId || 'current-user',
      userId: userProfile?.id,
      displayName: userProfile?.displayName,
      avatarURL: userProfile?.avatarURL,
      position: sessionState.avatarPosition,
      isCurrentUser: true,
      isMoving: sessionState.isMoving,
      role: userProfile?.role
    };
    
    // Get other users' avatars from avatarStore
    const otherUsersAvatars = avatarState.getAvatarsForCurrentMap();
    
    return [currentUserAvatar, ...otherUsersAvatars];
  }, [
    sessionState.sessionId,
    userProfile?.id,
    userProfile?.displayName,
    userProfile?.avatarURL,
    sessionState.avatarPosition,
    sessionState.isMoving,
    userProfile?.role,
    avatarStoreVersion // Use version to trigger re-renders when avatar store changes
  ])



  // Show loading screen while checking for profile
  if (!profileCheckComplete) {
    return (
      <div className="h-screen w-screen flex items-center justify-center bg-gray-100">
        <div className="text-center">
          <div className="animate-spin rounded-full h-12 w-12 border-b-2 border-blue-600 mx-auto mb-4"></div>
          <p className="text-gray-600">Loading BreakoutGlobe...</p>
        </div>
      </div>
    )
  }

  // Show profile creation modal if no profile exists
  if (showProfileCreation) {
    return (
      <ErrorBoundary>
        <div className="h-screen w-screen flex items-center justify-center bg-gray-100">
          <div className="text-center mb-8">
            <h1 className="text-4xl font-bold text-gray-900 mb-2">Welcome to BreakoutGlobe</h1>
            <p className="text-gray-600 mb-8">Create your profile to get started</p>
          </div>
          <ProfileCreationModal
            isOpen={true}
            onProfileCreated={handleProfileCreated}
            onClose={handleProfileCreationClose}
          />
        </div>
      </ErrorBoundary>
    )
  }

  // Show loading screen while initializing the app
  if (!isInitialized) {
    return (
      <div className="h-screen w-screen flex items-center justify-center bg-gray-100">
        <div className="text-center">
          <div className="animate-spin rounded-full h-12 w-12 border-b-2 border-blue-600 mx-auto mb-4"></div>
          <p className="text-gray-600">Initializing BreakoutGlobe...</p>
          {userProfile && (
            <p className="text-sm text-gray-500 mt-2">Welcome back, {userProfile.displayName}!</p>
          )}
        </div>
      </div>
    )
  }

  return (
    <ErrorBoundary>
      <div className="h-screen w-screen flex flex-col">
        {/* Header */}
        <div className="bg-blue-600 text-white p-4 shadow-lg">
          <div className="flex justify-between items-center">
            <div>
              <h1 className="text-2xl font-bold">BreakoutGlobe</h1>
              <p className="text-blue-100">Interactive Workshop Platform</p>
            </div>
            <div className="flex items-center space-x-4">
              <ConnectionStatus 
                status={connectionStatus}
                sessionId={sessionState.sessionId}
              />
              {userProfile && (
                <ProfileMenu userProfile={userProfile} />
              )}
            </div>
          </div>
        </div>

        {/* Map Container */}
        <div className="flex-1 relative">
          <MapContainer
            initialCenter={[10.0, 54.0]}
            initialZoom={4}
            avatars={avatars}
            pois={poiState.pois || []}
            onMapClick={handleMapClick}
            onAvatarMove={handleAvatarMove}
            onPOIClick={handlePOIClick}
            onPOICreate={handlePOICreate}
          />
          
          {/* POI Details Panel */}
          {selectedPOI && (
            <POIDetailsPanel
              poi={selectedPOI}
              currentUserId={sessionState.sessionId || ''}
              isUserParticipant={poiState.currentUserPOI === selectedPOI.id}
              onJoin={() => handleJoinPOI(selectedPOI.id)}
              onLeave={() => handleLeavePOI(selectedPOI.id)}
              onClose={() => setSelectedPOI(null)}
            />
          )}
        </div>

        {/* Status Bar */}
        <div className="bg-gray-800 text-white p-2 text-sm">
          <div className="flex justify-between items-center">
            <span>Connected Users: {avatars.length}</span>
            <span>
              {connectionStatus === WSConnectionStatus.CONNECTED 
                ? 'Click to move • Right-click to create POI'
                : 'Connecting...'
              }
            </span>
          </div>
        </div>

        {/* Modals */}
        {showPOICreation && poiCreationPosition && (
          <POICreationModal
            position={poiCreationPosition}
            onSubmit={handleCreatePOISubmit}
            onCancel={() => {
              setShowPOICreation(false)
              setPOICreationPosition(null)
            }}
          />
        )}

        {/* Notifications */}
        <NotificationCenter />
      </div>
    </ErrorBoundary>
  )
}

export default App
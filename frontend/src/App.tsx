import { useEffect, useState, useCallback, useRef, useMemo } from 'react'
import { MapContainer, AvatarData, POIData } from './components/MapContainer'
import { ConnectionStatus } from './components/ConnectionStatus'
import { NotificationCenter } from './components/NotificationCenter'
import { ErrorBoundary } from './components/ErrorBoundary'
import { POICreationModal } from './components/POICreationModal'
import { POIDetailsPanel } from './components/POIDetailsPanel'
import { VideoCallModal } from './components/VideoCallModal'
import { AvatarTooltip } from './components/AvatarTooltip'
import ProfileCreationModal from './components/ProfileCreationModal'
import ProfileMenu from './components/ProfileMenu'
import { sessionStore } from './stores/sessionStore'
import { poiStore } from './stores/poiStore'
import { errorStore } from './stores/errorStore'
import { avatarStore } from './stores/avatarStore'
import { videoCallStore, setWebSocketClient } from './stores/videoCallStore'
import { WebSocketClient, ConnectionStatus as WSConnectionStatus } from './services/websocket-client'
import { getCurrentUserProfile, createPOI, transformToCreatePOIRequest, transformFromPOIResponse, joinPOI, leavePOI, getPOIs } from './services/api'
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
  const videoCallState = videoCallStore()

  // Discussion timer is now handled directly in POIDetailsPanel component

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

  // Avatar tooltip state
  const [avatarTooltip, setAvatarTooltip] = useState<{
    isOpen: boolean;
    position: { x: number; y: number };
    avatar: AvatarData | null;
  }>({
    isOpen: false,
    position: { x: 0, y: 0 },
    avatar: null
  })

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
        const wsUrl = `ws://localhost:8080/ws?sessionId=${sessionId}`;

        // Initialize WebSocket connection
        const client = new WebSocketClient(wsUrl, sessionId!);

        // Make WebSocket client globally accessible for WebRTC signaling
        (window as any).wsClient = client;

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

        // Connect WebSocket (non-blocking for POC)
        try {
          await client.connect()
          setWsClient(client)

          // Set WebSocket client for video call store
          setWebSocketClient(client)

          // Request initial users after connection
          client.requestInitialUsers()
          console.log('✅ WebSocket connected successfully')
        } catch (wsError) {
          console.warn('⚠️ WebSocket connection failed, continuing without real-time features:', wsError)
          // Continue without WebSocket for POC testing
        }

        // Load initial POIs
        try {
          await loadPOIs()
        } catch (poiError) {
          console.warn('⚠️ Failed to load POIs, continuing with empty POI list:', poiError)
        }

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
      poiStore.getState().setError(null)

      const apiPOIs = await getPOIs('default-map')
      console.log('📦 Loaded POIs from API:', apiPOIs.length)
      console.log('🔍 API POIs with images:', apiPOIs.filter(poi => poi.imageUrl).map(poi => ({ name: poi.name, imageUrl: poi.imageUrl })))

      // Transform API responses to frontend format
      const transformedPOIs = apiPOIs.map(transformFromPOIResponse)
      console.log('🔄 Transformed POIs with images:', transformedPOIs.filter(poi => poi.imageUrl).map(poi => ({ name: poi.name, imageUrl: poi.imageUrl })))
      poiStore.getState().setPOIs(transformedPOIs)

    } catch (error) {
      console.error('❌ Failed to load POIs:', error)
      
      const errorMessage = error instanceof Error ? error.message : 'Failed to load POIs'
      poiStore.getState().setError(errorMessage)

      // Show error notification with retry option
      const isNetworkError = error instanceof Error && (
        error.message.includes('fetch') || 
        error.message.includes('network') || 
        error.message.includes('Failed to fetch')
      );

      errorStore.getState().addError({
        id: Date.now().toString(),
        message: isNetworkError 
          ? 'Network error occurred while loading POIs. Please check your connection.'
          : errorMessage,
        type: isNetworkError ? 'network' : 'api',
        severity: 'error',
        timestamp: new Date(),
        retryable: true,
        retryAction: () => loadPOIs(),
        autoRemoveAfter: 10000
      })
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
  const handleCreatePOISubmit = useCallback(async (poiData: {
    name: string;
    description: string;
    maxParticipants: number;
    position: { lat: number; lng: number };
    image?: File;
  }) => {
    if (!poiCreationPosition || !userProfile) return

    // Create optimistic POI for immediate UI feedback (declare outside try block)
    const optimisticPOI: POIData = {
      id: `temp-${Date.now()}`,
      name: poiData.name,
      description: poiData.description,
      position: poiData.position,
      participantCount: 0,
      maxParticipants: poiData.maxParticipants,
      createdBy: userProfile.id,
      createdAt: new Date()
    }

    try {
      // Set loading state
      poiStore.getState().setLoading(true)

      // Transform form data to API request format
      const apiRequest = transformToCreatePOIRequest(
        poiData,
        userProfile.id,
        'default-map'
      )

      console.log('🚀 Creating POI with data:', apiRequest)

      // Add optimistic POI to store
      poiStore.getState().createPOIOptimistic(optimisticPOI)

      // Call API to create POI
      const apiResponse = await createPOI(apiRequest)
      console.log('✅ POI created successfully:', apiResponse)

      // Transform API response to frontend format
      const createdPOI = transformFromPOIResponse(apiResponse)

      // Replace optimistic POI with real POI
      poiStore.getState().removePOI(optimisticPOI.id)
      poiStore.getState().addPOI(createdPOI)

      // Close modal
      setShowPOICreation(false)
      setPOICreationPosition(null)

    } catch (error) {
      console.error('❌ Failed to create POI:', error)
      
      // Remove optimistic POI on failure
      poiStore.getState().rollbackPOICreation(optimisticPOI.id)

      // Show error to user with retry option for network failures
      const isNetworkError = error instanceof Error && (
        error.message.includes('fetch') || 
        error.message.includes('network') || 
        error.message.includes('Failed to fetch')
      );

      errorStore.getState().addError({
        id: Date.now().toString(),
        message: isNetworkError 
          ? 'Network error occurred while creating POI. Please check your connection and try again.'
          : error instanceof Error ? error.message : 'Failed to create POI',
        type: isNetworkError ? 'network' : 'api',
        severity: 'error',
        timestamp: new Date(),
        retryable: isNetworkError,
        retryAction: isNetworkError ? () => {
          // Retry the POI creation
          handleCreatePOISubmit(poiData);
        } : undefined,
        autoRemoveAfter: 10000 // Auto-remove after 10 seconds
      })
    } finally {
      // Clear loading state
      poiStore.getState().setLoading(false)
    }
  }, [poiCreationPosition, userProfile])

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
    if (!userProfile) return

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

    try {
      // Optimistic update - join POI immediately for UI responsiveness
      const success = poiState.joinPOIOptimisticWithAutoLeave(poiId, userProfile.id)
      if (!success) {
        // POI is full or doesn't exist
        errorStore.getState().addError({
          id: Date.now().toString(),
          message: 'Cannot join POI - it may be full or no longer exist',
          type: 'validation',
          severity: 'warning',
          timestamp: new Date(),
          autoRemoveAfter: 5000
        })
        return
      }

      // Call API to join POI
      await joinPOI(poiId, userProfile.id)
      console.log('✅ Successfully joined POI:', poiId)

      // Confirm the optimistic update
      poiState.confirmJoinPOI(poiId, userProfile.id)

      // Refresh POI data to get updated participant list
      await loadPOIs()

    } catch (error) {
      console.error('❌ Failed to join POI:', error)
      
      // Rollback optimistic update
      poiState.rollbackJoinPOI(poiId, userProfile.id)

      // Show error with retry option
      const isNetworkError = error instanceof Error && (
        error.message.includes('fetch') || 
        error.message.includes('network') || 
        error.message.includes('Failed to fetch')
      );

      errorStore.getState().addError({
        id: Date.now().toString(),
        message: isNetworkError 
          ? 'Network error occurred while joining POI. Please try again.'
          : error instanceof Error ? error.message : 'Failed to join POI',
        type: isNetworkError ? 'network' : 'api',
        severity: 'error',
        timestamp: new Date(),
        retryable: isNetworkError,
        retryAction: isNetworkError ? () => handleJoinPOI(poiId) : undefined,
        autoRemoveAfter: 8000
      })
    }
  }, [userProfile, poiState, handleAvatarMove])

  const handleLeavePOI = useCallback(async (poiId: string) => {
    if (!userProfile) return

    try {
      // Optimistic update - leave POI immediately for UI responsiveness
      const success = poiState.leavePOI(poiId, userProfile.id)
      if (!success) {
        console.warn('User was not in POI:', poiId)
        return
      }

      // Call API to leave POI
      await leavePOI(poiId, userProfile.id)
      console.log('✅ Successfully left POI:', poiId)

      // Refresh POI data to get updated participant list
      await loadPOIs()

    } catch (error) {
      console.error('❌ Failed to leave POI:', error)
      
      // Rollback optimistic update by rejoining
      poiState.joinPOI(poiId, userProfile.id)

      // Show error
      errorStore.getState().addError({
        id: Date.now().toString(),
        message: error instanceof Error ? error.message : 'Failed to leave POI',
        type: 'api',
        severity: 'error',
        timestamp: new Date(),
        autoRemoveAfter: 5000
      })
    }
  }, [userProfile, poiState])



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
        const wsUrl = `ws://localhost:8080/ws?sessionId=${sessionId}`;
        const client = new WebSocketClient(wsUrl, sessionId);

        // Make WebSocket client globally accessible for WebRTC signaling
        (window as any).wsClient = client;

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

        // Set WebSocket client for video call store
        setWebSocketClient(client)

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

  // Video call handlers
  const handleAcceptCall = useCallback(() => {
    videoCallStore.getState().acceptCall();
  }, [])

  const handleRejectCall = useCallback(() => {
    videoCallStore.getState().rejectCall();
  }, [])

  const handleEndCall = useCallback(() => {
    videoCallStore.getState().endCall();
  }, [])

  const handleToggleAudio = useCallback(() => {
    videoCallStore.getState().toggleAudio();
  }, [])

  const handleToggleVideo = useCallback(() => {
    videoCallStore.getState().toggleVideo();
  }, [])

  const handleCloseVideoCall = useCallback(() => {
    if (videoCallState.callState === 'connected' || videoCallState.callState === 'calling') {
      videoCallStore.getState().endCall();
    } else {
      videoCallStore.getState().clearCall();
    }
  }, [videoCallState.callState])

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

    // Add mock avatars for POC testing (when no real users are connected)
    const mockAvatars: AvatarData[] = otherUsersAvatars.length === 0 ? [
      {
        sessionId: 'mock-session-1',
        userId: 'mock-user-1',
        displayName: 'Alice Johnson',
        avatarURL: undefined, // Will show initials
        position: { lat: 40.7589, lng: -73.9851 }, // Near NYC
        isCurrentUser: false,
        role: 'user'
      },
      {
        sessionId: 'mock-session-2',
        userId: 'mock-user-2',
        displayName: 'Bob Smith',
        avatarURL: undefined, // Will show initials
        position: { lat: 51.5074, lng: -0.1278 }, // London
        isCurrentUser: false,
        role: 'user'
      },
      {
        sessionId: 'mock-session-3',
        userId: 'mock-user-3',
        displayName: 'Carol Davis',
        avatarURL: undefined, // Will show initials
        position: { lat: 48.8566, lng: 2.3522 }, // Paris
        isCurrentUser: false,
        role: 'admin'
      }
    ] : [];

    return [currentUserAvatar, ...otherUsersAvatars, ...mockAvatars];
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

  // Handle avatar click to show tooltip (defined after avatars array)
  const handleAvatarClick = useCallback((userId: string, clickPosition: { x: number; y: number }) => {
    // Don't allow clicking on yourself
    if (userId === userProfile?.id) {
      console.log('Cannot interact with your own avatar');
      return;
    }

    // Find the avatar data for the clicked user from the complete avatars array
    const targetAvatar = avatars.find(avatar => avatar.userId === userId);
    if (!targetAvatar) {
      console.warn('Target avatar not found for user:', userId);
      return;
    }

    console.log('👤 Avatar clicked, showing tooltip for:', targetAvatar.displayName);

    // Show avatar tooltip
    setAvatarTooltip({
      isOpen: true,
      position: clickPosition,
      avatar: targetAvatar
    });
  }, [userProfile?.id, avatars])

  // Handle starting a video call from the tooltip
  const handleStartCall = useCallback(() => {
    if (!avatarTooltip.avatar) return;

    console.log('📞 Starting call to:', avatarTooltip.avatar.displayName);

    // Close tooltip
    setAvatarTooltip({ isOpen: false, position: { x: 0, y: 0 }, avatar: null });

    // Initiate video call
    videoCallStore.getState().initiateCall(
      avatarTooltip.avatar.userId || avatarTooltip.avatar.sessionId,
      avatarTooltip.avatar.displayName || avatarTooltip.avatar.sessionId,
      avatarTooltip.avatar.avatarURL
    );
  }, [avatarTooltip.avatar])

  // Handle closing the avatar tooltip
  const handleCloseTooltip = useCallback(() => {
    setAvatarTooltip({ isOpen: false, position: { x: 0, y: 0 }, avatar: null });
  }, [])

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
            onAvatarClick={handleAvatarClick}
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
            <div className="flex items-center space-x-4">
              {/* Test buttons for video call POC */}
              <button
                onClick={() => {
                  videoCallStore.getState().receiveCall(
                    'test-call-123',
                    'test-user-456',
                    'Test User',
                    undefined
                  );
                }}
                className="bg-blue-600 hover:bg-blue-700 px-3 py-1 rounded text-xs"
              >
                Test Incoming Call
              </button>
              <span>
                {connectionStatus === WSConnectionStatus.CONNECTED
                  ? 'Click avatar for video call • Right-click to create POI'
                  : 'Click avatar for video call (WebSocket connecting...)'
                }
              </span>
            </div>
          </div>
        </div>

        {/* Modals */}
        {showPOICreation && poiCreationPosition && (
          <POICreationModal
            isOpen={true}
            position={poiCreationPosition}
            onCreate={handleCreatePOISubmit}
            onCancel={() => {
              setShowPOICreation(false)
              setPOICreationPosition(null)
            }}
            isLoading={poiState.isLoading}
          />
        )}

        {/* Video Call Modal */}
        {videoCallState.callState !== 'idle' && videoCallState.currentCall && (
          <VideoCallModal
            isOpen={true}
            onClose={handleCloseVideoCall}
            callState={videoCallState.callState}
            targetUser={{
              id: videoCallState.currentCall.targetUserId,
              displayName: videoCallState.currentCall.targetUserName,
              avatarURL: videoCallState.currentCall.targetUserAvatar
            }}
            localStream={videoCallState.localStream}
            remoteStream={videoCallState.remoteStream}
            isAudioEnabled={videoCallState.isAudioEnabled}
            isVideoEnabled={videoCallState.isVideoEnabled}
            onAcceptCall={handleAcceptCall}
            onRejectCall={handleRejectCall}
            onEndCall={handleEndCall}
            onToggleAudio={handleToggleAudio}
            onToggleVideo={handleToggleVideo}
          />
        )}

        {/* Avatar Tooltip */}
        {avatarTooltip.avatar && (() => {
          // Get fresh avatar data from store to include updated call status
          const freshAvatar = avatarStore.getState().getAvatarBySessionId(avatarTooltip.avatar.sessionId);
          const avatarData = freshAvatar || avatarTooltip.avatar;
          
          return (
            <AvatarTooltip
              isOpen={avatarTooltip.isOpen}
              position={avatarTooltip.position}
              avatar={avatarData}
              onClose={handleCloseTooltip}
              onStartCall={handleStartCall}
            />
          );
        })()}

        {/* Notifications */}
        <NotificationCenter />
      </div>
    </ErrorBoundary>
  )
}

export default App
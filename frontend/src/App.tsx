import React, { useState } from 'react'
import { MapContainer, AvatarData } from './components/MapContainer'

function App() {
  const [avatars, setAvatars] = useState<AvatarData[]>([
    {
      sessionId: 'current-user',
      position: { lat: 40.7128, lng: -74.0060 }, // New York
      isCurrentUser: true,
      isMoving: false
    },
    {
      sessionId: 'other-user-1',
      position: { lat: 51.5074, lng: -0.1278 }, // London
      isCurrentUser: false,
      isMoving: false
    },
    {
      sessionId: 'other-user-2',
      position: { lat: 35.6762, lng: 139.6503 }, // Tokyo
      isCurrentUser: false,
      isMoving: false
    }
  ])

  const handleAvatarMove = (position: { lat: number; lng: number }) => {
    console.log('Avatar move requested:', position)
    
    // Update current user's position
    setAvatars(prev => prev.map(avatar => 
      avatar.isCurrentUser 
        ? { ...avatar, position, isMoving: true }
        : avatar
    ))

    // Reset moving state after animation
    setTimeout(() => {
      setAvatars(prev => prev.map(avatar => 
        avatar.isCurrentUser 
          ? { ...avatar, isMoving: false }
          : avatar
      ))
    }, 500)
  }

  const handleMapClick = (event: { lngLat: { lng: number; lat: number } }) => {
    console.log('Map clicked:', event.lngLat)
  }

  return (
    <div className="h-screen w-screen flex flex-col">
      {/* Header */}
      <div className="bg-blue-600 text-white p-4 shadow-lg">
        <h1 className="text-2xl font-bold">BreakoutGlobe</h1>
        <p className="text-blue-100">Interactive Workshop Platform</p>
      </div>

      {/* Map Container */}
      <div className="flex-1 relative">
        <MapContainer
          initialCenter={[0, 20]} // Center on Atlantic
          initialZoom={2}
          avatars={avatars}
          onMapClick={handleMapClick}
          onAvatarMove={handleAvatarMove}
        />
      </div>

      {/* Status Bar */}
      <div className="bg-gray-800 text-white p-2 text-sm">
        <div className="flex justify-between items-center">
          <span>Connected Users: {avatars.length}</span>
          <span>Click anywhere on the map to move your avatar</span>
        </div>
      </div>
    </div>
  )
}

export default App
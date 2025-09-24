package models

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestSession_Validate(t *testing.T) {
	tests := []struct {
		name    string
		session Session
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid session",
			session: Session{
				ID:         "session-123",
				UserID:     "user-456",
				MapID:      "map-789",
				AvatarPos:  LatLng{Lat: 40.7128, Lng: -74.0060},
				CreatedAt:  time.Now(),
				LastActive: time.Now(),
				IsActive:   true,
			},
			wantErr: false,
		},
		{
			name: "empty session ID",
			session: Session{
				ID:         "",
				UserID:     "user-456",
				MapID:      "map-789",
				AvatarPos:  LatLng{Lat: 40.7128, Lng: -74.0060},
				CreatedAt:  time.Now(),
				LastActive: time.Now(),
				IsActive:   true,
			},
			wantErr: true,
			errMsg:  "session ID is required",
		},
		{
			name: "empty user ID",
			session: Session{
				ID:         "session-123",
				UserID:     "",
				MapID:      "map-789",
				AvatarPos:  LatLng{Lat: 40.7128, Lng: -74.0060},
				CreatedAt:  time.Now(),
				LastActive: time.Now(),
				IsActive:   true,
			},
			wantErr: true,
			errMsg:  "user ID is required",
		},
		{
			name: "empty map ID",
			session: Session{
				ID:         "session-123",
				UserID:     "user-456",
				MapID:      "",
				AvatarPos:  LatLng{Lat: 40.7128, Lng: -74.0060},
				CreatedAt:  time.Now(),
				LastActive: time.Now(),
				IsActive:   true,
			},
			wantErr: true,
			errMsg:  "map ID is required",
		},
		{
			name: "invalid avatar position",
			session: Session{
				ID:         "session-123",
				UserID:     "user-456",
				MapID:      "map-789",
				AvatarPos:  LatLng{Lat: 91.0, Lng: -74.0060}, // Invalid latitude
				CreatedAt:  time.Now(),
				LastActive: time.Now(),
				IsActive:   true,
			},
			wantErr: true,
			errMsg:  "latitude must be between -90 and 90",
		},
		{
			name: "zero time created at",
			session: Session{
				ID:         "session-123",
				UserID:     "user-456",
				MapID:      "map-789",
				AvatarPos:  LatLng{Lat: 40.7128, Lng: -74.0060},
				CreatedAt:  time.Time{},
				LastActive: time.Now(),
				IsActive:   true,
			},
			wantErr: true,
			errMsg:  "created at is required",
		},
		{
			name: "zero time last active",
			session: Session{
				ID:         "session-123",
				UserID:     "user-456",
				MapID:      "map-789",
				AvatarPos:  LatLng{Lat: 40.7128, Lng: -74.0060},
				CreatedAt:  time.Now(),
				LastActive: time.Time{},
				IsActive:   true,
			},
			wantErr: true,
			errMsg:  "last active is required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.session.Validate()

			if tt.wantErr {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errMsg)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestSession_IsExpired(t *testing.T) {
	tests := []struct {
		name       string
		lastActive time.Time
		timeout    time.Duration
		want       bool
	}{
		{
			name:       "not expired - recent activity",
			lastActive: time.Now().Add(-5 * time.Minute),
			timeout:    30 * time.Minute,
			want:       false,
		},
		{
			name:       "expired - old activity",
			lastActive: time.Now().Add(-45 * time.Minute),
			timeout:    30 * time.Minute,
			want:       true,
		},
		{
			name:       "exactly at timeout boundary",
			lastActive: time.Now().Add(-30 * time.Minute),
			timeout:    30 * time.Minute,
			want:       true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			session := Session{
				LastActive: tt.lastActive,
			}

			result := session.IsExpired(tt.timeout)
			assert.Equal(t, tt.want, result)
		})
	}
}

func TestSession_UpdateActivity(t *testing.T) {
	session := Session{
		LastActive: time.Now().Add(-10 * time.Minute),
		IsActive:   false,
	}

	oldLastActive := session.LastActive

	session.UpdateActivity()

	assert.True(t, session.LastActive.After(oldLastActive))
	assert.True(t, session.IsActive)
	assert.WithinDuration(t, time.Now(), session.LastActive, time.Second)
}

func TestSession_UpdateAvatarPosition(t *testing.T) {
	session := Session{
		AvatarPos:  LatLng{Lat: 0.0, Lng: 0.0},
		LastActive: time.Now().Add(-10 * time.Minute),
	}

	newPos := LatLng{Lat: 40.7128, Lng: -74.0060}
	oldLastActive := session.LastActive

	err := session.UpdateAvatarPosition(newPos)

	assert.NoError(t, err)
	assert.Equal(t, newPos, session.AvatarPos)
	assert.True(t, session.LastActive.After(oldLastActive))
}

func TestNewSession(t *testing.T) {
	userID := "user-123"
	mapID := "map-456"
	initialPos := LatLng{Lat: 40.7128, Lng: -74.0060}

	session, err := NewSession(userID, mapID, initialPos)

	assert.NoError(t, err)
	assert.NotEmpty(t, session.ID)
	assert.Equal(t, userID, session.UserID)
	assert.Equal(t, mapID, session.MapID)
	assert.Equal(t, initialPos, session.AvatarPos)
	assert.True(t, session.IsActive)
	assert.WithinDuration(t, time.Now(), session.CreatedAt, time.Second)
	assert.WithinDuration(t, time.Now(), session.LastActive, time.Second)
}

func TestNewSession_InvalidInput(t *testing.T) {
	tests := []struct {
		name       string
		userID     string
		mapID      string
		initialPos LatLng
		wantErr    bool
		errMsg     string
	}{
		{
			name:       "empty user ID",
			userID:     "",
			mapID:      "map-456",
			initialPos: LatLng{Lat: 40.7128, Lng: -74.0060},
			wantErr:    true,
			errMsg:     "user ID is required",
		},
		{
			name:       "empty map ID",
			userID:     "user-123",
			mapID:      "",
			initialPos: LatLng{Lat: 40.7128, Lng: -74.0060},
			wantErr:    true,
			errMsg:     "map ID is required",
		},
		{
			name:       "invalid position",
			userID:     "user-123",
			mapID:      "map-456",
			initialPos: LatLng{Lat: 91.0, Lng: -74.0060},
			wantErr:    true,
			errMsg:     "invalid initial position",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			session, err := NewSession(tt.userID, tt.mapID, tt.initialPos)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, session)
				if err != nil {
					assert.Contains(t, err.Error(), tt.errMsg)
				}
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, session)
			}
		})
	}
}

func TestSession_UpdateAvatarPosition_InvalidCoordinates(t *testing.T) {
	session := Session{
		AvatarPos: LatLng{Lat: 0.0, Lng: 0.0},
	}

	invalidPos := LatLng{Lat: 91.0, Lng: -74.0060} // Invalid latitude

	err := session.UpdateAvatarPosition(invalidPos)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "latitude must be between -90 and 90")
	// Position should not be updated on error
	assert.Equal(t, LatLng{Lat: 0.0, Lng: 0.0}, session.AvatarPos)
}

// Tests for new relationship functionality
func TestSession_BelongsToUser(t *testing.T) {
	userID := "user-123"
	session := Session{
		UserID: userID,
	}

	tests := []struct {
		name       string
		testUserID string
		expected   bool
	}{
		{
			name:       "session belongs to user",
			testUserID: userID,
			expected:   true,
		},
		{
			name:       "session does not belong to other user",
			testUserID: "other-user-456",
			expected:   false,
		},
		{
			name:       "empty user ID does not match",
			testUserID: "",
			expected:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := session.BelongsToUser(tt.testUserID)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestSession_BelongsToMap(t *testing.T) {
	mapID := "map-123"
	session := Session{
		MapID: mapID,
	}

	tests := []struct {
		name      string
		testMapID string
		expected  bool
	}{
		{
			name:      "session belongs to map",
			testMapID: mapID,
			expected:  true,
		},
		{
			name:      "session does not belong to other map",
			testMapID: "other-map-456",
			expected:  false,
		},
		{
			name:      "empty map ID does not match",
			testMapID: "",
			expected:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := session.BelongsToMap(tt.testMapID)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestSession_CanBeAccessedBy(t *testing.T) {
	userID := "user-123"
	session := Session{
		UserID: userID,
	}

	tests := []struct {
		name     string
		user     *User
		expected bool
	}{
		{
			name:     "nil user cannot access",
			user:     nil,
			expected: false,
		},
		{
			name: "session owner can access",
			user: &User{
				ID:   userID,
				Role: UserRoleUser,
			},
			expected: true,
		},
		{
			name: "other regular user cannot access",
			user: &User{
				ID:   "other-user-456",
				Role: UserRoleUser,
			},
			expected: false,
		},
		{
			name: "admin can access any session",
			user: &User{
				ID:   "admin-789",
				Role: UserRoleAdmin,
			},
			expected: true,
		},
		{
			name: "superadmin can access any session",
			user: &User{
				ID:   "superadmin-999",
				Role: UserRoleSuperAdmin,
			},
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := session.CanBeAccessedBy(tt.user)
			assert.Equal(t, tt.expected, result)
		})
	}
}
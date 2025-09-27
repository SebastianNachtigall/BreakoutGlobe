import { avatarStore } from '../avatarStore';
import type { AvatarData } from '../../components/MapContainer';

describe('Avatar Store POI Hiding', () => {
  beforeEach(() => {
    // Reset store before each test
    avatarStore.getState().clearAllAvatars();
  });

  it('should hide avatars when users join POIs', () => {
    const avatar1: AvatarData = {
      sessionId: 'session-1',
      userId: 'user-1',
      displayName: 'Alice',
      position: { lat: 40.7128, lng: -74.0060 },
      isCurrentUser: false
    };

    const avatar2: AvatarData = {
      sessionId: 'session-2',
      userId: 'user-2',
      displayName: 'Bob',
      position: { lat: 40.7129, lng: -74.0061 },
      isCurrentUser: false
    };

    // Add avatars to store
    avatarStore.getState().addOrUpdateAvatar(avatar1);
    avatarStore.getState().addOrUpdateAvatar(avatar2);

    // Initially both avatars should be visible
    let visibleAvatars = avatarStore.getState().getOtherUsersAvatars();
    expect(visibleAvatars).toHaveLength(2);

    // Hide avatar1 when user joins POI
    avatarStore.getState().hideAvatarForPOI('user-1', 'poi-123');

    // Now only avatar2 should be visible
    visibleAvatars = avatarStore.getState().getOtherUsersAvatars();
    expect(visibleAvatars).toHaveLength(1);
    expect(visibleAvatars[0].userId).toBe('user-2');

    // Hide avatar2 as well
    avatarStore.getState().hideAvatarForPOI('user-2', 'poi-123');

    // Now no avatars should be visible
    visibleAvatars = avatarStore.getState().getOtherUsersAvatars();
    expect(visibleAvatars).toHaveLength(0);
  });

  it('should show avatars when users leave POIs', () => {
    const avatar: AvatarData = {
      sessionId: 'session-1',
      userId: 'user-1',
      displayName: 'Alice',
      position: { lat: 40.7128, lng: -74.0060 },
      isCurrentUser: false
    };

    // Add avatar and hide it
    avatarStore.getState().addOrUpdateAvatar(avatar);
    avatarStore.getState().hideAvatarForPOI('user-1', 'poi-123');

    // Avatar should be hidden
    let visibleAvatars = avatarStore.getState().getOtherUsersAvatars();
    expect(visibleAvatars).toHaveLength(0);

    // Show avatar when user leaves POI
    avatarStore.getState().showAvatarForPOI('user-1', 'poi-123');

    // Avatar should be visible again
    visibleAvatars = avatarStore.getState().getOtherUsersAvatars();
    expect(visibleAvatars).toHaveLength(1);
    expect(visibleAvatars[0].userId).toBe('user-1');
  });

  it('should filter hidden avatars from getAvatarsForCurrentMap', () => {
    const avatar: AvatarData = {
      sessionId: 'session-1',
      userId: 'user-1',
      displayName: 'Alice',
      position: { lat: 40.7128, lng: -74.0060 },
      isCurrentUser: false
    };

    // Add avatar
    avatarStore.getState().addOrUpdateAvatar(avatar);
    avatarStore.getState().setCurrentMap('map-1');

    // Initially avatar should be visible
    let mapAvatars = avatarStore.getState().getAvatarsForCurrentMap();
    expect(mapAvatars).toHaveLength(1);

    // Hide avatar
    avatarStore.getState().hideAvatarForPOI('user-1', 'poi-123');

    // Avatar should be hidden from map view
    mapAvatars = avatarStore.getState().getAvatarsForCurrentMap();
    expect(mapAvatars).toHaveLength(0);
  });

  it('should clear hidden avatars when clearing all avatars', () => {
    const avatar: AvatarData = {
      sessionId: 'session-1',
      userId: 'user-1',
      displayName: 'Alice',
      position: { lat: 40.7128, lng: -74.0060 },
      isCurrentUser: false
    };

    // Add and hide avatar
    avatarStore.getState().addOrUpdateAvatar(avatar);
    avatarStore.getState().hideAvatarForPOI('user-1', 'poi-123');

    // Clear all avatars
    avatarStore.getState().clearAllAvatars();

    // Add avatar again - should be visible (hidden state cleared)
    avatarStore.getState().addOrUpdateAvatar(avatar);
    const visibleAvatars = avatarStore.getState().getOtherUsersAvatars();
    expect(visibleAvatars).toHaveLength(1);
  });
});
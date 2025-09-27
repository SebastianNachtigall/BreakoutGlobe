import React from 'react';
import { render, screen, act } from '@testing-library/react';
import { vi, beforeEach, afterEach, describe, it, expect } from 'vitest';
import { POIDetailsPanel } from '../POIDetailsPanel';
import type { POIData } from '../MapContainer';

// Mock timer functions
beforeEach(() => {
  vi.useFakeTimers();
});

afterEach(() => {
  vi.useRealTimers();
});

const mockPOI: POIData = {
  id: 'poi-1',
  name: 'Test POI',
  description: 'Test description',
  position: { lat: 40.7128, lng: -74.0060 },
  participantCount: 2,
  maxParticipants: 10,
  participants: [
    { id: 'user-1', name: 'John Doe' },
    { id: 'user-2', name: 'Jane Smith' }
  ],
  createdBy: 'user-1',
  createdAt: new Date(),
  // Discussion timer fields - backend only tracks when 2+ users are present
  discussionStartTime: new Date(Date.now() - 30000), // 30 seconds ago
  isDiscussionActive: true
};

describe('POIDetailsPanel - Discussion Timer', () => {
  it('should calculate and display discussion duration from start time', () => {
    render(
      <POIDetailsPanel
        poi={mockPOI}
        currentUserId="user-1"
        isUserParticipant={true}
        onJoin={vi.fn()}
        onLeave={vi.fn()}
        onClose={vi.fn()}
      />
    );

    // Should show discussion active with calculated duration (30 seconds)
    expect(screen.getByText(/Discussion active for: 30 seconds/)).toBeInTheDocument();
  });

  it('should update discussion timer every second', () => {
    render(
      <POIDetailsPanel
        poi={mockPOI}
        currentUserId="user-1"
        isUserParticipant={true}
        onJoin={vi.fn()}
        onLeave={vi.fn()}
        onClose={vi.fn()}
      />
    );

    // Initially shows 30 seconds
    expect(screen.getByText(/Discussion active for: 30 seconds/)).toBeInTheDocument();

    // Advance time by 1 second
    act(() => {
      vi.advanceTimersByTime(1000);
    });

    // Should now show 31 seconds
    expect(screen.getByText(/Discussion active for: 31 seconds/)).toBeInTheDocument();

    // Advance time by another 29 seconds (total 60 seconds = 1 minute)
    act(() => {
      vi.advanceTimersByTime(29000);
    });

    // Should now show 1 minute
    expect(screen.getByText(/Discussion active for: 1 minute/)).toBeInTheDocument();
  });

  it('should show "No active discussion" when discussion is not active', () => {
    const inactivePOI = {
      ...mockPOI,
      isDiscussionActive: false,
      discussionStartTime: null
    };

    render(
      <POIDetailsPanel
        poi={inactivePOI}
        currentUserId="user-1"
        isUserParticipant={true}
        onJoin={vi.fn()}
        onLeave={vi.fn()}
        onClose={vi.fn()}
      />
    );

    expect(screen.getByText('No active discussion')).toBeInTheDocument();
  });

  it('should handle missing discussionStartTime gracefully', () => {
    const poiWithoutStartTime = {
      ...mockPOI,
      isDiscussionActive: true,
      discussionStartTime: null
    };

    render(
      <POIDetailsPanel
        poi={poiWithoutStartTime}
        currentUserId="user-1"
        isUserParticipant={true}
        onJoin={vi.fn()}
        onLeave={vi.fn()}
        onClose={vi.fn()}
      />
    );

    // Should show "No active discussion" when start time is missing
    expect(screen.getByText('No active discussion')).toBeInTheDocument();
  });

  it('should clean up timer on unmount', () => {
    const clearIntervalSpy = vi.spyOn(global, 'clearInterval');
    
    const { unmount } = render(
      <POIDetailsPanel
        poi={mockPOI}
        currentUserId="user-1"
        isUserParticipant={true}
        onJoin={vi.fn()}
        onLeave={vi.fn()}
        onClose={vi.fn()}
      />
    );

    unmount();

    // Should have called clearInterval to clean up the timer
    expect(clearIntervalSpy).toHaveBeenCalled();
  });
});
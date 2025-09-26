import React, { useState, useEffect } from 'react';
import { userProfileStore } from '../stores/userProfileStore';
import { updateUserProfile } from '../services/api';

interface ProfileSettingsModalProps {
  isOpen: boolean;
  onClose: () => void;
}

const ProfileSettingsModal: React.FC<ProfileSettingsModalProps> = ({ isOpen, onClose }) => {
  const { profile, setProfile } = userProfileStore();
  const [displayName, setDisplayName] = useState('');
  const [aboutMe, setAboutMe] = useState('');
  const [isLoading, setIsLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [hasChanges, setHasChanges] = useState(false);

  // Initialize form values when profile changes or modal opens
  useEffect(() => {
    console.log('🔄 ProfileSettingsModal: useEffect triggered', { 
      hasProfile: !!profile, 
      isOpen,
      profileId: profile?.id,
    });
    
    if (profile && isOpen) {
      console.log('📋 ProfileSettingsModal: Initializing form with profile:', {
        displayName: profile.displayName,
        aboutMe: profile.aboutMe,
        aboutMeType: typeof profile.aboutMe,
        aboutMeProcessed: profile.aboutMe || '',
      });
      
      setDisplayName(profile.displayName);
      setAboutMe(profile.aboutMe || '');
      setHasChanges(false);
      setError(null);
      
      console.log('✅ ProfileSettingsModal: Form initialized');
    } else {
      console.log('⏭️ ProfileSettingsModal: Skipping initialization', {
        reason: !profile ? 'no profile' : 'modal not open'
      });
    }
  }, [profile, isOpen]);

  // Track changes
  useEffect(() => {
    if (profile) {
      const displayNameChanged = displayName !== profile.displayName;
      const aboutMeChanged = aboutMe !== (profile.aboutMe || '');
      setHasChanges(displayNameChanged || aboutMeChanged);
    }
  }, [displayName, aboutMe, profile]);

  const validateForm = (): string | null => {
    // Validate display name for full accounts
    if (profile?.accountType === 'full' && displayName !== profile.displayName) {
      if (displayName.length < 3) {
        return 'Display name must be at least 3 characters';
      }
      if (displayName.length > 50) {
        return 'Display name must be less than 50 characters';
      }
    }

    // Validate about me length
    if (aboutMe.length > 1000) {
      return 'About me must be less than 1000 characters';
    }

    return null;
  };

  const handleSave = async () => {
    if (!profile) return;

    const validationError = validateForm();
    if (validationError) {
      setError(validationError);
      return;
    }

    setIsLoading(true);
    setError(null);

    try {
      const updates: { displayName?: string; aboutMe?: string } = {};

      // For guest accounts, only allow aboutMe updates
      if (profile.accountType === 'guest') {
        if (aboutMe !== (profile.aboutMe || '')) {
          updates.aboutMe = aboutMe;
        }
      } else {
        // For full accounts, allow both displayName and aboutMe updates
        if (displayName !== profile.displayName) {
          updates.displayName = displayName;
        }
        if (aboutMe !== (profile.aboutMe || '')) {
          updates.aboutMe = aboutMe;
        }
      }

      const updatedProfile = await updateUserProfile(updates, profile.id);
      setProfile(updatedProfile);
      onClose();
    } catch (err) {
      setError('Failed to update profile. Please try again.');
    } finally {
      setIsLoading(false);
    }
  };

  const handleCancel = () => {
    onClose();
  };

  const handleBackdropClick = (e: React.MouseEvent) => {
    if (e.target === e.currentTarget) {
      onClose();
    }
  };

  if (!isOpen || !profile) {
    return null;
  }

  const isGuestAccount = profile.accountType === 'guest';
  const canSave = hasChanges && !isLoading;

  return (
    <div 
      className="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center z-50"
      onClick={handleBackdropClick}
      data-testid="modal-backdrop"
    >
      <div className="bg-white rounded-lg shadow-xl max-w-md w-full mx-4 max-h-[90vh] overflow-y-auto">
        <div className="p-6">
          <h2 className="text-xl font-semibold text-gray-900 mb-6">Profile Settings</h2>
          
          <div className="space-y-4">
            {/* Display Name Field */}
            <div>
              <label htmlFor="displayName" className="block text-sm font-medium text-gray-700 mb-1">
                Display Name
              </label>
              <input
                id="displayName"
                type="text"
                value={displayName}
                onChange={(e) => setDisplayName(e.target.value)}
                disabled={isGuestAccount || isLoading}
                className={`w-full px-3 py-2 border border-gray-300 rounded-md shadow-sm focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-blue-500 text-gray-900 ${
                  isGuestAccount ? 'bg-gray-100 cursor-not-allowed' : ''
                }`}
                placeholder="Enter your display name"
              />
              {isGuestAccount && (
                <p className="text-xs text-gray-500 mt-1">
                  Display name cannot be changed for guest accounts
                </p>
              )}
            </div>

            {/* About Me Field */}
            <div>
              <label htmlFor="aboutMe" className="block text-sm font-medium text-gray-700 mb-1">
                About Me
              </label>
              <textarea
                id="aboutMe"
                value={aboutMe}
                onChange={(e) => setAboutMe(e.target.value)}
                disabled={isLoading}
                rows={4}
                className="w-full px-3 py-2 border border-gray-300 rounded-md shadow-sm focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-blue-500 resize-vertical text-gray-900"
                placeholder="Tell others about yourself..."
                maxLength={1000}
              />
              <p className="text-xs text-gray-500 mt-1">
                {aboutMe.length}/1000 characters
              </p>
            </div>
          </div>

          {/* Error Message */}
          {error && (
            <div className="mt-4 p-3 bg-red-50 border border-red-200 rounded-md">
              <p className="text-sm text-red-600">{error}</p>
            </div>
          )}

          {/* Action Buttons */}
          <div className="flex justify-end space-x-3 mt-6">
            <button
              type="button"
              onClick={handleCancel}
              disabled={isLoading}
              className="px-4 py-2 text-sm font-medium text-gray-700 bg-white border border-gray-300 rounded-md shadow-sm hover:bg-gray-50 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-blue-500 disabled:opacity-50 disabled:cursor-not-allowed"
            >
              Cancel
            </button>
            <button
              type="button"
              onClick={handleSave}
              disabled={!canSave}
              className="px-4 py-2 text-sm font-medium text-white bg-blue-600 border border-transparent rounded-md shadow-sm hover:bg-blue-700 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-blue-500 disabled:opacity-50 disabled:cursor-not-allowed"
            >
              {isLoading ? 'Saving...' : 'Save Changes'}
            </button>
          </div>
        </div>
      </div>
    </div>
  );
};

export default ProfileSettingsModal;
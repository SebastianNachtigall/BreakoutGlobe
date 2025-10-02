import React, { useState } from 'react';
import { UserProfile } from '../types/models';
import { createGuestProfile, APIError } from '../services/api';
import { AvatarImageUpload } from './AvatarImageUpload';

interface ProfileCreationModalProps {
  isOpen: boolean;
  onProfileCreated: (profile: UserProfile) => void;
  onClose: () => void;
}

interface FormData {
  displayName: string;
  aboutMe: string;
  avatarFile?: File;
}

interface FormErrors {
  displayName?: string;
  aboutMe?: string;
  avatar?: string;
  general?: string;
}

const ProfileCreationModal: React.FC<ProfileCreationModalProps> = ({
  isOpen,
  onProfileCreated,
  onClose,
}) => {
  const [formData, setFormData] = useState<FormData>({
    displayName: '',
    aboutMe: '',
  });
  const [errors, setErrors] = useState<FormErrors>({});
  const [isLoading, setIsLoading] = useState(false);

  const validateForm = (): boolean => {
    const newErrors: FormErrors = {};

    // Display name validation
    if (!formData.displayName.trim()) {
      newErrors.displayName = 'Display name is required';
    } else if (formData.displayName.trim().length < 3) {
      newErrors.displayName = 'Display name must be at least 3 characters';
    } else if (formData.displayName.trim().length > 50) {
      newErrors.displayName = 'Display name must be 50 characters or less';
    }

    // About me validation
    if (formData.aboutMe.length > 500) {
      newErrors.aboutMe = 'About me must be 500 characters or less';
    }

    // Avatar validation is now handled by AvatarImageUpload component

    setErrors(newErrors);
    return Object.keys(newErrors).length === 0;
  };

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();

    if (!validateForm()) {
      return;
    }

    setIsLoading(true);
    setErrors({});

    try {
      console.log('🚀 ProfileCreationModal: Starting profile creation...');
      console.log('📝 ProfileCreationModal: Form data:', {
        displayName: formData.displayName.trim(),
        aboutMe: formData.aboutMe.trim(),
        aboutMeLength: formData.aboutMe.trim().length,
        aboutMeAfterProcessing: formData.aboutMe.trim() || undefined,
        avatarFile: formData.avatarFile ? {
          name: formData.avatarFile.name,
          size: formData.avatarFile.size,
          type: formData.avatarFile.type,
        } : undefined,
      });

      const requestData = {
        displayName: formData.displayName.trim(),
        aboutMe: formData.aboutMe.trim() || undefined,
        avatarFile: formData.avatarFile,
      };

      console.log('📡 ProfileCreationModal: Sending to API:', requestData);

      const profile = await createGuestProfile(requestData);

      console.log('✅ ProfileCreationModal: Profile created successfully:', profile);
      console.log('📦 ProfileCreationModal: aboutMe in response:', {
        aboutMe: profile.aboutMe,
        aboutMeType: typeof profile.aboutMe,
        aboutMeLength: profile.aboutMe?.length || 0,
      });

      onProfileCreated(profile);
    } catch (error) {
      console.error('❌ ProfileCreationModal: Failed to create profile:', error);

      if (error instanceof APIError) {
        setErrors({ general: error.message });
      } else {
        setErrors({ general: 'Failed to create profile. Please try again.' });
      }
    } finally {
      setIsLoading(false);
    }
  };

  const handleInputChange = (field: keyof FormData, value: string) => {
    setFormData(prev => ({ ...prev, [field]: value }));
    // Clear error when user starts typing
    if (errors[field]) {
      setErrors(prev => ({ ...prev, [field]: undefined }));
    }
  };

  const handleAvatarSelected = (file: File) => {
    console.log('🖼️ ProfileCreationModal: Avatar selected:', {
      name: file.name,
      size: file.size,
      type: file.type,
    });
    setFormData(prev => ({ ...prev, avatarFile: file }));
    // Clear any avatar errors when a new file is selected
    setErrors(prev => ({ ...prev, avatar: undefined }));
  };

  const handleAvatarError = (error: string) => {
    setErrors(prev => ({ ...prev, avatar: error }));
  };

  const handleBackdropClick = (e: React.MouseEvent) => {
    // Prevent closing modal by clicking outside - profile creation is required
    e.preventDefault();
  };

  if (!isOpen) {
    return null;
  }

  return (
    <div
      className="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center z-[9999]"
      data-testid="modal-backdrop"
      onClick={handleBackdropClick}
    >
      <div className="bg-white rounded-lg shadow-xl max-w-md w-full mx-4 max-h-[90vh] overflow-y-auto">
        <div className="p-6">
          <div className="mb-6">
            <h2 className="text-2xl font-bold text-gray-900 mb-2">
              Create Your Profile
            </h2>
            <p className="text-gray-600">
              A profile is required to use the app. Set up your profile to join the map and collaborate with others.
            </p>
          </div>

          <form onSubmit={handleSubmit} className="space-y-4">
            {/* Display Name */}
            <div>
              <label htmlFor="displayName" className="block text-sm font-medium text-gray-700 mb-1">
                Display Name *
              </label>
              <input
                type="text"
                id="displayName"
                value={formData.displayName}
                onChange={(e) => handleInputChange('displayName', e.target.value)}
                className={`w-full px-3 py-2 border rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500 ${errors.displayName ? 'border-red-500' : 'border-gray-300'
                  }`}
                placeholder="Enter your display name"
                maxLength={50}
                disabled={isLoading}
              />
              {errors.displayName && (
                <p className="mt-1 text-sm text-red-600">{errors.displayName}</p>
              )}
            </div>

            {/* About Me */}
            <div>
              <label htmlFor="aboutMe" className="block text-sm font-medium text-gray-700 mb-1">
                About Me
              </label>
              <textarea
                id="aboutMe"
                value={formData.aboutMe}
                onChange={(e) => handleInputChange('aboutMe', e.target.value)}
                className={`w-full px-3 py-2 border rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500 resize-none ${errors.aboutMe ? 'border-red-500' : 'border-gray-300'
                  }`}
                placeholder="Tell others about yourself (optional)"
                rows={3}
                maxLength={500}
                disabled={isLoading}
              />
              <div className="flex justify-between items-center mt-1">
                {errors.aboutMe && (
                  <p className="text-sm text-red-600">{errors.aboutMe}</p>
                )}
                <p className="text-sm text-gray-500 ml-auto">
                  {formData.aboutMe.length}/500
                </p>
              </div>
            </div>

            {/* Avatar Upload */}
            <div>
              <label className="block text-sm font-medium text-gray-700 mb-2">
                Avatar Image
              </label>
              <AvatarImageUpload
                onImageSelected={handleAvatarSelected}
                onError={handleAvatarError}
                disabled={isLoading}
              />
              {errors.avatar && (
                <p className="mt-2 text-sm text-red-600">{errors.avatar}</p>
              )}
            </div>

            {/* General Error */}
            {errors.general && (
              <div className="p-3 bg-red-50 border border-red-200 rounded-md">
                <p className="text-sm text-red-600">{errors.general}</p>
              </div>
            )}

            {/* Buttons */}
            <div className="flex justify-end pt-4">
              <button
                type="submit"
                className="px-6 py-2 text-white bg-blue-600 border border-transparent rounded-md hover:bg-blue-700 focus:outline-none focus:ring-2 focus:ring-blue-500 disabled:opacity-50 disabled:cursor-not-allowed"
                disabled={isLoading}
              >
                {isLoading ? 'Creating Profile...' : 'Create Profile'}
              </button>
            </div>
          </form>
        </div>
      </div>
    </div>
  );
};

export default ProfileCreationModal;
import React from 'react';
import type { UserProfile } from '../types/models';

interface ProfileCardProps {
  userProfile: UserProfile;
  onClose: () => void;
}

export const ProfileCard: React.FC<ProfileCardProps> = ({ userProfile, onClose }) => {
  const generateInitials = (displayName: string): string => {
    const words = displayName.trim().split(/\s+/);
    if (words.length >= 2) {
      return (words[0][0] + words[1][0]).toUpperCase();
    } else if (words.length === 1 && words[0].length >= 2) {
      return words[0].substring(0, 2).toUpperCase();
    } else {
      return words[0][0].toUpperCase();
    }
  };

  const getRoleBadgeColor = (role: string) => {
    switch (role) {
      case 'admin':
        return 'bg-yellow-100 text-yellow-800 border-yellow-200';
      case 'superadmin':
        return 'bg-red-100 text-red-800 border-red-200';
      default:
        return 'bg-blue-100 text-blue-800 border-blue-200';
    }
  };

  return (
    <div 
      className="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center z-50"
      onClick={onClose}
    >
      <div 
        className="bg-white rounded-lg shadow-xl max-w-sm w-full mx-4 p-6"
        data-testid="profile-card"
        onClick={(e) => e.stopPropagation()}
      >
        {/* Header with close button */}
        <div className="flex justify-between items-start mb-4">
          <h3 className="text-lg font-semibold text-gray-900">User Profile</h3>
          <button
            onClick={onClose}
            className="text-gray-400 hover:text-gray-600 transition-colors"
            aria-label="Close profile card"
          >
            <svg className="w-6 h-6" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M6 18L18 6M6 6l12 12" />
            </svg>
          </button>
        </div>

        {/* Avatar and basic info */}
        <div className="flex items-center space-x-4 mb-4">
          <div className="relative">
            {userProfile.avatarURL ? (
              <img
                src={userProfile.avatarURL}
                alt={userProfile.displayName}
                className="w-16 h-16 rounded-full object-cover border-2 border-gray-200"
              />
            ) : (
              <div className="w-16 h-16 rounded-full bg-gray-500 border-2 border-gray-200 flex items-center justify-center text-white text-lg font-bold">
                {generateInitials(userProfile.displayName)}
              </div>
            )}
            
            {/* Role badge */}
            <div className={`absolute -bottom-1 -right-1 px-2 py-1 rounded-full text-xs font-medium border ${getRoleBadgeColor(userProfile.role)}`}>
              {userProfile.role}
            </div>
          </div>

          <div className="flex-1">
            <h4 className="text-xl font-semibold text-gray-900">{userProfile.displayName}</h4>
            <div className="flex items-center space-x-2 mt-1">
              <span className={`inline-flex items-center px-2 py-1 rounded-full text-xs font-medium ${
                userProfile.accountType === 'full' 
                  ? 'bg-green-100 text-green-800' 
                  : 'bg-gray-100 text-gray-800'
              }`}>
                {userProfile.accountType === 'full' ? 'Full Account' : 'Guest'}
              </span>
              
              {userProfile.accountType === 'full' && (
                <span className={`inline-flex items-center px-2 py-1 rounded-full text-xs font-medium ${
                  userProfile.emailVerified 
                    ? 'bg-green-100 text-green-800' 
                    : 'bg-yellow-100 text-yellow-800'
                }`}>
                  {userProfile.emailVerified ? 'Verified' : 'Unverified'}
                </span>
              )}
            </div>
          </div>
        </div>

        {/* About me section */}
        {userProfile.aboutMe && (
          <div className="mb-4">
            <h5 className="text-sm font-medium text-gray-700 mb-2">About</h5>
            <p className="text-sm text-gray-600 leading-relaxed">{userProfile.aboutMe}</p>
          </div>
        )}

        {/* Member since */}
        <div className="text-xs text-gray-500 border-t pt-3">
          Member since {userProfile.createdAt.toLocaleDateString()}
        </div>
      </div>
    </div>
  );
};
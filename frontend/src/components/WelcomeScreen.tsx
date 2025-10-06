import React from 'react';

interface WelcomeScreenProps {
  isOpen: boolean;
  onGetStarted?: () => void;
  onCreateProfile: () => void;
  onSignup?: () => void;
  onLogin?: () => void;
  hideContent?: boolean; // Hide content when modals are open
}

const WelcomeScreen: React.FC<WelcomeScreenProps> = ({ isOpen, onGetStarted, onCreateProfile, onSignup, onLogin, hideContent = false }) => {
  if (!isOpen) {
    return null;
  }

  return (
    <div className={`fixed inset-0 bg-gray-50 overflow-y-auto ${hideContent ? 'z-0' : 'z-[9999]'}`}>
      <div className={`min-h-full flex items-center justify-center p-4 sm:p-6 ${hideContent ? 'invisible' : ''}`}>
        <div className="max-w-md w-full text-center py-8 sm:py-12">
          {/* Welcome Title - Responsive sizing */}
          <h1 className="text-4xl sm:text-5xl md:text-6xl font-bold text-gray-900 mb-6 sm:mb-8">
            Welcome
          </h1>

          {/* Map Illustration using BreakoutGlobe.svg - Mobile optimized */}
          <div className="relative rounded-xl sm:rounded-2xl p-4 sm:p-6 md:p-8 mb-6 sm:mb-8 overflow-hidden border border-gray-200 bg-white">
            <img
              src="/BreakoutGlobe2.svg"
              alt="BreakoutGlobe map illustration showing POIs and video call functionality"
              className="w-full h-auto max-w-xs sm:max-w-sm mx-auto"
            />
          </div>

          {/* Description - Mobile optimized text */}
          <div className="mb-6 sm:mb-8 px-2 sm:px-0">
            <h2 className="text-xl sm:text-2xl font-bold text-gray-900 mb-3 sm:mb-4 leading-tight">
              Join POIs on the map to initiate video calls.
            </h2>
            <p className="text-base sm:text-lg text-gray-600 leading-relaxed">
              Useful for user-driven breakout sessions in a workshop scenario.
            </p>
          </div>

          {/* Authentication Options - Mobile optimized */}
          <div className="space-y-3 sm:space-y-4">
            {/* Primary: Create Full Account */}
            {onSignup && (
              <button
                onClick={onSignup}
                className="w-full bg-blue-600 text-white text-base sm:text-lg font-semibold py-3 sm:py-4 px-6 sm:px-8 rounded-lg sm:rounded-xl hover:bg-blue-700 active:bg-blue-800 transition-colors duration-200 shadow-lg touch-manipulation"
              >
                Create Full Account
              </button>
            )}
            
            {/* Secondary: Login */}
            {onLogin && (
              <button
                onClick={onLogin}
                className="w-full bg-gray-600 text-white text-base sm:text-lg font-semibold py-3 sm:py-4 px-6 sm:px-8 rounded-lg sm:rounded-xl hover:bg-gray-700 active:bg-gray-800 transition-colors duration-200 shadow-lg touch-manipulation"
              >
                Login
              </button>
            )}
            
            {/* Tertiary: Continue as Guest */}
            <button
              onClick={onCreateProfile}
              className="w-full border-2 border-gray-300 text-gray-700 text-base sm:text-lg font-semibold py-3 sm:py-4 px-6 sm:px-8 rounded-lg sm:rounded-xl hover:bg-gray-50 active:bg-gray-100 transition-colors duration-200 touch-manipulation"
            >
              Continue as Guest
            </button>
            
            {/* Fallback for old onGetStarted prop */}
            {!onSignup && !onLogin && onGetStarted && (
              <button
                onClick={onGetStarted}
                className="w-full bg-blue-600 text-white text-base sm:text-lg font-semibold py-3 sm:py-4 px-6 sm:px-8 rounded-lg sm:rounded-xl hover:bg-blue-700 active:bg-blue-800 transition-colors duration-200 shadow-lg touch-manipulation"
              >
                Get Started
              </button>
            )}
          </div>
        </div>
      </div>
    </div>
  );
};

export default WelcomeScreen;
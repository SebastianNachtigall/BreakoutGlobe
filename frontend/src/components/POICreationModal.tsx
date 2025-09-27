import React, { useState, useEffect, useCallback } from 'react';

interface POICreationData {
    name: string;
    description: string;
    maxParticipants: number;
    position: { lat: number; lng: number };
}

interface POICreationModalProps {
    isOpen: boolean;
    position: { lat: number; lng: number };
    onCreate: (data: POICreationData) => void;
    onCancel: () => void;
    isLoading?: boolean;
}

interface FormData {
    name: string;
    description: string;
    maxParticipants: number;
    latitude: number;
    longitude: number;
}

interface FormErrors {
    name?: string;
    description?: string;
    maxParticipants?: string;
    latitude?: string;
    longitude?: string;
}

export const POICreationModal: React.FC<POICreationModalProps> = ({
    isOpen,
    position,
    onCreate,
    onCancel,
    isLoading = false
}) => {
    const [formData, setFormData] = useState<FormData>({
        name: '',
        description: '',
        maxParticipants: 5,
        latitude: position.lat,
        longitude: position.lng
    });

    const [errors, setErrors] = useState<FormErrors>({});
    const [touched, setTouched] = useState<Record<string, boolean>>({});

    // Reset form when modal opens/closes
    useEffect(() => {
        if (isOpen) {
            setFormData({
                name: '',
                description: '',
                maxParticipants: 5,
                latitude: position.lat,
                longitude: position.lng
            });
            setErrors({});
            setTouched({});
        }
    }, [isOpen, position]);

    // Validation functions
    const validateName = (name: string): string | undefined => {
        if (!name.trim()) return 'Name is required';
        if (name.length < 3) return 'Name must be at least 3 characters';
        if (name.length > 100) return 'Name must be less than 100 characters';
        return undefined;
    };

    const validateDescription = (description: string): string | undefined => {
        if (!description.trim()) return 'Description is required';
        if (description.length > 500) return 'Description must be less than 500 characters';
        return undefined;
    };

    const validateMaxParticipants = (maxParticipants: number): string | undefined => {
        if (maxParticipants < 1) return 'Must be at least 1 participant';
        if (maxParticipants > 100) return 'Cannot exceed 100 participants';
        return undefined;
    };

    const validateLatitude = (latitude: number): string | undefined => {
        if (latitude < -90 || latitude > 90) return 'Latitude must be between -90 and 90';
        return undefined;
    };

    const validateLongitude = (longitude: number): string | undefined => {
        if (longitude < -180 || longitude > 180) return 'Longitude must be between -180 and 180';
        return undefined;
    };

    // Validate form
    const validateForm = useCallback((): FormErrors => {
        return {
            name: validateName(formData.name),
            description: validateDescription(formData.description),
            maxParticipants: validateMaxParticipants(formData.maxParticipants),
            latitude: validateLatitude(formData.latitude),
            longitude: validateLongitude(formData.longitude)
        };
    }, [formData]);

    const handleInputChange = (field: keyof FormData, value: string | number) => {
        const newFormData = { ...formData, [field]: value };
        setFormData(newFormData);
        setTouched(prev => ({ ...prev, [field]: true }));
        
        // Validate with the new form data
        const newErrors = {
            name: validateName(newFormData.name),
            description: validateDescription(newFormData.description),
            maxParticipants: validateMaxParticipants(newFormData.maxParticipants),
            latitude: validateLatitude(newFormData.latitude),
            longitude: validateLongitude(newFormData.longitude)
        };
        setErrors(newErrors);
    };

    const handleBlur = (field: keyof FormData) => {
        setTouched(prev => ({ ...prev, [field]: true }));
        
        // Validate on blur
        const newErrors = {
            name: validateName(formData.name),
            description: validateDescription(formData.description),
            maxParticipants: validateMaxParticipants(formData.maxParticipants),
            latitude: validateLatitude(formData.latitude),
            longitude: validateLongitude(formData.longitude)
        };
        setErrors(newErrors);
    };

    const isFormValid = () => {
        const formErrors = validateForm();
        return !Object.values(formErrors).some(error => error !== undefined);
    };

    const handleSubmit = (e: React.FormEvent) => {
        e.preventDefault();
        
        // Mark all fields as touched to show validation errors
        setTouched({
            name: true,
            description: true,
            maxParticipants: true,
            latitude: true,
            longitude: true
        });

        if (isFormValid()) {
            onCreate({
                name: formData.name,
                description: formData.description,
                maxParticipants: formData.maxParticipants,
                position: {
                    lat: formData.latitude,
                    lng: formData.longitude
                }
            });
        }
    };

    const handleKeyDown = (e: React.KeyboardEvent) => {
        if (e.key === 'Escape') {
            onCancel();
        }
    };

    // Handle escape key globally when modal is open
    useEffect(() => {
        const handleEscapeKey = (e: KeyboardEvent) => {
            if (e.key === 'Escape') {
                onCancel();
            }
        };

        if (isOpen) {
            document.addEventListener('keydown', handleEscapeKey);
            return () => document.removeEventListener('keydown', handleEscapeKey);
        }
    }, [isOpen, onCancel]);

    const handleOverlayClick = (e: React.MouseEvent) => {
        if (e.target === e.currentTarget) {
            onCancel();
        }
    };

    if (!isOpen) {
        return null;
    }

    return (
        <div 
            className="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center z-[9999]"
            data-testid="modal-overlay"
            onClick={handleOverlayClick}
            onKeyDown={handleKeyDown}
            tabIndex={-1}
        >
            <div className="bg-white rounded-lg shadow-xl max-w-md w-full mx-4 max-h-[90vh] overflow-hidden" data-testid="poi-creation-modal">
                <div className="bg-gray-800 text-white p-4 flex justify-between items-center">
                    <h2 className="text-lg font-semibold">Create New POI</h2>
                    <button 
                        className="text-gray-400 hover:text-white text-xl transition-colors"
                        onClick={onCancel}
                        aria-label="Close modal"
                    >
                        ✕
                    </button>
                </div>

                <form onSubmit={handleSubmit} className="p-6 space-y-4">
                    <div className="space-y-2">
                        <label htmlFor="poi-name" className="block text-sm font-medium text-gray-700">
                            Name *
                        </label>
                        <input
                            id="poi-name"
                            type="text"
                            value={formData.name}
                            onChange={(e) => handleInputChange('name', e.target.value)}
                            onBlur={() => handleBlur('name')}
                            className={`w-full px-3 py-2 border rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500 ${
                                errors.name && touched.name 
                                    ? 'border-red-500 bg-red-50' 
                                    : 'border-gray-300'
                            }`}
                            disabled={isLoading}
                            placeholder="Enter POI name"
                        />
                        <div className="flex justify-end">
                            <span className="text-xs text-gray-500">{formData.name.length}/100</span>
                        </div>
                        {errors.name && touched.name && (
                            <div className="text-sm text-red-600">{errors.name}</div>
                        )}
                    </div>

                    <div className="space-y-2">
                        <label htmlFor="poi-description" className="block text-sm font-medium text-gray-700">
                            Description *
                        </label>
                        <textarea
                            id="poi-description"
                            value={formData.description}
                            onChange={(e) => handleInputChange('description', e.target.value)}
                            onBlur={() => handleBlur('description')}
                            className={`w-full px-3 py-2 border rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500 resize-none ${
                                errors.description && touched.description 
                                    ? 'border-red-500 bg-red-50' 
                                    : 'border-gray-300'
                            }`}
                            rows={3}
                            disabled={isLoading}
                            placeholder="Describe what this POI is for"
                        />
                        <div className="flex justify-end">
                            <span className="text-xs text-gray-500">{formData.description.length}/500</span>
                        </div>
                        {errors.description && touched.description && (
                            <div className="text-sm text-red-600">{errors.description}</div>
                        )}
                    </div>

                    <div className="space-y-2">
                        <label htmlFor="poi-max-participants" className="block text-sm font-medium text-gray-700">
                            Max Participants *
                        </label>
                        <input
                            id="poi-max-participants"
                            type="number"
                            min="1"
                            max="100"
                            value={formData.maxParticipants}
                            onChange={(e) => handleInputChange('maxParticipants', parseInt(e.target.value) || 0)}
                            onBlur={() => handleBlur('maxParticipants')}
                            className={`w-full px-3 py-2 border rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500 ${
                                errors.maxParticipants && touched.maxParticipants 
                                    ? 'border-red-500 bg-red-50' 
                                    : 'border-gray-300'
                            }`}
                            disabled={isLoading}
                        />
                        {errors.maxParticipants && touched.maxParticipants && (
                            <div className="text-sm text-red-600">{errors.maxParticipants}</div>
                        )}
                    </div>

                    <div className="grid grid-cols-2 gap-4">
                        <div className="space-y-2">
                            <label htmlFor="poi-latitude" className="block text-sm font-medium text-gray-700">
                                Latitude
                            </label>
                            <input
                                id="poi-latitude"
                                type="number"
                                step="any"
                                value={formData.latitude}
                                onChange={(e) => handleInputChange('latitude', parseFloat(e.target.value) || 0)}
                                onBlur={() => handleBlur('latitude')}
                                className={`w-full px-3 py-2 border rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500 ${
                                    errors.latitude && touched.latitude 
                                        ? 'border-red-500 bg-red-50' 
                                        : 'border-gray-300'
                                }`}
                                disabled={isLoading}
                            />
                            {errors.latitude && touched.latitude && (
                                <div className="text-sm text-red-600">{errors.latitude}</div>
                            )}
                        </div>

                        <div className="space-y-2">
                            <label htmlFor="poi-longitude" className="block text-sm font-medium text-gray-700">
                                Longitude
                            </label>
                            <input
                                id="poi-longitude"
                                type="number"
                                step="any"
                                value={formData.longitude}
                                onChange={(e) => handleInputChange('longitude', parseFloat(e.target.value) || 0)}
                                onBlur={() => handleBlur('longitude')}
                                className={`w-full px-3 py-2 border rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500 ${
                                    errors.longitude && touched.longitude 
                                        ? 'border-red-500 bg-red-50' 
                                        : 'border-gray-300'
                                }`}
                                disabled={isLoading}
                            />
                            {errors.longitude && touched.longitude && (
                                <div className="text-sm text-red-600">{errors.longitude}</div>
                            )}
                        </div>
                    </div>

                    <div className="flex justify-end space-x-3 pt-6 border-t border-gray-200">
                        <button
                            type="button"
                            onClick={onCancel}
                            className="px-4 py-2 text-sm font-medium text-gray-700 bg-white border border-gray-300 rounded-md hover:bg-gray-50 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-blue-500 disabled:opacity-50 disabled:cursor-not-allowed"
                            disabled={isLoading}
                        >
                            Cancel
                        </button>
                        <button
                            type="submit"
                            className="px-4 py-2 text-sm font-medium text-white bg-blue-600 border border-transparent rounded-md hover:bg-blue-700 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-blue-500 disabled:opacity-50 disabled:cursor-not-allowed"
                            disabled={!isFormValid() || isLoading}
                        >
                            {isLoading ? (
                                <>
                                    <svg className="animate-spin -ml-1 mr-2 h-4 w-4 text-white inline" xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24">
                                        <circle className="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" strokeWidth="4"></circle>
                                        <path className="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"></path>
                                    </svg>
                                    Creating...
                                </>
                            ) : (
                                'Create POI'
                            )}
                        </button>
                    </div>
                </form>
            </div>
        </div>
    );
};
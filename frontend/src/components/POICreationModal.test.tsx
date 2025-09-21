import { describe, it, expect, vi, beforeEach } from 'vitest';
import { render, screen, waitFor } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { POICreationModal } from './POICreationModal';

const mockOnCreate = vi.fn();
const mockOnCancel = vi.fn();

const defaultProps = {
    isOpen: true,
    position: { lat: 40.7128, lng: -74.0060 },
    onCreate: mockOnCreate,
    onCancel: mockOnCancel
};

describe('POICreationModal', () => {
    beforeEach(() => {
        vi.clearAllMocks();
    });

    it('should not render when isOpen is false', () => {
        render(<POICreationModal {...defaultProps} isOpen={false} />);
        
        expect(screen.queryByTestId('poi-creation-modal')).not.toBeInTheDocument();
    });

    it('should render modal when isOpen is true', () => {
        render(<POICreationModal {...defaultProps} />);
        
        expect(screen.getByTestId('poi-creation-modal')).toBeInTheDocument();
        expect(screen.getByText(/create new poi/i)).toBeInTheDocument();
    });

    it('should display form fields', () => {
        render(<POICreationModal {...defaultProps} />);
        
        expect(screen.getByLabelText(/name/i)).toBeInTheDocument();
        expect(screen.getByLabelText(/description/i)).toBeInTheDocument();
        expect(screen.getByLabelText(/max participants/i)).toBeInTheDocument();
        expect(screen.getByDisplayValue('40.7128')).toBeInTheDocument(); // latitude
        expect(screen.getByDisplayValue('-74.006')).toBeInTheDocument(); // longitude
    });

    it('should show validation errors for empty required fields', async () => {
        const user = userEvent.setup();
        render(<POICreationModal {...defaultProps} />);
        
        // First focus and blur the fields to trigger validation
        const nameInput = screen.getByLabelText(/name/i);
        const descriptionInput = screen.getByLabelText(/description/i);
        
        await user.click(nameInput);
        await user.tab(); // This will blur the name field
        
        await waitFor(() => {
            expect(screen.getByText(/name is required/i)).toBeInTheDocument();
        });
        
        await user.click(descriptionInput);
        await user.tab(); // This will blur the description field
        
        await waitFor(() => {
            expect(screen.getByText(/description is required/i)).toBeInTheDocument();
        });
    });

    it('should validate name length', async () => {
        const user = userEvent.setup();
        render(<POICreationModal {...defaultProps} />);
        
        const nameInput = screen.getByLabelText(/name/i);
        
        // Test too short name
        await user.type(nameInput, 'ab');
        await user.tab(); // Trigger blur for validation
        
        await waitFor(() => {
            expect(screen.getByText(/name must be at least 3 characters/i)).toBeInTheDocument();
        });
        
        // Test too long name
        await user.clear(nameInput);
        await user.type(nameInput, 'a'.repeat(101));
        await user.tab();
        
        await waitFor(() => {
            expect(screen.getByText(/name must be less than 100 characters/i)).toBeInTheDocument();
        });
    });

    it('should validate description length', async () => {
        const user = userEvent.setup();
        render(<POICreationModal {...defaultProps} />);
        
        const descriptionInput = screen.getByLabelText(/description/i);
        
        // Test too long description
        await user.type(descriptionInput, 'a'.repeat(501));
        await user.tab();
        
        await waitFor(() => {
            expect(screen.getByText(/description must be less than 500 characters/i)).toBeInTheDocument();
        });
    });

    it('should validate max participants range', async () => {
        const user = userEvent.setup();
        render(<POICreationModal {...defaultProps} />);
        
        const maxParticipantsInput = screen.getByLabelText(/max participants/i);
        
        // Test too low
        await user.clear(maxParticipantsInput);
        await user.type(maxParticipantsInput, '0');
        await user.tab();
        
        await waitFor(() => {
            expect(screen.getByText(/must be at least 1 participant/i)).toBeInTheDocument();
        });
        
        // Test too high
        await user.clear(maxParticipantsInput);
        await user.type(maxParticipantsInput, '101');
        await user.tab();
        
        await waitFor(() => {
            expect(screen.getByText(/cannot exceed 100 participants/i)).toBeInTheDocument();
        });
    });

    it('should validate coordinate ranges', async () => {
        const user = userEvent.setup();
        render(<POICreationModal {...defaultProps} />);
        
        const latInput = screen.getByDisplayValue('40.7128');
        const lngInput = screen.getByDisplayValue('-74.006');
        
        // Test invalid latitude
        await user.clear(latInput);
        await user.type(latInput, '91');
        await user.tab();
        
        await waitFor(() => {
            expect(screen.getByText(/latitude must be between -90 and 90/i)).toBeInTheDocument();
        });
        
        // Test invalid longitude
        await user.clear(lngInput);
        await user.type(lngInput, '181');
        await user.tab();
        
        await waitFor(() => {
            expect(screen.getByText(/longitude must be between -180 and 180/i)).toBeInTheDocument();
        });
    });

    it('should show real-time character count for name and description', async () => {
        const user = userEvent.setup();
        render(<POICreationModal {...defaultProps} />);
        
        const nameInput = screen.getByLabelText(/name/i);
        const descriptionInput = screen.getByLabelText(/description/i);
        
        await user.type(nameInput, 'Test POI');
        await waitFor(() => {
            expect(screen.getByText('8/100')).toBeInTheDocument();
        });
        
        await user.type(descriptionInput, 'Test description');
        await waitFor(() => {
            expect(screen.getByText('16/500')).toBeInTheDocument();
        });
    });

    it('should call onCreate with valid form data', async () => {
        const user = userEvent.setup();
        render(<POICreationModal {...defaultProps} />);
        
        // Fill in valid form data
        await user.type(screen.getByLabelText(/name/i), 'Test Meeting Room');
        await user.type(screen.getByLabelText(/description/i), 'A test meeting room for discussions');
        await user.clear(screen.getByLabelText(/max participants/i));
        await user.type(screen.getByLabelText(/max participants/i), '10');
        
        const createButton = screen.getByText(/create poi/i);
        await user.click(createButton);
        
        expect(mockOnCreate).toHaveBeenCalledWith({
            name: 'Test Meeting Room',
            description: 'A test meeting room for discussions',
            maxParticipants: 10,
            position: { lat: 40.7128, lng: -74.0060 }
        });
    });

    it('should not call onCreate with invalid form data', async () => {
        const user = userEvent.setup();
        render(<POICreationModal {...defaultProps} />);
        
        // Leave name empty (invalid)
        await user.type(screen.getByLabelText(/description/i), 'Test description');
        
        const createButton = screen.getByText(/create poi/i);
        await user.click(createButton);
        
        expect(mockOnCreate).not.toHaveBeenCalled();
    });

    it('should call onCancel when cancel button is clicked', async () => {
        const user = userEvent.setup();
        render(<POICreationModal {...defaultProps} />);
        
        const cancelButton = screen.getByText(/cancel/i);
        await user.click(cancelButton);
        
        expect(mockOnCancel).toHaveBeenCalled();
    });

    it('should call onCancel when escape key is pressed', async () => {
        const user = userEvent.setup();
        render(<POICreationModal {...defaultProps} />);
        
        await user.keyboard('{Escape}');
        
        expect(mockOnCancel).toHaveBeenCalled();
    });

    it('should call onCancel when clicking outside modal', async () => {
        const user = userEvent.setup();
        render(<POICreationModal {...defaultProps} />);
        
        const overlay = screen.getByTestId('modal-overlay');
        await user.click(overlay);
        
        expect(mockOnCancel).toHaveBeenCalled();
    });

    it('should disable create button while form is invalid', () => {
        render(<POICreationModal {...defaultProps} />);
        
        const createButton = screen.getByText(/create poi/i);
        expect(createButton).toBeDisabled();
    });

    it('should enable create button when form is valid', async () => {
        const user = userEvent.setup();
        render(<POICreationModal {...defaultProps} />);
        
        // Fill in valid form data
        await user.type(screen.getByLabelText(/name/i), 'Test POI');
        await user.type(screen.getByLabelText(/description/i), 'Test description');
        
        const createButton = screen.getByText(/create poi/i);
        
        await waitFor(() => {
            expect(createButton).toBeEnabled();
        });
    });

    it('should show loading state when creating POI', async () => {
        const user = userEvent.setup();
        render(<POICreationModal {...defaultProps} isLoading={true} />);
        
        expect(screen.getByText(/creating.../i)).toBeInTheDocument();
        
        const createButton = screen.getByText(/creating.../i);
        expect(createButton).toBeDisabled();
    });

    it('should reset form when modal is reopened', () => {
        const { rerender } = render(<POICreationModal {...defaultProps} isOpen={false} />);
        
        // Open modal and fill form
        rerender(<POICreationModal {...defaultProps} isOpen={true} />);
        
        const nameInput = screen.getByLabelText(/name/i) as HTMLInputElement;
        const descriptionInput = screen.getByLabelText(/description/i) as HTMLInputElement;
        
        expect(nameInput.value).toBe('');
        expect(descriptionInput.value).toBe('');
    });
});
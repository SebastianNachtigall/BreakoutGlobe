import { describe, it, expect, vi } from 'vitest';
import { render, screen, fireEvent } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { POIContextMenu } from './POIContextMenu';

describe('POIContextMenu', () => {
  const mockPosition = { x: 100, y: 200 };
  const mockMapPosition = { lat: 40.7128, lng: -74.0060 };

  describe('Rendering', () => {
    it('should render context menu at correct position', () => {
      const onCreatePOI = vi.fn();
      const onClose = vi.fn();
      
      render(
        <POIContextMenu
          position={mockPosition}
          mapPosition={mockMapPosition}
          onCreatePOI={onCreatePOI}
          onClose={onClose}
        />
      );
      
      const menu = screen.getByTestId('poi-context-menu');
      expect(menu).toBeInTheDocument();
      expect(menu).toHaveStyle({
        left: '100px',
        top: '200px'
      });
    });

    it('should display "Create POI" option', () => {
      const onCreatePOI = vi.fn();
      const onClose = vi.fn();
      
      render(
        <POIContextMenu
          position={mockPosition}
          mapPosition={mockMapPosition}
          onCreatePOI={onCreatePOI}
          onClose={onClose}
        />
      );
      
      expect(screen.getByText('Create POI')).toBeInTheDocument();
    });
  });

  describe('Interaction', () => {
    it('should call onCreatePOI when Create POI is clicked', async () => {
      const user = userEvent.setup();
      const onCreatePOI = vi.fn();
      const onClose = vi.fn();
      
      render(
        <POIContextMenu
          position={mockPosition}
          mapPosition={mockMapPosition}
          onCreatePOI={onCreatePOI}
          onClose={onClose}
        />
      );
      
      await user.click(screen.getByText('Create POI'));
      
      expect(onCreatePOI).toHaveBeenCalledWith(mockMapPosition);
    });

    it('should close menu when clicking outside', () => {
      const onCreatePOI = vi.fn();
      const onClose = vi.fn();
      
      render(
        <div>
          <POIContextMenu
            position={mockPosition}
            mapPosition={mockMapPosition}
            onCreatePOI={onCreatePOI}
            onClose={onClose}
          />
          <div data-testid="outside">Outside</div>
        </div>
      );
      
      fireEvent.mouseDown(screen.getByTestId('outside'));
      expect(onClose).toHaveBeenCalled();
    });

    it('should close menu when pressing Escape', () => {
      const onCreatePOI = vi.fn();
      const onClose = vi.fn();
      
      render(
        <POIContextMenu
          position={mockPosition}
          mapPosition={mockMapPosition}
          onCreatePOI={onCreatePOI}
          onClose={onClose}
        />
      );
      
      fireEvent.keyDown(document, { key: 'Escape' });
      expect(onClose).toHaveBeenCalled();
    });
  });
});
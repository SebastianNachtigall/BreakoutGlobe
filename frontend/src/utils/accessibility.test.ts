import { describe, it, expect } from 'vitest';
import { 
  hasAccessibleName, 
  hasProperFocusManagement, 
  hasKeyboardNavigation,
  checkAriaAttributes,
  isSemanticHTML 
} from './accessibility';

describe('Accessibility utilities', () => {
  describe('hasAccessibleName', () => {
    it('should detect accessible name from aria-label', () => {
      const button = document.createElement('button');
      button.setAttribute('aria-label', 'Close dialog');
      button.textContent = 'X';
      
      expect(hasAccessibleName(button)).toBe(true);
    });

    it('should detect accessible name from text content', () => {
      const button = document.createElement('button');
      button.textContent = 'Save Changes';
      
      expect(hasAccessibleName(button)).toBe(true);
    });

    it('should return false for elements without accessible name', () => {
      const button = document.createElement('button');
      
      expect(hasAccessibleName(button)).toBe(false);
    });
  });

  describe('hasProperFocusManagement', () => {
    it('should detect focusable elements', () => {
      const button = document.createElement('button');
      
      expect(hasProperFocusManagement(button)).toBe(true);
    });

    it('should detect elements with tabindex', () => {
      const div = document.createElement('div');
      div.setAttribute('tabindex', '0');
      
      expect(hasProperFocusManagement(div)).toBe(true);
    });

    it('should return false for non-focusable elements', () => {
      const div = document.createElement('div');
      
      expect(hasProperFocusManagement(div)).toBe(false);
    });

    it('should handle disabled elements', () => {
      const button = document.createElement('button');
      button.setAttribute('disabled', 'true');
      
      expect(hasProperFocusManagement(button)).toBe(false);
    });
  });

  describe('hasKeyboardNavigation', () => {
    it('should detect role="button" with keyboard support', () => {
      const div = document.createElement('div');
      div.setAttribute('role', 'button');
      div.setAttribute('tabindex', '0');
      
      expect(hasKeyboardNavigation(div)).toBe(true);
    });

    it('should return false for elements without keyboard support', () => {
      const div = document.createElement('div');
      
      expect(hasKeyboardNavigation(div)).toBe(false);
    });

    it('should detect native interactive elements', () => {
      const button = document.createElement('button');
      
      expect(hasKeyboardNavigation(button)).toBe(true);
    });
  });

  describe('checkAriaAttributes', () => {
    it('should validate correct ARIA attributes', () => {
      const button = document.createElement('button');
      button.setAttribute('aria-expanded', 'false');
      button.setAttribute('aria-haspopup', 'true');
      
      const result = checkAriaAttributes(button);
      expect(result.isValid).toBe(true);
      expect(result.errors).toHaveLength(0);
    });

    it('should detect invalid ARIA attribute values', () => {
      const button = document.createElement('button');
      button.setAttribute('aria-expanded', 'maybe');
      
      const result = checkAriaAttributes(button);
      expect(result.isValid).toBe(false);
      expect(result.errors).toContain('Invalid aria-expanded value: maybe');
    });
  });

  describe('isSemanticHTML', () => {
    it('should recognize semantic HTML elements', () => {
      const button = document.createElement('button');
      
      expect(isSemanticHTML(button)).toBe(true);
    });

    it('should recognize semantic form elements', () => {
      const input = document.createElement('input');
      
      expect(isSemanticHTML(input)).toBe(true);
    });

    it('should return false for non-semantic elements', () => {
      const div = document.createElement('div');
      
      expect(isSemanticHTML(div)).toBe(false);
    });
  });
});
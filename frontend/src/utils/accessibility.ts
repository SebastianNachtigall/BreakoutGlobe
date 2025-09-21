/**
 * Accessibility utility functions for testing and validation
 */

export interface AriaValidationResult {
  isValid: boolean;
  errors: string[];
}

// Semantic HTML elements that don't need roles
const SEMANTIC_ELEMENTS = new Set([
  'button', 'input', 'select', 'textarea', 'a', 'nav', 'main', 'header', 
  'footer', 'section', 'article', 'aside', 'h1', 'h2', 'h3', 'h4', 'h5', 'h6',
  'form', 'fieldset', 'legend', 'label', 'table', 'thead', 'tbody', 'tr', 'th', 'td'
]);

// Focusable elements
const FOCUSABLE_ELEMENTS = new Set([
  'button', 'input', 'select', 'textarea', 'a', 'area', 'object', 'embed', 'iframe'
]);

// Valid ARIA attribute values
const ARIA_BOOLEAN_VALUES = new Set(['true', 'false']);
const ARIA_TRISTATE_VALUES = new Set(['true', 'false', 'mixed']);

/**
 * Check if an element has an accessible name
 */
export function hasAccessibleName(element: Element): boolean {
  // Check aria-label
  if (element.getAttribute('aria-label')?.trim()) {
    return true;
  }

  // Check aria-labelledby
  const labelledBy = element.getAttribute('aria-labelledby');
  if (labelledBy) {
    const labelElement = document.getElementById(labelledBy);
    if (labelElement?.textContent?.trim()) {
      return true;
    }
  }

  // Check text content for buttons and links
  const tagName = element.tagName.toLowerCase();
  if (['button', 'a'].includes(tagName)) {
    return Boolean(element.textContent?.trim());
  }

  // Check associated label for form elements
  if (['input', 'select', 'textarea'].includes(tagName)) {
    const id = element.getAttribute('id');
    if (id) {
      const label = document.querySelector(`label[for="${id}"]`);
      if (label?.textContent?.trim()) {
        return true;
      }
    }
    
    // Check if wrapped in label
    const parentLabel = element.closest('label');
    if (parentLabel?.textContent?.trim()) {
      return true;
    }
  }

  return false;
}

/**
 * Check if an element has proper focus management
 */
export function hasProperFocusManagement(element: Element): boolean {
  const tagName = element.tagName.toLowerCase();
  
  // Check if element is disabled
  if (element.hasAttribute('disabled') || element.getAttribute('aria-disabled') === 'true') {
    return false;
  }

  // Naturally focusable elements
  if (FOCUSABLE_ELEMENTS.has(tagName)) {
    return true;
  }

  // Elements with tabindex
  const tabIndex = element.getAttribute('tabindex');
  if (tabIndex !== null) {
    const tabIndexValue = parseInt(tabIndex, 10);
    return !isNaN(tabIndexValue) && tabIndexValue >= 0;
  }

  // Elements with interactive roles
  const role = element.getAttribute('role');
  if (role && ['button', 'link', 'menuitem', 'tab', 'option'].includes(role)) {
    return element.getAttribute('tabindex') !== null;
  }

  return false;
}

/**
 * Check if an element has keyboard navigation support
 */
export function hasKeyboardNavigation(element: Element): boolean {
  // Check for keyboard event handlers
  const hasKeyHandlers = ['onkeydown', 'onkeyup', 'onkeypress'].some(handler => 
    element.hasAttribute(handler) || (element as any)[handler]
  );

  if (hasKeyHandlers) {
    return true;
  }

  // Interactive roles should have keyboard support
  const role = element.getAttribute('role');
  if (role && ['button', 'link', 'menuitem', 'tab'].includes(role)) {
    return hasProperFocusManagement(element);
  }

  // Naturally interactive elements
  const tagName = element.tagName.toLowerCase();
  return ['button', 'a', 'input', 'select', 'textarea'].includes(tagName);
}

/**
 * Validate ARIA attributes on an element
 */
export function checkAriaAttributes(element: Element): AriaValidationResult {
  const errors: string[] = [];
  const attributes = element.attributes;

  for (let i = 0; i < attributes.length; i++) {
    const attr = attributes[i];
    
    if (attr.name.startsWith('aria-')) {
      const value = attr.value;
      
      // Check boolean ARIA attributes
      if (['aria-expanded', 'aria-selected', 'aria-checked', 'aria-disabled', 'aria-hidden'].includes(attr.name)) {
        if (!ARIA_BOOLEAN_VALUES.has(value)) {
          errors.push(`Invalid ${attr.name} value: ${value}`);
        }
      }
      
      // Check tristate ARIA attributes
      if (['aria-pressed'].includes(attr.name)) {
        if (!ARIA_TRISTATE_VALUES.has(value)) {
          errors.push(`Invalid ${attr.name} value: ${value}`);
        }
      }
    }
  }

  // Check for required attributes based on role
  const role = element.getAttribute('role');
  if (role === 'button' && !hasProperFocusManagement(element)) {
    errors.push('Interactive elements with role="button" must be focusable (add tabindex="0")');
  }

  return {
    isValid: errors.length === 0,
    errors
  };
}

/**
 * Check if an element uses semantic HTML
 */
export function isSemanticHTML(element: Element): boolean {
  const tagName = element.tagName.toLowerCase();
  return SEMANTIC_ELEMENTS.has(tagName);
}

/**
 * Get the accessible name of an element
 */
export function getAccessibleName(element: Element): string {
  // Check aria-label first
  const ariaLabel = element.getAttribute('aria-label');
  if (ariaLabel?.trim()) {
    return ariaLabel.trim();
  }

  // Check aria-labelledby
  const labelledBy = element.getAttribute('aria-labelledby');
  if (labelledBy) {
    const labelElement = document.getElementById(labelledBy);
    if (labelElement?.textContent?.trim()) {
      return labelElement.textContent.trim();
    }
  }

  // Check text content
  if (element.textContent?.trim()) {
    return element.textContent.trim();
  }

  // Check associated label for form elements
  const tagName = element.tagName.toLowerCase();
  if (['input', 'select', 'textarea'].includes(tagName)) {
    const id = element.getAttribute('id');
    if (id) {
      const label = document.querySelector(`label[for="${id}"]`);
      if (label?.textContent?.trim()) {
        return label.textContent.trim();
      }
    }
  }

  return '';
}

/**
 * Check if an element meets WCAG contrast requirements
 * Note: This is a simplified check - real contrast checking requires color analysis
 */
export function hasGoodContrast(element: Element): boolean {
  const styles = window.getComputedStyle(element);
  const backgroundColor = styles.backgroundColor;
  const color = styles.color;
  
  // This is a simplified check - in practice you'd need to calculate actual contrast ratios
  // For now, just check that both colors are defined and not transparent
  return backgroundColor !== 'rgba(0, 0, 0, 0)' && 
         backgroundColor !== 'transparent' && 
         color !== 'rgba(0, 0, 0, 0)' && 
         color !== 'transparent';
}
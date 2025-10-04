# Feedback Feature - Visual Guide

## Button Location

The "💡 Feature Idea" button is located in the **bottom-left corner** of the screen, next to the development nuke buttons:

```
┌─────────────────────────────────────────────────────┐
│                                                     │
│                   MAP AREA                          │
│                                                     │
│                                                     │
│                                                     │
│                                                     │
│                                                     │
│                                                     │
│  ┌──────────────────────────────────────────────┐  │
│  │ [🧹 Nuke POIs] [👥 Nuke Users] [💡 Feature Idea] │
│  │ Click avatar for video call • Right-click...  │  │
│  └──────────────────────────────────────────────┘  │
└─────────────────────────────────────────────────────┘
```

## Button Styling

- **Color**: Blue (`bg-blue-600 hover:bg-blue-700`)
- **Size**: Small (`px-3 py-1 text-xs`)
- **Icon**: 💡 (light bulb emoji)
- **Text**: "Feature Idea"
- **Tooltip**: "Share your ideas and feedback"

## Modal Interface

When clicked, a centered modal appears:

```
┌─────────────────────────────────────────────────────┐
│                                                     │
│         ┌───────────────────────────────┐          │
│         │ Share Your Idea           [X] │          │
│         ├───────────────────────────────┤          │
│         │                               │          │
│         │ Category                      │          │
│         │ [Feature Request ▼]           │          │
│         │                               │          │
│         │ Title                         │          │
│         │ [Brief description...]        │          │
│         │                               │          │
│         │ Description                   │          │
│         │ ┌───────────────────────────┐ │          │
│         │ │ Describe your idea...     │ │          │
│         │ │                           │ │          │
│         │ │                           │ │          │
│         │ └───────────────────────────┘ │          │
│         │ 0/1000 characters             │          │
│         │                               │          │
│         │ [Cancel]        [Submit]      │          │
│         └───────────────────────────────┘          │
│                                                     │
└─────────────────────────────────────────────────────┘
```

## Success State

After successful submission:

```
┌─────────────────────────────────────────────────────┐
│                                                     │
│         ┌───────────────────────────────┐          │
│         │ Share Your Idea           [X] │          │
│         ├───────────────────────────────┤          │
│         │                               │          │
│         │           ✓                   │          │
│         │      (large green)            │          │
│         │                               │          │
│         │  Thanks for your feedback!    │          │
│         │                               │          │
│         │ Your idea has been submitted  │          │
│         │    as a GitHub issue.         │          │
│         │                               │          │
│         └───────────────────────────────┘          │
│                                                     │
└─────────────────────────────────────────────────────┘
```

## Form Fields

### Category Dropdown
- **Feature Request** (default)
- **Bug Report**
- **Improvement**

### Title Input
- Placeholder: "Brief description of your idea"
- Min length: 5 characters
- Max length: 100 characters
- Required field

### Description Textarea
- Placeholder: "Describe your idea in detail..."
- Min length: 10 characters
- Max length: 1000 characters
- Character counter shown below
- Required field
- Non-resizable (fixed height)

## Validation States

### Valid State
- Submit button: Blue, enabled
- No error messages

### Invalid State (too short)
- Submit button: Gray, disabled
- Character counter turns red when below minimum

### Error State (submission failed)
```
┌─────────────────────────────────────────────────────┐
│ ⚠ Failed to submit feedback                        │
│   Please try again later                           │
└─────────────────────────────────────────────────────┘
```

## GitHub Issue Format

When submitted, creates an issue like:

```
Title: Add dark mode

## Description

It would be great to have a dark mode option for better 
visibility at night.

---

*Submitted via in-app feedback*

Labels: enhancement, user-feedback
```

## Color Scheme

- **Button**: Blue (#2563eb / #1d4ed8 on hover)
- **Modal Background**: White
- **Modal Overlay**: Black with 50% opacity
- **Success Checkmark**: Green (#10b981)
- **Error Message**: Red (#dc2626)
- **Text**: Gray scale (#111827, #4b5563, #6b7280)
- **Borders**: Light gray (#d1d5db)
- **Focus Ring**: Blue (#3b82f6)

## Responsive Design

- Modal width: `max-w-md` (28rem / 448px)
- Padding: 1.5rem (24px)
- Margin: 1rem (16px) on mobile
- Centered on screen
- Scrollable if content exceeds viewport

## Accessibility

- ✅ Keyboard navigation (Tab, Enter, Escape)
- ✅ Focus indicators on all interactive elements
- ✅ ARIA labels for screen readers
- ✅ Disabled state clearly indicated
- ✅ Error messages announced
- ✅ Success state announced

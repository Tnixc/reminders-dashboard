# UI Components

This directory contains reusable UI components for the reminder application. The components are built using [Charm's Lipgloss](https://github.com/charmbracelet/lipgloss) library for terminal styling.

## Architecture

The UI is built using a component-based architecture that separates concerns and makes the codebase more maintainable and reusable.

```
┌─────────────────────────────────────────────────────┐
│                    Layout                           │
│  ┌──────────────────────┐  ┌───────────────────┐   │
│  │                      │  │   Sidebar         │   │
│  │  Main Content        │  │  ┌─────────────┐  │   │
│  │  ┌────────────────┐  │  │  │ Settings    │  │   │
│  │  │ ColumnView or  │  │  │  │ Button      │  │   │
│  │  │ ListView       │  │  │  └─────────────┘  │   │
│  │  │                │  │  │                   │   │
│  │  │  ┌──────────┐  │  │  │  ┌─────────────┐  │   │
│  │  │  │  Card    │  │  │  │  │  Calendar   │  │   │
│  │  │  └──────────┘  │  │  │  │             │  │   │
│  │  └────────────────┘  │  │  └─────────────┘  │   │
│  └──────────────────────┘  └───────────────────┘   │
│  ┌──────────────────────────────────────────────┐  │
│  │              Footer                          │  │
│  └──────────────────────────────────────────────┘  │
└─────────────────────────────────────────────────────┘
```

## Components

### 1. Card (`card.go`)

Renders individual reminder items as cards.

**Key Functions:**
- `RenderCard(reminder, style, countdown, urgency)` - Renders a single reminder card
- `RenderEmptyCard(width, height, message)` - Renders placeholder card
- `DefaultCardStyle()` - Returns default card styling configuration

**Usage:**
```go
style := DefaultCardStyle()
style.Width = 30
style.Focused = true
style.ListColor = "205"

card := RenderCard(reminder, style, "in 2 days", 2)
```

**Features:**
- Configurable width and height
- Focus highlighting with border color change
- List color integration
- Truncation for long text
- Countdown and urgency display

### 2. Column View (`column_view.go`)

Renders multiple columns of reminder cards in a grid layout.

**Key Functions:**
- `RenderColumnView(config)` - Renders the complete column view
- `renderSingleColumn()` - Renders a single column (internal)

**Configuration:**
```go
type ColumnViewConfig struct {
    Width          int
    Height         int
    Columns        [][]Reminder
    ColumnNames    []string
    FocusedColumn  int
    FocusedItem    int
    ListColors     map[string]string
    ScrollOffsets  []int
    GetCountdown   func(string) (string, int)
}
```

**Features:**
- Dynamic column width calculation
- Responsive layout (adapts to terminal size)
- Scroll support per column
- Focus highlighting
- Overflow indicator for hidden columns

### 3. List View (`list_view.go`)

Renders reminders in a compact list format.

**Key Functions:**
- `RenderListView(config)` - Standard list view with details
- `RenderCompactList(config)` - Compact one-line per item view

**Configuration:**
```go
type ListViewConfig struct {
    Width         int
    Height        int
    Reminders     []Reminder
    FocusedIndex  int
    ScrollOffset  int
    ListColors    map[string]string
    GetCountdown  func(string) (string, int)
    FormatDueDate func(string) string
}
```

**Features:**
- Two view modes: detailed and compact
- Scroll support
- Focus highlighting
- Color-coded lists
- Urgency-based countdown coloring

**Format (Compact):**
```
Item 1  list 1  in 1 day  due dd/mm
Item 2  list 2  in 2 days  due dd/mm
```

### 4. Settings Panel (`settings.go`)

Renders the settings overlay/panel with configuration options.

**Key Functions:**
- `RenderSettingsPanel(config)` - Main settings panel
- `RenderSettingsButton(width, height, focused)` - Settings trigger button
- `renderDaysFilterSettings()` - Days filter section (internal)
- `renderListFilterSettings()` - List visibility section (internal)
- `renderColorSettings()` - Color configuration section (internal)

**Configuration:**
```go
type SettingsConfig struct {
    Width            int
    Height           int
    Visible          bool
    DaysFilter       int
    DaysOptions      []int
    SelectedLists    map[string]bool
    AvailableLists   []string
    ListColors       map[string]string
    AvailableColors  []string
    ColorNames       []string
    FocusedSection   int
    CursorPosition   int
}
```

**Features:**
- Three sections: days filter, list filter, color config
- Visual focus indicators
- Color palette picker (1-9, 0)
- Checkbox toggles for lists
- Radio buttons for days filter

### 5. Calendar (`calendar.go`)

Renders a small calendar widget.

**Key Functions:**
- `RenderCalendar(config)` - Full calendar with month view
- `RenderMiniCalendar(width, height, focused)` - Minimal calendar

**Configuration:**
```go
type CalendarConfig struct {
    Width       int
    Height      int
    CurrentDate time.Time
    Focused     bool
}
```

**Features:**
- Month and year display
- Current day highlighting
- Proper weekday alignment
- Minimal and full view modes

### 6. Footer (`footer.go`)

Renders help text and navigation hints.

**Key Functions:**
- `RenderFooter(config)` - Renders footer bar
- `GetHelpText(mode, settingsVisible)` - Gets context-appropriate help text

**Features:**
- Context-aware help text
- Full-width styling
- Distinct background color

### 7. Layout (`layout.go`)

Combines all components into the final layout.

**Key Functions:**
- `RenderLayout(config)` - Main layout renderer
- `renderLayoutWithSidebar()` - Normal mode (internal)
- `renderLayoutWithSettingsOverlay()` - Settings mode (internal)
- `buildSidebar()` - Sidebar construction (internal)

**Features:**
- Two-panel layout (main + sidebar)
- Settings overlay mode
- Responsive width calculations
- Footer integration

## Design Patterns

### 1. Configuration Objects

All components use configuration structs for parameters:
- Makes function signatures cleaner
- Easy to add new options
- Self-documenting code

### 2. Separation of Concerns

- **Components** handle rendering only
- **Business logic** stays in main package
- **Data transformation** via callbacks (e.g., `GetCountdown`)

### 3. Composability

Components can be:
- Used independently
- Nested within other components
- Styled externally

### 4. Responsiveness

All components handle:
- Dynamic width/height
- Overflow gracefully
- Terminal resize events

## Color Palette

The components use a consistent color palette:

```go
availableColors := []string{
    "205", // Pink (focus/accent)
    "141", // Purple (headers)
    "81",  // Blue
    "51",  // Cyan
    "48",  // Teal
    "118", // Green
    "226", // Yellow (warning)
    "208", // Orange (urgent)
    "196", // Red (overdue)
    "248", // Gray (normal)
    "245", // Dark gray (meta)
    "241", // Darker gray (footer)
    "240", // Border gray
    "235", // Background
}
```

## Usage Example

Here's how the main application uses these components:

```go
// In views.go
func (m model) View() string {
    // Render main content
    mainContent := m.renderColumnView(width, height)
    
    // Render sidebar
    sidebar := m.renderSidebar(sidebarWidth, height)
    
    // Combine with layout
    content := lipgloss.JoinHorizontal(
        lipgloss.Top,
        mainContent,
        sidebar,
    )
    
    // Add footer
    footer := m.renderFooter()
    return lipgloss.JoinVertical(lipgloss.Left, content, footer)
}
```

## Future Enhancements

Potential improvements:

1. **Animation Support**
   - Smooth transitions between views
   - Loading spinners

2. **Theming**
   - Switchable color schemes
   - Dark/light mode

3. **More Views**
   - Timeline view
   - Priority-based view
   - Calendar integration view

4. **Interactive Widgets**
   - Date picker
   - Time input
   - Multi-select lists

5. **Accessibility**
   - High contrast mode
   - Screen reader hints
   - Keyboard shortcuts help

## Testing

Each component should be testable independently:

```go
func TestRenderCard(t *testing.T) {
    reminder := Reminder{
        Title: "Test",
        DueDate: "2024-01-01T00:00:00Z",
    }
    style := DefaultCardStyle()
    result := RenderCard(reminder, style, "in 1 day", 1)
    
    // Assert result contains expected elements
}
```

## Contributing

When adding new components:

1. Follow the configuration object pattern
2. Add comprehensive documentation
3. Include usage examples
4. Handle edge cases (empty data, small screens)
5. Use consistent styling
6. Make components responsive
# Theme Configuration Examples

This directory contains example configuration files for customizing the appearance of your reminders dashboard.

## Quick Start

1. Copy `config-with-theme.toml` to your config directory:
   ```bash
   # On macOS/Linux
   cp config-with-theme.toml ~/.config/reminders-dashboard/config.toml
   
   # On Windows
   copy config-with-theme.toml %APPDATA%\reminders-dashboard\config.toml
   ```

2. Edit the file to customize colors to your preference.

3. Restart the dashboard to see your changes.

## Theme Structure

The theme is organized into three main sections:

### Text Colors

Control the appearance of text throughout the application:

- **primary**: Main text color, used for titles when selected/hovered, important content
- **secondary**: Secondary text color for less critical information
- **faint**: Subtle text for placeholders, disabled items, and very low-priority content
- **inverted**: Text color used on highlighted/selected backgrounds
- **success**: Green color for completed items and positive states
- **warning**: Orange/yellow color for moderate urgency (items due within 48 hours)
- **error**: Red color for high urgency (overdue or due within 24 hours)

### Background Colors

- **selected**: Background color when hovering over or selecting items in tables

### Border Colors

- **primary**: Main borders between sections and major UI elements
- **secondary**: Less prominent borders and dividers
- **faint**: Very subtle borders for minimal separation

## Urgency Color System

The dashboard automatically colors reminder due dates based on urgency:

- **🔴 Red** (error): Overdue or due within 24 hours
- **🟠 Orange** (warning): Due within 48 hours
- **🟡 Yellow**: Due within 1 week (uses a yellow shade between warning and grey)
- **⚪ Grey** (faint): Due later than 1 week

This gradient helps you quickly identify which items need immediate attention.

## Color Format

Colors can be specified in two formats:

1. **Hex colors**: `"#c6d0f5"` - Full 6-digit hex color codes
2. **Terminal colors**: `"001"` - ANSI terminal color codes (0-255)

Hex colors are recommended for consistency across different terminals.

## Included Color Palettes

### Catppuccin Frappe (Default)

A soothing dark theme with pastel colors. Perfect for extended use and easy on the eyes.

**Base colors:**
- Darkest: `#232634` (Crust)
- Background: `#303446` (Base)
- Surfaces: `#414559`, `#51576d`, `#626880`

**Text colors:**
- Primary: `#c6d0f5` (Text)
- Secondary: `#b5bfe2` (Subtext 1)
- Faint: `#a5adce` (Subtext 0)

**Accent colors:**
- Success/Green: `#a6d189`
- Warning/Peach: `#ef9f76`
- Error/Red: `#e78284`

See the full palette in `config-with-theme.toml` for additional accent colors you can use for list colors.

## List Colors

Customize colors for specific reminder lists to help differentiate them visually:

```toml
[listColors]
"Work" = "#e78284"      # Red for work items
"Personal" = "#a6d189"  # Green for personal items
"Shopping" = "#ef9f76"  # Orange for shopping lists
"Ideas" = "#ca9ee6"     # Purple for ideas
```

## Tips

1. **Use high contrast**: Ensure text colors contrast well with your terminal background
2. **Test visibility**: Check that urgent reminders (red/orange) stand out clearly
3. **Consistent palette**: Use colors from the same palette for a cohesive look
4. **Terminal compatibility**: If using hex colors, make sure your terminal supports true color

## Creating Your Own Theme

1. Start with `config-with-theme.toml` as a template
2. Choose a color palette you like (e.g., from [terminal.sexy](https://terminal.sexy/) or [Catppuccin](https://github.com/catppuccin/catppuccin))
3. Map the palette colors to the theme sections:
   - Text: Use 3-4 shades from light to dark
   - Backgrounds: Use subtle, darker colors
   - Borders: Use colors slightly lighter than backgrounds
   - Status: Use distinct colors (green/orange/red) that stand out

## Troubleshooting

**Colors not showing up?**
- Verify your config file is in the correct location
- Check that hex colors start with `#`
- Ensure your terminal supports true color (most modern terminals do)

**Colors look wrong?**
- Your terminal's color scheme might be overriding the colors
- Try using hex colors instead of terminal color codes
- Disable your terminal's color theme or use a neutral one

**Theme not loading?**
- Check for syntax errors in your TOML file
- Ensure all required fields are present
- Look for error messages when starting the dashboard

## Resources

- [Catppuccin](https://github.com/catppuccin/catppuccin) - Soothing pastel theme
- [TOML Specification](https://toml.io/) - Configuration file format
- [Terminal Color Codes](https://en.wikipedia.org/wiki/ANSI_escape_code#Colors) - ANSI color reference
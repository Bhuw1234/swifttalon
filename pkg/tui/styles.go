// SwiftTalon TUI - Dracula-inspired Dark Theme
// Clean, modern terminal aesthetic with Glamour markdown support

package tui

import (
	"strings"

	"github.com/charmbracelet/glamour"
	"github.com/charmbracelet/glamour/ansi"
	"github.com/charmbracelet/lipgloss"
)

// Dracula-inspired Color Palette
// Based on the design concept from swifttalon-tui.zip
var (
	// Primary colors - Dracula palette
	Cyan    = lipgloss.Color("#8be9fd") // Cyan - primary accent
	Green   = lipgloss.Color("#50fa7b") // Green - success/status
	Pink    = lipgloss.Color("#ff79c6") // Pink - highlights/selection
	Purple  = lipgloss.Color("#bd93f9") // Purple - secondary accent
	Orange  = lipgloss.Color("#ffb86c") // Orange - warnings
	Yellow  = lipgloss.Color("#f1fa8c") // Yellow - alerts
	Red     = lipgloss.Color("#ff5555") // Red - errors

	// Background colors - True black theme
	Black       = lipgloss.Color("#000000") // Pure black background
	DarkBg      = lipgloss.Color("#1e1e1e") // Dark background
	PanelBg     = lipgloss.Color("#252526") // Panel background
	SurfaceBg   = lipgloss.Color("#2d2d2d") // Surface
	BorderDark  = lipgloss.Color("#3c3c3c") // Border
	BorderLight = lipgloss.Color("#6272a4") // Light border (Dracula comment)

	// Text colors
	TextPrimary   = lipgloss.Color("#f8f8f2") // White-ish
	TextSecondary = lipgloss.Color("#cccccc") // Gray
	TextMuted     = lipgloss.Color("#858585") // Muted gray
)

// Typing animation frames
var TypingFrames = []string{"⠋", "⠙", "⠹", "⠸", "⠼", "⠴", "⠦", "⠧", "⠇", "⠏"}

// Core Styles
var (
	// Logo style with pink accent
	LogoStyle = lipgloss.NewStyle().
			Foreground(Pink).
			Bold(true).
			Padding(0, 1)

	// Title style with cyan
	TitleStyle = lipgloss.NewStyle().
			Foreground(Cyan).
			Bold(true).
			Padding(0, 1)

	// Subtitle style
	SubtitleStyle = lipgloss.NewStyle().
			Foreground(TextSecondary).
			Italic(true)

	// Glowing text effect
	GlowText = lipgloss.NewStyle().
			Foreground(Green).
			Bold(true)
)

// Panel Styles
var (
	// Main panel with dark background
	MainPanel = lipgloss.NewStyle().
			Background(PanelBg).
			Border(lipgloss.RoundedBorder()).
			BorderForeground(BorderDark).
			Padding(1, 2)

	// Content panel for viewport
	ContentPanel = lipgloss.NewStyle().
			Background(DarkBg).
			Border(lipgloss.NormalBorder()).
			BorderForeground(BorderDark).
			Padding(1, 1)

	// Input box with green border
	InputBox = lipgloss.NewStyle().
			Background(SurfaceBg).
			Foreground(TextPrimary).
			Border(lipgloss.NormalBorder()).
			BorderForeground(Green).
			Padding(0, 1).
			Width(60)
)

// Status Styles
var (
	// Online status - green
	StatusOnline = lipgloss.NewStyle().
			Foreground(Green).
			Bold(true)

	// Processing status - orange
	StatusProcessing = lipgloss.NewStyle().
				Foreground(Orange).
				Bold(true)

	// Error status - red
	StatusError = lipgloss.NewStyle().
			Foreground(Red).
			Bold(true)

	// Model selector with cyan
	ModelSelector = lipgloss.NewStyle().
			Background(SurfaceBg).
			Foreground(Cyan).
			Border(lipgloss.RoundedBorder()).
			BorderForeground(Cyan).
			Padding(0, 2).
			Bold(true)

	// Session active indicator
	SessionActive = lipgloss.NewStyle().
			Foreground(Yellow).
			Bold(true).
			PaddingLeft(2)
)

// Message Styles
var (
	// User message - cyan
	UserStyle = lipgloss.NewStyle().
			Foreground(Cyan).
			Bold(true)

	// Assistant message - green
	AssistantStyle = lipgloss.NewStyle().
			Foreground(Green).
			Bold(true)

	// System message - pink
	SystemStyle = lipgloss.NewStyle().
			Foreground(Pink).
			Bold(true)
)

// Menu Styles (from design concept)
var (
	// Menu container
	MenuPanel = lipgloss.NewStyle().
			Background(PanelBg).
			Border(lipgloss.RoundedBorder()).
			BorderForeground(BorderDark).
			Padding(1, 2).
			Width(50)

	// Selected menu item
	MenuItemSelected = lipgloss.NewStyle().
				Background(SurfaceBg).
				Foreground(TextPrimary).
				BorderLeft(true).
				BorderForeground(Pink).
				Padding(0, 1).
				Bold(true)

	// Unselected menu item
	MenuItem = lipgloss.NewStyle().
			Foreground(TextSecondary).
			Padding(0, 1)

	// Selection indicator
	SelectionIndicator = lipgloss.NewStyle().
				Foreground(Pink).
				Bold(true)
)

// Help text style
var HelpStyle = lipgloss.NewStyle().
		Foreground(TextMuted).
		Italic(true)

// GetTypingFrame returns the current typing animation frame
func GetTypingFrame(frame int) string {
	return TypingFrames[frame%len(TypingFrames)]
}

// GetGlamourRenderer creates a glamour renderer with dark theme
func GetGlamourRenderer(width int) (*glamour.TermRenderer, error) {
	return glamour.NewTermRenderer(
		glamour.WithStandardStyle("dracula"),
		glamour.WithWordWrap(width),
	)
}

// GetGlamourRendererFromStyle creates a glamour renderer with custom style
func GetGlamourRendererFromStyle(width int) (*glamour.TermRenderer, error) {
	// Custom Dracula-inspired dark theme for glamour
	darkStyleConfig := ansi.StyleConfig{
		Document: ansi.StyleBlock{
			StylePrimitive: ansi.StylePrimitive{
				BlockPrefix: "\n",
				BlockSuffix: "\n",
				Color:       stringPtr("#f8f8f2"),
				BackgroundColor: stringPtr("#1e1e1e"),
			},
			Margin: uintPtr(2),
		},
		BlockQuote: ansi.StyleBlock{
			StylePrimitive: ansi.StylePrimitive{
				Italic: boolPtr(true),
				Color:  stringPtr("#858585"),
			},
			Indent:      uintPtr(1),
			IndentToken: stringPtr("│ "),
		},
		List: ansi.StyleList{
			StyleBlock: ansi.StyleBlock{
				StylePrimitive: ansi.StylePrimitive{
					Color: stringPtr("#f8f8f2"),
				},
				Indent: uintPtr(2),
			},
		},
		Heading: ansi.StyleBlock{
			StylePrimitive: ansi.StylePrimitive{
				Bold:        boolPtr(true),
				Color:       stringPtr("#ff79c6"),
				BlockSuffix: "\n",
			},
		},
		H1: ansi.StyleBlock{
			StylePrimitive: ansi.StylePrimitive{
				Bold:        boolPtr(true),
				Color:       stringPtr("#8be9fd"),
				BackgroundColor: stringPtr("#2d2d2d"),
			},
		},
		H2: ansi.StyleBlock{
			StylePrimitive: ansi.StylePrimitive{
				Bold:  boolPtr(true),
				Color: stringPtr("#50fa7b"),
			},
		},
		H3: ansi.StyleBlock{
			StylePrimitive: ansi.StylePrimitive{
				Bold:  boolPtr(true),
				Color: stringPtr("#ffb86c"),
			},
		},
		Code: ansi.StyleBlock{
			StylePrimitive: ansi.StylePrimitive{
				Color:  stringPtr("#ff79c6"),
				Italic: boolPtr(true),
			},
		},
		CodeBlock: ansi.StyleCodeBlock{
			StyleBlock: ansi.StyleBlock{
				StylePrimitive: ansi.StylePrimitive{
					Color:          stringPtr("#f8f8f2"),
					BackgroundColor: stringPtr("#252526"),
				},
				Indent: uintPtr(2),
			},
		},
		Emph: ansi.StylePrimitive{
			Italic: boolPtr(true),
			Color:  stringPtr("#ffb86c"),
		},
		Strong: ansi.StylePrimitive{
			Bold:  boolPtr(true),
			Color: stringPtr("#50fa7b"),
		},
		Link: ansi.StylePrimitive{
			Color:     stringPtr("#8be9fd"),
			Underline: boolPtr(true),
		},
	}

	return glamour.NewTermRenderer(
		glamour.WithStyles(darkStyleConfig),
		glamour.WithWordWrap(width),
	)
}

// Helper functions for pointer types
func stringPtr(s string) *string       { return &s }
func boolPtr(b bool) *bool             { return &b }
func uintPtr(u uint) *uint             { return &u }

// RenderTacticalHeader creates a tactical-style header (kept for compatibility)
func RenderTacticalHeader(title string, width int) string {
	left := lipgloss.NewStyle().
		Foreground(Pink).
		Render("◆")

	right := lipgloss.NewStyle().
		Foreground(Pink).
		Render("◆")

	titleStyle := lipgloss.NewStyle().
		Foreground(Orange).
		Bold(true).
		Render(title)

	line := lipgloss.NewStyle().
		Foreground(BorderDark).
		Render(strings.Repeat("─", (width-len(title)-6)/2))

	return line + " " + left + " " + titleStyle + " " + right + " " + line
}

// RenderNeonText creates a neon glow effect (kept for compatibility)
func RenderNeonText(text string, color lipgloss.Color) string {
	return lipgloss.NewStyle().
		Foreground(color).
		Bold(true).
		Render(text)
}

// SwiftTalon TUI - Styles
// AGGRESSIVE and SUPER EASY TO USE
// Package tui provides styling for the terminal interface

package tui

import (
	"github.com/charmbracelet/lipgloss"
)

// Color palette - AGGRESSIVE vibrant theme
var (
	// Primary colors - BOLD and vibrant
	colorPrimary    = lipgloss.Color("#FF00FF") // MAGENTA - attention grabbing
	colorSecondary  = lipgloss.Color("#00FFFF") // CYAN - high contrast
	colorAccent     = lipgloss.Color("#FFFF00") // YELLOW - punchy
	colorSuccess    = lipgloss.Color("#00FF00") // BRIGHT GREEN
	colorError      = lipgloss.Color("#FF3333") // BRIGHT RED
	colorWarning    = lipgloss.Color("#FF9500") // ORANGE

	// Background colors - DARK for contrast
	colorBgPrimary   = lipgloss.Color("#0A0A0A") // Near black
	colorBgSecondary = lipgloss.Color("#141414") // Slightly lighter
	colorBgTertiary  = lipgloss.Color("#1E1E1E") // Panel background
	colorBgHover     = lipgloss.Color("#2A2A2A") // Hover state
	colorBgInput     = lipgloss.Color("#0F0F0F") // Input background

	// Text colors - HIGH CONTRAST
	colorTextPrimary   = lipgloss.Color("#FFFFFF") // PURE WHITE
	colorTextSecondary = lipgloss.Color("#CCCCCC") // Light gray
	colorTextMuted     = lipgloss.Color("#888888") // Muted

	// Border colors - BOLD
	colorBorder       = lipgloss.Color("#333333")
	colorBorderActive = lipgloss.Color("#FF00FF") // MAGENTA for active
	colorBorderInput  = lipgloss.Color("#00FFFF") // CYAN for input focus

	// Special colors
	colorUserBg    = lipgloss.Color("#00AAFF") // Bright blue for user
	colorAgentBg   = lipgloss.Color("#AA00FF") // Purple for agent
	colorHighlight = lipgloss.Color("#FFD700") // Gold for highlights
)

// Base styles
var (
	// Base style for the entire app
	BaseStyle = lipgloss.NewStyle().
			Background(colorBgPrimary)

	// Title style - BOLD and BIG
	TitleStyle = lipgloss.NewStyle().
			Foreground(colorPrimary).
			Background(colorBgSecondary).
			Bold(true).
			Padding(0, 2).
			Margin(0, 1)

	// Subtitle style
	SubtitleStyle = lipgloss.NewStyle().
			Foreground(colorSecondary).
			Italic(true).
			Bold(true)
)

// Sidebar styles - COMPACT
var (
	// Sidebar container
	SidebarStyle = lipgloss.NewStyle().
			Background(colorBgSecondary).
			BorderRight(true).
			BorderForeground(colorBorder).
			Padding(1, 1).
			Width(26)

	// Session item - subdued
	SessionItemStyle = lipgloss.NewStyle().
				Padding(0, 1).
				Foreground(colorTextSecondary)

	// Active session - BRIGHT and OBVIOUS
	SessionActiveStyle = lipgloss.NewStyle().
				Padding(0, 1).
				Foreground(colorPrimary).
				Background(colorBgHover).
				Bold(true).
				BorderLeft(true).
				BorderForeground(colorPrimary)

	// Session title
	SessionTitleStyle = lipgloss.NewStyle().
				Foreground(colorTextPrimary).
				Width(20).
				Bold(false)

	// Session meta info
	SessionMetaStyle = lipgloss.NewStyle().
				Foreground(colorTextMuted).
				Faint(true)
)

// Chat styles - FOCUS ON CONTENT
var (
	// Chat container
	ChatStyle = lipgloss.NewStyle().
			Background(colorBgPrimary).
			Padding(1, 2)

	// User message - BRIGHT
	UserMessageStyle = lipgloss.NewStyle().
				Foreground(colorTextPrimary).
				Background(lipgloss.Color("#0A1A2A")).
				Padding(0, 2).
				BorderLeft(true).
				BorderForeground(colorUserBg).
				MarginLeft(1)

	// Assistant message
	AssistantMessageStyle = lipgloss.NewStyle().
				Foreground(colorTextPrimary).
				Background(lipgloss.Color("#1A0A2A")).
				Padding(0, 2).
				BorderLeft(true).
				BorderForeground(colorAgentBg).
				MarginLeft(1)

	// System message
	SystemMessageStyle = lipgloss.NewStyle().
				Foreground(colorTextMuted).
				Italic(true).
				Padding(0, 1).
				MarginLeft(2)

	// Timestamp style
	TimestampStyle = lipgloss.NewStyle().
			Foreground(colorTextMuted).
			Faint(true)

	// Role badge - BOLD
	RoleUserStyle = lipgloss.NewStyle().
			Foreground(colorBgPrimary).
			Background(colorUserBg).
			Padding(0, 2).
			Bold(true).
			MarginRight(1)

	RoleAssistantStyle = lipgloss.NewStyle().
				Foreground(colorBgPrimary).
				Background(colorAgentBg).
				Padding(0, 2).
				Bold(true).
				MarginRight(1)

	RoleSystemStyle = lipgloss.NewStyle().
				Foreground(colorBgPrimary).
				Background(colorTextMuted).
				Padding(0, 1).
				Italic(true)
)

// Input styles - PROMINENT!
var (
	// Input container - ALWAYS FOCUSED FEEL
	InputStyle = lipgloss.NewStyle().
			Background(colorBgInput).
			BorderTop(true).
			BorderForeground(colorBorder).
			Padding(1, 2).
			MarginTop(1)

	// Input box when focused - OBVIOUS
	InputFocusedStyle = lipgloss.NewStyle().
				Background(colorBgInput).
				BorderTop(true).
				BorderForeground(colorBorderInput).
				Padding(1, 2).
				MarginTop(1)

	// Placeholder style
	PlaceholderStyle = lipgloss.NewStyle().
				Foreground(colorTextMuted).
				Italic(true)

	// Cursor style
	CursorStyle = lipgloss.NewStyle().
			Foreground(colorPrimary).
			Bold(true)

	// Prompt indicator - BRIGHT
	PromptStyle = lipgloss.NewStyle().
			Foreground(colorSecondary).
			Background(colorBgSecondary).
			Bold(true).
			Padding(0, 1)
)

// Status bar styles - ALWAYS VISIBLE SHORTCUTS
var (
	// Status bar container
	StatusBarStyle = lipgloss.NewStyle().
			Background(colorBgTertiary).
			Foreground(colorTextPrimary).
			Padding(0, 1).
			Height(1)

	// Status indicator (connected, etc.)
	StatusConnectedStyle = lipgloss.NewStyle().
				Foreground(colorSuccess).
				Bold(true)

	StatusDisconnectedStyle = lipgloss.NewStyle().
				Foreground(colorError).
				Bold(true)

	StatusLoadingStyle = lipgloss.NewStyle().
				Foreground(colorAccent).
				Bold(true).
				Blink(true)

	// Model badge - OBVIOUS
	ModelBadgeStyle = lipgloss.NewStyle().
				Foreground(colorBgPrimary).
				Background(colorPrimary).
				Padding(0, 2).
				Bold(true)

	// Focus indicator
	FocusIndicatorStyle = lipgloss.NewStyle().
				Foreground(colorBgPrimary).
				Background(colorSecondary).
				Padding(0, 1).
				Bold(true)

	// Provider badge
	ProviderBadgeStyle = lipgloss.NewStyle().
				Foreground(colorBgPrimary).
				Background(colorSecondary).
				Padding(0, 1)
)

// Modal styles
var (
	// Modal container - BOLD borders
	ModalStyle = lipgloss.NewStyle().
			Background(colorBgTertiary).
			Border(lipgloss.DoubleBorder()).
			BorderForeground(colorPrimary).
			Padding(1, 3).
			Margin(2, 4)

	// Modal title - BIG
	ModalTitleStyle = lipgloss.NewStyle().
				Foreground(colorPrimary).
				Background(colorBgSecondary).
				Bold(true).
				Padding(0, 2).
				MarginBottom(1)

	// Model selector item
	ModelItemStyle = lipgloss.NewStyle().
			Padding(0, 2).
			Foreground(colorTextPrimary)

	ModelItemSelectedStyle = lipgloss.NewStyle().
				Padding(0, 2).
				Foreground(colorBgPrimary).
				Background(colorPrimary).
				Bold(true)
)

// Help styles - KEYBOARD SHORTCUTS ALWAYS VISIBLE
var (
	// Help panel
	HelpStyle = lipgloss.NewStyle().
			Background(colorBgSecondary).
			Border(lipgloss.DoubleBorder()).
			BorderForeground(colorPrimary).
			Padding(1, 2).
			Margin(1, 2)

	// Key binding - OBVIOUS
	KeyStyle = lipgloss.NewStyle().
			Foreground(colorBgPrimary).
			Background(colorSecondary).
			Padding(0, 1).
			Bold(true)

	// Key description
	KeyDescStyle = lipgloss.NewStyle().
			Foreground(colorTextSecondary).
			Padding(0, 1)
)

// Error styles - IN YOUR FACE
var (
	// Error message - OBVIOUS
	ErrorStyle = lipgloss.NewStyle().
			Foreground(colorTextPrimary).
			Background(colorError).
			Padding(1, 2).
			Border(lipgloss.DoubleBorder()).
			BorderForeground(colorError).
			Margin(1, 2).
			Bold(true)

	// Success message
	SuccessStyle = lipgloss.NewStyle().
			Foreground(colorBgPrimary).
			Background(colorSuccess).
			Padding(1, 2).
			Bold(true)
)

// Shortcut bar styles - ALWAYS VISIBLE
var (
	// Shortcut bar at bottom
	ShortcutBarStyle = lipgloss.NewStyle().
				Background(colorBgSecondary).
				Foreground(colorTextSecondary).
				Padding(0, 1).
				Height(1)

	// Individual shortcut key
	ShortcutKeyStyle = lipgloss.NewStyle().
				Foreground(colorBgPrimary).
				Background(colorTextMuted).
				Padding(0, 1).
				Bold(true)

	// Shortcut description
	ShortcutDescStyle = lipgloss.NewStyle().
				Foreground(colorTextMuted)
)

// Welcome styles
var (
	WelcomeTitleStyle = lipgloss.NewStyle().
				Foreground(colorPrimary).
				Background(colorBgSecondary).
				Bold(true).
				Padding(1, 3).
				MarginBottom(1)

	WelcomeTextStyle = lipgloss.NewStyle().
				Foreground(colorTextSecondary).
				Padding(0, 1)
)

// Typing indicator styles
var (
	TypingIndicatorStyle = lipgloss.NewStyle().
				Foreground(colorAccent).
				Bold(true).
				Italic(true)
)

// Minimum terminal size
const (
	MinTerminalWidth  = 80
	MinTerminalHeight = 24
)

// Helper functions

// Width returns a style with fixed width
func Width(w int) lipgloss.Style {
	return lipgloss.NewStyle().Width(w)
}

// Height returns a style with fixed height
func Height(h int) lipgloss.Style {
	return lipgloss.NewStyle().Height(h)
}

// Border returns a bordered style
func Border(color lipgloss.Color) lipgloss.Style {
	return lipgloss.NewStyle().
		Border(lipgloss.NormalBorder()).
		BorderForeground(color)
}

// Padding returns a style with padding
func Padding(top, right, bottom, left int) lipgloss.Style {
	return lipgloss.NewStyle().Padding(top, right, bottom, left)
}

// Margin returns a style with margin
func Margin(top, right, bottom, left int) lipgloss.Style {
	return lipgloss.NewStyle().Margin(top, right, bottom, left)
}

// JoinHorizontal joins strings horizontally with separator
func JoinHorizontal(sep string, strs ...string) string {
	return lipgloss.JoinHorizontal(lipgloss.Top, strs...)
}

// JoinVertical joins strings vertically with separator
func JoinVertical(sep string, strs ...string) string {
	return lipgloss.JoinVertical(lipgloss.Left, strs...)
}

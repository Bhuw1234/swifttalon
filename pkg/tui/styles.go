// SwiftTalon TUI - RAMBO Style 3D Interface
// Aggressive, bold, tactical design

package tui

import (
	"github.com/charmbracelet/lipgloss"
)

// RAMBO Color Palette - Tactical, aggressive, bold
var (
	// Primary colors - Military inspired
	RamboRed     = lipgloss.Color("#FF2D2D")    // Aggressive red
	RamboOrange  = lipgloss.Color("#FF6B2D")    // Warning orange
	RamboGreen   = lipgloss.Color("#2DFF6B")    // Tactical green
	RamboYellow  = lipgloss.Color("#FFE42D")    // Alert yellow
	RamboCyan    = lipgloss.Color("#2DFFF0")    // Tech cyan
	
	// Background colors - Dark tactical
	DarkBase     = lipgloss.Color("#0A0A0F")    // Near black
	DarkPanel    = lipgloss.Color("#12121A")    // Panel background
	DarkSurface  = lipgloss.Color("#1A1A25")    // Surface
	DarkBorder   = lipgloss.Color("#2A2A3A")    // Border
	
	// Accent colors
	NeonPink     = lipgloss.Color("#FF2D7A")    // Neon accent
	ElectricBlue = lipgloss.Color("#2D7AFF")    // Electric blue
	PlasmaPurple = lipgloss.Color("#9D2DFF")    // Plasma purple
	
	// Text colors
	TextPrimary   = lipgloss.Color("#FFFFFF")
	TextSecondary = lipgloss.Color("#AAAAAA")
	TextMuted     = lipgloss.Color("#666666")
)

// 3D Effect Styles
var (
	// 3D Panel with depth effect
	Panel3D = lipgloss.NewStyle().
		Background(DarkPanel).
		Border(lipgloss.RoundedBorder()).
		BorderBackground(DarkBase).
		BorderForeground(DarkBorder).
		Padding(1, 2)

	// 3D Box with shadow effect
	Box3D = lipgloss.NewStyle().
		Background(DarkSurface).
		Border(lipgloss.NormalBorder()).
		BorderForeground(lipgloss.Color("#3A3A4A")).
		Padding(0, 1)

	// Neon border effect
	NeonBorder = lipgloss.NewStyle().
		Border(lipgloss.ThickBorder()).
		BorderForeground(RamboRed).
		BorderBackground(DarkBase).
		Padding(1, 2)

	// Glowing text effect
	GlowText = lipgloss.NewStyle().
		Foreground(RamboRed).
		Bold(true)

	// Title style with 3D effect
	Title3D = lipgloss.NewStyle().
		Foreground(RamboOrange).
		Background(DarkBase).
		Bold(true).
		Underline(true).
		Padding(0, 1)

	// Subtitle style
	SubtitleStyle = lipgloss.NewStyle().
		Foreground(TextSecondary).
		Italic(true)

	// Input box with tactical styling
	InputBox = lipgloss.NewStyle().
		Background(DarkSurface).
		Border(lipgloss.NormalBorder()).
		BorderForeground(RamboGreen).
		Padding(0, 1).
		Width(60)

	// Response panel
	ResponsePanel = lipgloss.NewStyle().
		Background(DarkPanel).
		Border(lipgloss.DoubleBorder()).
		BorderForeground(ElectricBlue).
		Padding(1, 2).
		Width(80)

	// Status indicator styles
	StatusOnline = lipgloss.NewStyle().
		Foreground(RamboGreen).
		Bold(true)

	StatusProcessing = lipgloss.NewStyle().
		Foreground(RamboOrange).
		Bold(true)

	StatusError = lipgloss.NewStyle().
		Foreground(RamboRed).
		Bold(true)

	// Model selector
	ModelSelector = lipgloss.NewStyle().
		Background(DarkSurface).
		Foreground(RamboCyan).
		Border(lipgloss.RoundedBorder()).
		BorderForeground(RamboCyan).
		Padding(0, 2).
		Bold(true)

	// Session style
	SessionStyle = lipgloss.NewStyle().
		Foreground(TextSecondary).
		PaddingLeft(2)

	SessionActive = lipgloss.NewStyle().
		Foreground(RamboYellow).
		Bold(true).
		PaddingLeft(2)

	// Help text style
	HelpStyle = lipgloss.NewStyle().
		Foreground(TextMuted).
		Italic(true)

	// Typing indicator animation frames
	TypingFrames = []string{"⠋", "⠙", "⠹", "⠸", "⠼", "⠴", "⠦", "⠧", "⠇", "⠏"}

	// 3D Bar for progress/loading
	Bar3D = lipgloss.NewStyle().
		Foreground(RamboGreen).
		Background(DarkSurface)

	// Logo style
	LogoStyle = lipgloss.NewStyle().
		Foreground(RamboRed).
		Bold(true).
		Padding(0, 1)

	// User message style
	UserStyle = lipgloss.NewStyle().
		Foreground(RamboCyan).
		Bold(true)

	// Assistant message style
	AssistantStyle = lipgloss.NewStyle().
		Foreground(RamboOrange).
		Bold(true)
)

// CreateGradient creates a visual gradient effect using ANSI colors
func CreateGradient(text string, startColor, endColor lipgloss.Color) string {
	// Simple gradient simulation using color codes
	// Real gradients require more complex terminal support
	style := lipgloss.NewStyle().
		Foreground(startColor).
		Bold(true)
	return style.Render(text)
}

// GetTypingFrame returns the current typing animation frame
func GetTypingFrame(frame int) string {
	return TypingFrames[frame%len(TypingFrames)]
}

// Render3DBox renders a 3D-style box with shadow effect
func Render3DBox(content string, width int) string {
	// Top border with light effect
	top := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#4A4A5A")).
		Render("╔" + repeat("═", width-2) + "╗")
	
	// Content area
	middle := Box3D.Width(width).Render(content)
	
	// Bottom border with shadow effect
	bottom := lipgloss.NewStyle().
		Foreground(DarkBorder).
		Render("╚" + repeat("═", width-2) + "╝")
	
	return top + "\n" + middle + "\n" + bottom
}

// RenderNeonText creates a neon glow effect
func RenderNeonText(text string, color lipgloss.Color) string {
	glow := lipgloss.NewStyle().
		Foreground(color).
		Bold(true).
		Render(text)
	
	return glow
}

// RenderTacticalHeader creates a tactical-style header
func RenderTacticalHeader(title string, width int) string {
	left := lipgloss.NewStyle().
		Foreground(RamboRed).
		Render("◆")
	
	right := lipgloss.NewStyle().
		Foreground(RamboRed).
		Render("◆")
	
	titleStyle := lipgloss.NewStyle().
		Foreground(RamboOrange).
		Bold(true).
		Render(title)
	
	line := lipgloss.NewStyle().
		Foreground(DarkBorder).
		Render(repeat("─", (width-len(title)-6)/2))
	
	return line + " " + left + " " + titleStyle + " " + right + " " + line
}

func repeat(s string, n int) string {
	result := ""
	for i := 0; i < n; i++ {
		result += s
	}
	return result
}

// SwiftTalon TUI - Dracula Dark Theme Interface
// Clean, modern terminal UI with Bubble Tea and Glamour

package tui

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/glamour"
	"github.com/charmbracelet/lipgloss"

	"github.com/Bhuw1234/swifttalon/pkg/agent"
	"github.com/Bhuw1234/swifttalon/pkg/bus"
	"github.com/Bhuw1234/swifttalon/pkg/config"
	"github.com/Bhuw1234/swifttalon/pkg/providers"
)

// Key bindings
type keyMap struct {
	Up      key.Binding
	Down    key.Binding
	Enter   key.Binding
	Tab     key.Binding
	Esc     key.Binding
	CtrlC   key.Binding
	CtrlN   key.Binding
	CtrlM   key.Binding
	Help    key.Binding
	Clear   key.Binding
}

func defaultKeyMap() keyMap {
	return keyMap{
		Up: key.NewBinding(
			key.WithKeys("up", "k"),
			key.WithHelp("↑/k", "up"),
		),
		Down: key.NewBinding(
			key.WithKeys("down", "j"),
			key.WithHelp("↓/j", "down"),
		),
		Enter: key.NewBinding(
			key.WithKeys("enter"),
			key.WithHelp("↵", "send"),
		),
		Tab: key.NewBinding(
			key.WithKeys("tab"),
			key.WithHelp("tab", "switch"),
		),
		Esc: key.NewBinding(
			key.WithKeys("esc"),
			key.WithHelp("esc", "close"),
		),
		CtrlC: key.NewBinding(
			key.WithKeys("ctrl+c"),
			key.WithHelp("ctrl+c", "quit"),
		),
		CtrlN: key.NewBinding(
			key.WithKeys("ctrl+n"),
			key.WithHelp("ctrl+n", "new session"),
		),
		CtrlM: key.NewBinding(
			key.WithKeys("ctrl+m"),
			key.WithHelp("ctrl+m", "model"),
		),
		Help: key.NewBinding(
			key.WithKeys("?"),
			key.WithHelp("?", "help"),
		),
		Clear: key.NewBinding(
			key.WithKeys("ctrl+l"),
			key.WithHelp("ctrl+l", "clear"),
		),
	}
}

// Messages
type (
	MsgResponse struct {
		Content string
	}
	MsgError struct {
		Error error
	}
	MsgTypingTick time.Time
)

// Model represents the TUI state
type Model struct {
	// UI Components
	keys        keyMap
	textInput   textinput.Model
	viewport    viewport.Model
	glamRender  *glamour.TermRenderer

	// State
	width      int
	height     int
	focus      FocusArea
	ready      bool
	err        error
	showError  bool

	// Messages
	messages   []Message
	isTyping   bool
	typingFrame int

	// Sessions
	sessions   []Session
	sessionIdx int

	// Models
	models      []string
	modelIdx    int
	modelOpen   bool

	// Config & Agent
	cfg        *config.Config
	agentLoop  *agent.AgentLoop
	ctx        context.Context
	cancel     context.CancelFunc
	wg         *sync.WaitGroup
	sessionKey string
}

// Run starts the TUI
func Run(cfg *config.Config) error {
	provider, err := providers.CreateProvider(cfg)
	if err != nil {
		return fmt.Errorf("failed to create provider: %w", err)
	}
	provider = agent.CreateProviderWithFallback(cfg, provider)

	msgBus := bus.NewMessageBus()
	agentLoop := agent.NewAgentLoop(cfg, msgBus, provider)

	ctx, cancel := context.WithCancel(context.Background())

	// Initialize text input
	ti := textinput.New()
	ti.Placeholder = "Type your message..."
	ti.Focus()
	ti.CharLimit = 500
	ti.Width = 60

	// Initialize viewport
	vp := viewport.New(80, 20)
	vp.SetContent("")

	// Initialize glamour renderer
	glamRender, err := glamour.NewTermRenderer(
		glamour.WithStandardStyle("dracula"),
		glamour.WithWordWrap(78),
	)
	if err != nil {
		glamRender = nil // Fallback to no glamour
	}

	m := Model{
		cfg:        cfg,
		agentLoop:  agentLoop,
		ctx:        ctx,
		cancel:     cancel,
		wg:         new(sync.WaitGroup),
		sessionKey: "tui:default",
		messages:   make([]Message, 0),
		textInput:  ti,
		viewport:   vp,
		glamRender: glamRender,
		sessions: []Session{
			{ID: "default", Name: "Main Session", CreatedAt: time.Now()},
		},
		models: []string{
			cfg.Agents.Defaults.Model,
			"gpt-4o",
			"claude-3-opus",
			"claude-3-sonnet",
			"gemini-2.0-flash",
		},
	}

	p := tea.NewProgram(m, tea.WithAltScreen())
	_, err = p.Run()
	return err
}

// Init implements tea.Model
func (m Model) Init() tea.Cmd {
	return tea.Batch(
		textinput.Blink,
		tea.Tick(100*time.Millisecond, func(t time.Time) tea.Msg {
			return MsgTypingTick(t)
		}),
	)
}

// Update implements tea.Model
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		return m.handleKeyPress(msg)

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.viewport.Width = msg.Width - 4
		m.viewport.Height = msg.Height - 10
		m.ready = true
		// Re-create glamour renderer with new width
		if m.viewport.Width > 10 {
			gr, err := glamour.NewTermRenderer(
				glamour.WithStandardStyle("dracula"),
				glamour.WithWordWrap(m.viewport.Width - 4),
			)
			if err == nil {
				m.glamRender = gr
			}
		}

	case MsgResponse:
		m.isTyping = false
		m.messages = append(m.messages, Message{
			Role:      "assistant",
			Content:   msg.Content,
			Timestamp: time.Now(),
		})
		m.updateViewport()

	case MsgError:
		m.isTyping = false
		m.err = msg.Error
		m.showError = true

	case MsgTypingTick:
		if m.isTyping {
			m.typingFrame = (m.typingFrame + 1) % len(TypingFrames)
		}
		return m, tea.Tick(100*time.Millisecond, func(t time.Time) tea.Msg {
			return MsgTypingTick(t)
		})
	}

	return m, tea.Batch(cmds...)
}

// View implements tea.Model
func (m Model) View() string {
	if !m.ready {
		return m.renderLoading()
	}

	// Layout
	var sections []string

	// 1. Header with ASCII art
	sections = append(sections, m.renderHeader())

	// 2. Main content area (viewport)
	sections = append(sections, m.renderContent())

	// 3. Input area
	sections = append(sections, m.renderInput())

	// 4. Status bar
	sections = append(sections, m.renderStatus())

	// 5. Help bar
	sections = append(sections, m.renderHelp())

	return lipgloss.JoinVertical(lipgloss.Left, sections...)
}

func (m Model) renderLoading() string {
	loadingStyle := lipgloss.NewStyle().
		Foreground(Orange).
		Bold(true).
		Padding(2, 4)

	frame := GetTypingFrame(m.typingFrame)
	return loadingStyle.Render(fmt.Sprintf("\n  %s INITIALIZING SWIFTTALON...\n", frame))
}

func (m Model) renderHeader() string {
	// ASCII Art Banner
	banner := m.renderAsciiBanner()
	
	// Model selector
	model := m.models[m.modelIdx]
	modelDisplay := ModelSelector.Render(fmt.Sprintf("⚡ %s", model))

	// Session indicator
	sessionName := m.sessions[m.sessionIdx].Name
	sessionDisplay := SessionActive.Render(fmt.Sprintf("◈ %s", sessionName))

	// Header line with border style
	headerLine := lipgloss.NewStyle().
		Foreground(BorderLight).
		Render(strings.Repeat("─", m.width-4))

	// Combine header info
	infoLine := lipgloss.JoinHorizontal(lipgloss.Top, modelDisplay, "  ", sessionDisplay)
	space := strings.Repeat(" ", max(0, m.width-lipgloss.Width(infoLine)-4))

	header := lipgloss.NewStyle().
		Background(Black).
		Padding(0, 2).
		Render(infoLine + space)

	return lipgloss.JoinVertical(lipgloss.Left, banner, header, headerLine)
}

func (m Model) renderAsciiBanner() string {
	// SwiftTalon ASCII art with Dracula colors
	bannerStyle := lipgloss.NewStyle().
		Foreground(Cyan).
		Bold(true)

	// Simple compact banner
	banner := `
   ███████╗██╗    ██╗██╗███████╗████████╗████████╗ █████╗ ██╗      ██████╗ ███╗   ██╗
   ██╔════╝██║    ██║██║██╔════╝╚══██╔══╝╚══██╔══╝██╔══██╗██║     ██╔═══██╗████╗  ██║
   ███████╗██║ █╗ ██║██║█████╗     ██║      ██║   ███████║██║     ██║   ██║██╔██╗ ██║
   ╚════██║██║███╗██║██║██╔══╝     ██║      ██║   ██╔══██║██║     ██║   ██║██║╚██╗██║
   ███████║╚███╔███╔╝██║██║        ██║      ██║   ██║  ██║███████╗╚██████╔╝██║ ╚████║
   ╚══════╝ ╚══╝╚══╝ ╚═╝╚═╝        ╚═╝      ╚═╝   ╚═╝  ╚═╝╚══════╝ ╚═════╝ ╚═╝  ╚═══╝`

	// Style the banner
	styledBanner := bannerStyle.Render(banner)

	// Subtitle
	subtitleStyle := lipgloss.NewStyle().
		Foreground(Pink).
		Bold(true).
		PaddingLeft(4)

	subtitle := subtitleStyle.Render("🐙 THE OCTOPUS PROJECT 🐙")

	// Container
	containerStyle := lipgloss.NewStyle().
		Background(Black).
		Padding(0, 1).
		Width(m.width - 4)

	return containerStyle.Render(styledBanner + "\n" + subtitle)
}

func (m Model) renderContent() string {
	// Content panel with dark theme
	contentStyle := lipgloss.NewStyle().
		Background(DarkBg).
		Border(lipgloss.RoundedBorder()).
		BorderForeground(BorderDark).
		Padding(1, 1).
		Width(m.width - 4).
		Height(m.viewport.Height)

	content := m.viewport.View()
	if content == "" {
		// Welcome message
		welcome := RenderTacticalHeader("TACTICAL AI INTERFACE", m.width-8)
		welcome += "\n\n"
		welcome += GlowText.Render("  Ready for commands.")
		welcome += "\n"
		welcome += SubtitleStyle.Render("  Type your message and press Enter to engage.")
		content = welcome
	}

	return contentStyle.Render(content)
}

func (m Model) renderInput() string {
	// Input prompt with green styling
	prompt := lipgloss.NewStyle().
		Foreground(Green).
		Bold(true).
		Render("▸ ")

	// Input box
	inputStyle := lipgloss.NewStyle().
		Background(SurfaceBg).
		Foreground(TextPrimary).
		Border(lipgloss.NormalBorder()).
		BorderForeground(Green).
		Padding(0, 1).
		Width(m.width - 10)

	input := m.textInput.View()
	inputBox := inputStyle.Render(prompt + input)

	// Typing indicator
	if m.isTyping {
		frame := GetTypingFrame(m.typingFrame)
		typingStyle := lipgloss.NewStyle().
			Foreground(Orange).
			Bold(true)
		indicator := typingStyle.Render(fmt.Sprintf("  %s AI PROCESSING...", frame))
		return inputBox + "\n" + indicator
	}

	return inputBox
}

func (m Model) renderStatus() string {
	// Status bar with dark styling
	statusStyle := lipgloss.NewStyle().
		Background(SurfaceBg).
		Foreground(TextSecondary).
		Padding(0, 2).
		Width(m.width - 4)

	// Left: Connection status
	connStatus := StatusOnline.Render("● CONNECTED")

	// Right: Message count
	msgCount := fmt.Sprintf("MSG: %d", len(m.messages))
	msgDisplay := lipgloss.NewStyle().
		Foreground(Cyan).
		Render(msgCount)

	// Center: Space
	space := strings.Repeat(" ", m.width-lipgloss.Width(connStatus)-lipgloss.Width(msgDisplay)-12)

	return statusStyle.Render(connStatus + space + msgDisplay)
}

func (m Model) renderHelp() string {
	helpStyle := lipgloss.NewStyle().
		Foreground(TextMuted).
		Background(Black).
		Padding(0, 2).
		Width(m.width - 4)

	shortcuts := []string{
		"↵ Send",
		"Ctrl+N New",
		"Ctrl+M Model",
		"Ctrl+L Clear",
		"? Help",
		"Esc Quit",
	}

	return helpStyle.Render(strings.Join(shortcuts, "  │  "))
}

func (m Model) handleKeyPress(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch {
	case key.Matches(msg, m.keys.CtrlC):
		return m, tea.Quit

	case key.Matches(msg, m.keys.Esc):
		if m.showError {
			m.showError = false
			return m, nil
		}
		return m, tea.Quit

	case key.Matches(msg, m.keys.Enter):
		return m.handleSend()

	case key.Matches(msg, m.keys.CtrlN):
		return m.handleNewSession()

	case key.Matches(msg, m.keys.CtrlM):
		m.modelIdx = (m.modelIdx + 1) % len(m.models)
		return m, nil

	case key.Matches(msg, m.keys.Clear):
		m.messages = nil
		m.viewport.SetContent("")
		return m, nil
	}

	// Handle text input
	var cmd tea.Cmd
	m.textInput, cmd = m.textInput.Update(msg)
	return m, cmd
}

func (m Model) handleSend() (tea.Model, tea.Cmd) {
	input := strings.TrimSpace(m.textInput.Value())
	if input == "" {
		return m, nil
	}

	// Add user message
	m.messages = append(m.messages, Message{
		Role:      "user",
		Content:   input,
		Timestamp: time.Now(),
	})

	// Clear input
	m.textInput.SetValue("")
	m.isTyping = true

	// Update viewport
	m.updateViewport()

	// Create channel for async response
	responseChan := make(chan struct {
		content string
		err     error
	}, 1)

	// Send to agent in goroutine
	go func() {
		response, err := m.agentLoop.ProcessDirect(m.ctx, input, m.sessionKey)
		responseChan <- struct {
			content string
			err     error
		}{content: response, err: err}
	}()

	// Return tea.Cmd that waits for response and sends appropriate message
	return m, func() tea.Msg {
		result := <-responseChan
		if result.err != nil {
			return MsgError{Error: result.err}
		}
		return MsgResponse{Content: result.content}
	}
}

func (m Model) handleNewSession() (tea.Model, tea.Cmd) {
	newSession := Session{
		ID:        fmt.Sprintf("session-%d", time.Now().Unix()),
		Name:      fmt.Sprintf("Session %d", len(m.sessions)+1),
		CreatedAt: time.Now(),
	}
	m.sessions = append(m.sessions, newSession)
	m.sessionIdx = len(m.sessions) - 1
	m.messages = nil
	m.sessionKey = newSession.ID
	m.viewport.SetContent("")
	return m, nil
}

func (m *Model) updateViewport() {
	var content strings.Builder

	for _, msg := range m.messages {
		var style lipgloss.Style
		var prefix string

		if msg.Role == "user" {
			style = UserStyle
			prefix = "👤 YOU"
		} else {
			style = AssistantStyle
			prefix = "🤖 AI"
		}

		// Message header
		header := lipgloss.NewStyle().
			Foreground(style.GetForeground()).
			Bold(true).
			Render(prefix)

		// Timestamp
		ts := lipgloss.NewStyle().
			Foreground(TextMuted).
			Render(msg.Timestamp.Format("15:04"))

		content.WriteString(fmt.Sprintf("%s %s\n", header, ts))

		// Message content - render with glamour if available
		msgContent := msg.Content
		if m.glamRender != nil {
			// Try to render as markdown
			rendered, err := m.glamRender.Render(msgContent)
			if err == nil {
				msgContent = strings.TrimSpace(rendered)
			}
		}

		// Message content with styling
		contentStyle := lipgloss.NewStyle().
			Foreground(TextPrimary).
			PaddingLeft(2).
			Width(m.viewport.Width - 4)

		content.WriteString(contentStyle.Render(msgContent))
		content.WriteString("\n\n")
	}

	m.viewport.SetContent(content.String())
	m.viewport.GotoBottom()
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
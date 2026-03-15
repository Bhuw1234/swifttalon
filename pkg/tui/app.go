// SwiftTalon TUI - RAMBO 3D Interface
// Bold, aggressive, tactical terminal UI with Bubble Tea

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
	keys      keyMap
	textInput textinput.Model
	viewport  viewport.Model

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
	wg         sync.WaitGroup
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

	m := Model{
		cfg:        cfg,
		agentLoop:  agentLoop,
		ctx:        ctx,
		cancel:     cancel,
		sessionKey: "tui:default",
		messages:   make([]Message, 0),
		textInput:  ti,
		viewport:   vp,
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

	// 1. Header with 3D effect
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
		Foreground(RamboOrange).
		Bold(true).
		Padding(2, 4)

	frame := GetTypingFrame(m.typingFrame)
	return loadingStyle.Render(fmt.Sprintf("\n  %s INITIALIZING SWIFTTALON TACTICAL INTERFACE...\n", frame))
}

func (m Model) renderHeader() string {
	// Logo with neon effect
	logo := LogoStyle.Render("🐙 SWIFTTALON")

	// Model selector
	model := m.models[m.modelIdx]
	modelDisplay := ModelSelector.Render(fmt.Sprintf("⚡ %s", model))

	// Session indicator
	sessionName := m.sessions[m.sessionIdx].Name
	sessionDisplay := SessionActive.Render(fmt.Sprintf("◈ %s", sessionName))

	// Tactical header line
	headerLine := lipgloss.NewStyle().
		Foreground(DarkBorder).
		Render(strings.Repeat("─", m.width-4))

	// Combine
	leftSection := lipgloss.JoinHorizontal(lipgloss.Top, logo, "  ", modelDisplay)
	rightSection := sessionDisplay

	// Space between
	space := strings.Repeat(" ", max(0, m.width-lipgloss.Width(leftSection)-lipgloss.Width(rightSection)-4))

	header := lipgloss.NewStyle().
		Background(DarkBase).
		Padding(0, 2).
		Render(leftSection + space + rightSection)

	return lipgloss.JoinVertical(lipgloss.Left, header, headerLine)
}

func (m Model) renderContent() string {
	// 3D panel for content
	contentStyle := lipgloss.NewStyle().
		Background(DarkPanel).
		Border(lipgloss.DoubleBorder()).
		BorderForeground(ElectricBlue).
		Padding(1, 2).
		Width(m.width - 4).
		Height(m.viewport.Height)

	content := m.viewport.View()
	if content == "" {
		// Welcome message with 3D effect
		welcome := RenderTacticalHeader("TACTICAL AI INTERFACE", m.width-8)
		welcome += "\n\n"
		welcome += GlowText.Render("  Ready for combat.")
		welcome += "\n"
		welcome += SubtitleStyle.Render("  Type your command and press Enter to engage.")
		content = welcome
	}

	return contentStyle.Render(content)
}

func (m Model) renderInput() string {
	// Input prompt with tactical styling
	prompt := GlowText.Render("▸ ")

	// Input box
	inputStyle := lipgloss.NewStyle().
		Background(DarkSurface).
		Foreground(TextPrimary).
		Border(lipgloss.NormalBorder()).
		BorderForeground(RamboGreen).
		Padding(0, 1).
		Width(m.width - 10)

	input := m.textInput.View()
	inputBox := inputStyle.Render(prompt + input)

	// Typing indicator
	if m.isTyping {
		frame := GetTypingFrame(m.typingFrame)
		typingStyle := lipgloss.NewStyle().
			Foreground(RamboOrange).
			Bold(true)
		indicator := typingStyle.Render(fmt.Sprintf("  %s AI PROCESSING...", frame))
		return inputBox + "\n" + indicator
	}

	return inputBox
}

func (m Model) renderStatus() string {
	// Status bar with tactical styling
	statusStyle := lipgloss.NewStyle().
		Background(DarkSurface).
		Foreground(TextSecondary).
		Padding(0, 2).
		Width(m.width - 4)

	// Left: Connection status
	connStatus := StatusOnline.Render("● CONNECTED")

	// Right: Message count
	msgCount := fmt.Sprintf("MSG: %d", len(m.messages))
	msgDisplay := lipgloss.NewStyle().
		Foreground(RamboCyan).
		Render(msgCount)

	// Center: Space
	space := strings.Repeat(" ", m.width-lipgloss.Width(connStatus)-lipgloss.Width(msgDisplay)-12)

	return statusStyle.Render(connStatus + space + msgDisplay)
}

func (m Model) renderHelp() string {
	helpStyle := lipgloss.NewStyle().
		Foreground(TextMuted).
		Background(DarkBase).
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

	// Send to agent
	go func() {
		response, err := m.agentLoop.ProcessDirect(m.ctx, input, m.sessionKey)
		if err != nil {
			// Send error
			return
		}
		// Response will be handled via message
		_ = response
	}()

	return m, nil
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

		// Message content with word wrap
		contentStyle := lipgloss.NewStyle().
			Foreground(TextPrimary).
			PaddingLeft(2).
			Width(m.viewport.Width - 4)

		content.WriteString(contentStyle.Render(msg.Content))
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

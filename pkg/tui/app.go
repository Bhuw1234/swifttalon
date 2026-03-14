// SwiftTalon TUI - Main Application
// AGGRESSIVE and SUPER EASY TO USE
// Package tui provides a beautiful terminal interface for SwiftTalon

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

// keyMap defines the keybindings for the TUI
type keyMap struct {
	Up         key.Binding
	Down       key.Binding
	Enter      key.Binding
	CtrlN      key.Binding // New session
	CtrlM      key.Binding // Model selector
	CtrlQ      key.Binding // Quit
	CtrlC      key.Binding // Quit
	Esc        key.Binding // Close modal / dismiss error
	Help       key.Binding
}

// defaultKeyMap returns the default keybindings - SIMPLIFIED
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
			key.WithHelp("enter", "send"),
		),
		CtrlN: key.NewBinding(
			key.WithKeys("ctrl+n"),
			key.WithHelp("ctrl+n", "new"),
		),
		CtrlM: key.NewBinding(
			key.WithKeys("ctrl+m"),
			key.WithHelp("ctrl+m", "model"),
		),
		CtrlQ: key.NewBinding(
			key.WithKeys("ctrl+q"),
			key.WithHelp("ctrl+q", "quit"),
		),
		CtrlC: key.NewBinding(
			key.WithKeys("ctrl+c"),
			key.WithHelp("ctrl+c", "quit"),
		),
		Esc: key.NewBinding(
			key.WithKeys("esc"),
			key.WithHelp("esc", "close"),
		),
		Help: key.NewBinding(
			key.WithKeys("?"),
			key.WithHelp("?", "help"),
		),
	}
}

// Model represents the TUI application state
type Model struct {
	// UI components
	keys        keyMap
	textInput   textinput.Model
	viewport    viewport.Model
	sessions    []Session
	sessionIdx  int

	// State
	width       int
	height      int
	focus       FocusArea
	ready       bool
	err         error
	showError   bool // Track if error is being displayed

	// Size error
	sizeError   bool

	// Modals
	showHelp         bool
	showModelSelector bool
	modelListIdx     int

	// Agent
	cfg         *config.Config
	agentLoop   *agent.AgentLoop
	ctx         context.Context
	cancel      context.CancelFunc

	// Current state
	currentSession string
	currentModel   string
	messages       []ChatMessage
	isStreaming    bool
	streamContent  string
	streamMu       sync.RWMutex // Mutex for streamContent race condition fix

	// Request tracking for cancellation
	pendingCancel context.CancelFunc
	pendingMu     sync.Mutex

	// Typing animation
	typingDots     int
}

// AgentResponseMsg represents a response from the agent
type AgentResponseMsg struct {
	Content string
	Error   error
	Done    bool
}

// TickMsg is used for animations
type TickMsg time.Time

// New creates a new TUI Model
func New(cfg *config.Config) (*Model, error) {
	// Create provider
	provider, err := providers.CreateProvider(cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to create provider: %w", err)
	}
	provider = agent.CreateProviderWithFallback(cfg, provider)

	// Create message bus
	msgBus := bus.NewMessageBus()

	// Create agent loop
	agentLoop := agent.NewAgentLoop(cfg, msgBus, provider)

	// Initialize text input - PROMINENT
	ti := textinput.New()
	ti.Placeholder = "Type your message and press Enter..."
	ti.Focus()
	ti.CharLimit = 4000
	ti.Width = 60

	// Style the text input
	ti.PromptStyle = PromptStyle
	ti.TextStyle = lipgloss.NewStyle().Foreground(colorTextPrimary)

	// Initialize viewport for chat
	vp := viewport.New(80, 20)

	m := &Model{
		keys:       defaultKeyMap(),
		textInput:  ti,
		viewport:   vp,
		sessions:   []Session{},
		sessionIdx: 0,
		focus:      FocusInput,
		cfg:        cfg,
		agentLoop:  agentLoop,
		currentModel: cfg.Agents.Defaults.Model,
		messages:   []ChatMessage{},
		typingDots: 0,
	}

	m.ctx, m.cancel = context.WithCancel(context.Background())

	// Load sessions
	m.loadSessions()

	return m, nil
}

// loadSessions loads sessions from disk
func (m *Model) loadSessions() {
	// Add a default session
	m.sessions = []Session{
		{
			Key:       "cli:default",
			Title:     "Chat",
			LastMsg:   "Start a conversation...",
			Updated:   time.Now(),
			MessageCount: 0,
		},
	}
	m.currentSession = "cli:default"
}

// Init initializes the TUI
func (m *Model) Init() tea.Cmd {
	return tea.Batch(
		textinput.Blink,
		m.tickCmd(),
	)
}

// tickCmd returns a tick command for animations
func (m *Model) tickCmd() tea.Cmd {
	return tea.Tick(time.Millisecond*200, func(t time.Time) tea.Msg {
		return TickMsg(t)
	})
}

// Update handles incoming messages
func (m *Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var (
		cmd  tea.Cmd
		cmds []tea.Cmd
	)

	switch msg := msg.(type) {
	case tea.KeyMsg:
		// Dismiss error with any key
		if m.showError {
			m.showError = false
			m.err = nil
			return m, nil
		}

		// Handle quit
		if key.Matches(msg, m.keys.CtrlQ) || key.Matches(msg, m.keys.CtrlC) {
			// Cancel any pending requests
			m.pendingMu.Lock()
			if m.pendingCancel != nil {
				m.pendingCancel()
				m.pendingCancel = nil
			}
			m.pendingMu.Unlock()

			// Cancel main context
			m.cancel()
			return m, tea.Quit
		}

		// Handle help toggle
		if key.Matches(msg, m.keys.Help) {
			m.showHelp = !m.showHelp
			return m, nil
		}

		// Handle escape - close modals or dismiss errors
		if key.Matches(msg, m.keys.Esc) {
			if m.showHelp || m.showModelSelector {
				m.showHelp = false
				m.showModelSelector = false
			}
			return m, nil
		}

		// Handle model selector
		if key.Matches(msg, m.keys.CtrlM) && !m.showModelSelector && !m.isStreaming {
			m.showModelSelector = true
			return m, nil
		}

		// If modal is open, handle modal input
		if m.showModelSelector {
			return m.handleModelSelector(msg)
		}

		// Handle new session
		if key.Matches(msg, m.keys.CtrlN) && !m.isStreaming {
			return m.newSession()
		}

		// Handle based on current focus
		switch m.focus {
		case FocusSidebar:
			return m.handleSidebarInput(msg)
		case FocusInput:
			return m.handleInputInput(msg)
		case FocusChat:
			return m.handleChatInput(msg)
		}

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height

		// Check minimum size
		if m.width < MinTerminalWidth || m.height < MinTerminalHeight {
			m.sizeError = true
			m.ready = false
			return m, nil
		}
		m.sizeError = false

		m.updateLayout()
		m.ready = true

	case AgentResponseMsg:
		if msg.Error != nil {
			m.err = msg.Error
			m.showError = true
			m.isStreaming = false
			return m, nil
		}
		if msg.Done {
			m.isStreaming = false
			// Update the last message with mutex protection
			m.streamMu.Lock()
			content := m.streamContent
			m.streamMu.Unlock()

			if len(m.messages) > 0 {
				m.messages[len(m.messages)-1].Content = content
				m.messages[len(m.messages)-1].Streaming = false
			}
			m.updateViewport()
		} else {
			// Thread-safe update of stream content
			m.streamMu.Lock()
			m.streamContent += msg.Content
			content := m.streamContent
			m.streamMu.Unlock()

			// Update the streaming message
			if len(m.messages) > 0 {
				m.messages[len(m.messages)-1].Content = content
			}
			m.updateViewport()
		}
		return m, nil

	case TickMsg:
		// Update typing animation
		if m.isStreaming {
			m.typingDots = (m.typingDots + 1) % 4
		}
		cmds = append(cmds, m.tickCmd())
	}

	// Update components
	m.textInput, cmd = m.textInput.Update(msg)
	cmds = append(cmds, cmd)

	m.viewport, cmd = m.viewport.Update(msg)
	cmds = append(cmds, cmd)

	return m, tea.Batch(cmds...)
}

// handleSidebarInput handles input when sidebar is focused
func (m *Model) handleSidebarInput(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch {
	case key.Matches(msg, m.keys.Up):
		if m.sessionIdx > 0 {
			m.sessionIdx--
		}
	case key.Matches(msg, m.keys.Down):
		if m.sessionIdx < len(m.sessions)-1 {
			m.sessionIdx++
		}
	case key.Matches(msg, m.keys.Enter):
		m.switchSession(m.sessionIdx)
		m.focus = FocusInput
	}
	return m, nil
}

// handleInputInput handles input when text input is focused
func (m *Model) handleInputInput(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch {
	case key.Matches(msg, m.keys.Enter) && m.textInput.Value() != "" && !m.isStreaming:
		return m.sendMessage()
	}

	m.textInput, cmd = m.textInput.Update(msg)
	return m, cmd
}

// handleChatInput handles input when chat is focused
func (m *Model) handleChatInput(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch {
	case key.Matches(msg, m.keys.Up):
		m.viewport.LineUp(1)
	case key.Matches(msg, m.keys.Down):
		m.viewport.LineDown(1)
	}

	m.viewport, cmd = m.viewport.Update(msg)
	return m, cmd
}

// handleModelSelector handles input in model selector modal
func (m *Model) handleModelSelector(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	models := m.getAvailableModels()

	switch {
	case key.Matches(msg, m.keys.Up):
		if m.modelListIdx > 0 {
			m.modelListIdx--
		}
	case key.Matches(msg, m.keys.Down):
		if m.modelListIdx < len(models)-1 {
			m.modelListIdx++
		}
	case key.Matches(msg, m.keys.Enter):
		if m.modelListIdx < len(models) {
			m.currentModel = models[m.modelListIdx].ID
		}
		m.showModelSelector = false
	case key.Matches(msg, m.keys.Esc):
		m.showModelSelector = false
	}
	return m, nil
}

// sendMessage sends the current input as a message
func (m *Model) sendMessage() (tea.Model, tea.Cmd) {
	content := m.textInput.Value()
	if content == "" {
		return m, nil
	}

	// Add user message
	m.messages = append(m.messages, ChatMessage{
		Role:      "user",
		Content:   content,
		Timestamp: time.Now(),
	})

	// Add placeholder for assistant response
	m.messages = append(m.messages, ChatMessage{
		Role:      "assistant",
		Content:   "",
		Timestamp: time.Now(),
		Streaming: true,
	})

	// Clear stream content with mutex protection
	m.streamMu.Lock()
	m.streamContent = ""
	m.streamMu.Unlock()

	m.isStreaming = true
	m.textInput.SetValue("")
	m.updateViewport()

	// Send to agent
	return m, m.sendToAgent(content)
}

// sendToAgent sends a message to the agent with proper cancellation
func (m *Model) sendToAgent(content string) tea.Cmd {
	return func() tea.Msg {
		// Get configurable timeout (default 120s)
		timeout := 120 * time.Second
		if m.cfg.Agents.Defaults.MaxTokens > 0 {
			// Scale timeout based on expected response size
			// More tokens = longer timeout
			if m.cfg.Agents.Defaults.MaxTokens > 8000 {
				timeout = 180 * time.Second
			}
		}

		ctx, cancel := context.WithTimeout(m.ctx, timeout)

		// Track this request for cancellation
		m.pendingMu.Lock()
		m.pendingCancel = cancel
		m.pendingMu.Unlock()

		// Cleanup on exit
		defer func() {
			m.pendingMu.Lock()
			m.pendingCancel = nil
			m.pendingMu.Unlock()
			cancel()
		}()

		response, err := m.agentLoop.ProcessDirect(ctx, content, m.currentSession)
		if err != nil {
			return AgentResponseMsg{Error: err, Done: true}
		}

		return AgentResponseMsg{Content: response, Done: true}
	}
}

// newSession creates a new chat session
func (m *Model) newSession() (tea.Model, tea.Cmd) {
	sessionKey := fmt.Sprintf("cli:session-%d", time.Now().Unix())
	newSession := Session{
		Key:       sessionKey,
		Title:     fmt.Sprintf("Chat %d", len(m.sessions)+1),
		LastMsg:   "New conversation",
		Updated:   time.Now(),
		MessageCount: 0,
	}

	m.sessions = append([]Session{newSession}, m.sessions...)
	m.sessionIdx = 0
	m.currentSession = sessionKey
	m.messages = []ChatMessage{}
	m.updateViewport()
	m.focus = FocusInput

	return m, nil
}

// switchSession switches to a different session
func (m *Model) switchSession(idx int) {
	if idx >= 0 && idx < len(m.sessions) {
		m.sessionIdx = idx
		m.currentSession = m.sessions[idx].Key
		m.messages = []ChatMessage{}
		m.updateViewport()
	}
}

// updateLayout updates the layout based on window size
func (m *Model) updateLayout() {
	sidebarWidth := 24
	inputHeight := 5
	statusHeight := 2 // Increased for shortcuts bar

	// Update viewport
	chatWidth := m.width - sidebarWidth - 4
	chatHeight := m.height - inputHeight - statusHeight - 4

	if chatWidth > 0 && chatHeight > 0 {
		m.viewport.Width = chatWidth
		m.viewport.Height = chatHeight
	}

	// Update text input
	if m.width > sidebarWidth {
		m.textInput.Width = m.width - sidebarWidth - 10
	}

	m.updateViewport()
}

// updateViewport updates the chat viewport content
func (m *Model) updateViewport() {
	var content strings.Builder

	if len(m.messages) == 0 {
		content.WriteString(m.renderWelcome())
	} else {
		for _, msg := range m.messages {
			content.WriteString(m.renderMessage(msg))
			content.WriteString("\n")
		}
	}

	m.viewport.SetContent(content.String())
}

// renderWelcome renders the welcome message
func (m *Model) renderWelcome() string {
	var b strings.Builder

	b.WriteString(WelcomeTitleStyle.Render("⚡ SWIFTTALON"))
	b.WriteString("\n\n")
	b.WriteString(WelcomeTextStyle.Render("Ultra-lightweight AI Assistant"))
	b.WriteString("\n\n")
	b.WriteString("Just type your message and press Enter.\n")
	b.WriteString("That's it. Simple.\n")

	return b.String()
}

// renderMessage renders a single message
func (m *Model) renderMessage(msg ChatMessage) string {
	var b strings.Builder

	// Role badge - BOLD
	var roleBadge string
	switch msg.Role {
	case "user":
		roleBadge = RoleUserStyle.Render(" YOU ")
	case "assistant":
		roleBadge = RoleAssistantStyle.Render(" AI ")
	default:
		roleBadge = RoleSystemStyle.Render(" SYS ")
	}

	// Timestamp
	timestamp := TimestampStyle.Render(msg.Timestamp.Format("15:04"))

	b.WriteString(lipgloss.JoinHorizontal(lipgloss.Top, roleBadge, " ", timestamp))
	b.WriteString("\n")

	// Content
	var contentStyle lipgloss.Style
	switch msg.Role {
	case "user":
		contentStyle = UserMessageStyle
	case "assistant":
		contentStyle = AssistantMessageStyle
	default:
		contentStyle = SystemMessageStyle
	}

	content := msg.Content
	if msg.Streaming {
		content += "▌" // Cursor for streaming
	}

	// Word wrap content
	wrappedContent := m.wrapText(content, m.viewport.Width-6)
	b.WriteString(contentStyle.Render(wrappedContent))
	b.WriteString("\n")

	return b.String()
}

// wrapText wraps text to a given width
func (m *Model) wrapText(text string, width int) string {
	if width <= 0 {
		return text
	}

	var result strings.Builder
	lines := strings.Split(text, "\n")

	for _, line := range lines {
		if len(line) <= width {
			result.WriteString(line)
			result.WriteString("\n")
			continue
		}

		// Simple word wrap
		words := strings.Fields(line)
		currentLen := 0
		for _, word := range words {
			wordLen := len(word)
			if currentLen+wordLen+1 > width {
				result.WriteString("\n")
				currentLen = 0
			}
			if currentLen > 0 {
				result.WriteString(" ")
				currentLen++
			}
			result.WriteString(word)
			currentLen += wordLen
		}
		result.WriteString("\n")
	}

	return result.String()
}

// getAvailableModels returns a list of available models
func (m *Model) getAvailableModels() []ModelInfo {
	return []ModelInfo{
		{ID: "glm-4.7", Name: "GLM-4.7", Provider: "Zhipu", Description: "Best for Chinese"},
		{ID: "gpt-4o", Name: "GPT-4o", Provider: "OpenAI", Description: "Best overall"},
		{ID: "claude-3-5-sonnet", Name: "Claude 3.5 Sonnet", Provider: "Anthropic", Description: "Fast & capable"},
		{ID: "gemini-2.0-flash", Name: "Gemini 2.0 Flash", Provider: "Google", Description: "Fast responses"},
		{ID: "deepseek-chat", Name: "DeepSeek Chat", Provider: "DeepSeek", Description: "Reasoning"},
	}
}

// View renders the TUI
func (m *Model) View() string {
	// Show size error if terminal is too small
	if m.sizeError {
		return m.renderSizeError()
	}

	if !m.ready {
		return "\n  Loading..."
	}

	// Render sidebar
	sidebar := m.renderSidebar()

	// Render chat area
	chat := m.renderChat()

	// Render input area
	input := m.renderInput()

	// Render status bar with shortcuts
	status := m.renderStatusBar()

	// Render shortcuts bar - ALWAYS VISIBLE
	shortcuts := m.renderShortcutsBar()

	// Combine layout
	mainContent := lipgloss.JoinHorizontal(
		lipgloss.Top,
		sidebar,
		chat,
	)

	fullContent := lipgloss.JoinVertical(
		lipgloss.Left,
		mainContent,
		input,
		status,
		shortcuts,
	)

	// Render modals on top
	if m.showHelp {
		fullContent = m.renderHelpModal(fullContent)
	}

	if m.showModelSelector {
		fullContent = m.renderModelSelectorModal(fullContent)
	}

	// Render error if any - OBVIOUS
	if m.err != nil && m.showError {
		errMsg := ErrorStyle.Render(fmt.Sprintf(" ⚠ ERROR: %v - Press any key to dismiss ", m.err))
		fullContent = lipgloss.JoinVertical(lipgloss.Left, errMsg, fullContent)
	}

	return fullContent
}

// renderSizeError renders an error when terminal is too small
func (m *Model) renderSizeError() string {
	errMsg := ErrorStyle.Render(
		fmt.Sprintf(" ⚠ TERMINAL TOO SMALL! Need %dx%d, got %dx%d ",
			MinTerminalWidth, MinTerminalHeight, m.width, m.height))
	return "\n" + errMsg + "\n\nResize your terminal window.\n"
}

// renderSidebar renders the session sidebar - COMPACT
func (m *Model) renderSidebar() string {
	var b strings.Builder

	// Header
	b.WriteString(TitleStyle.Render(" SESSIONS "))
	b.WriteString("\n\n")

	// Session list
	for i, session := range m.sessions {
		var style lipgloss.Style
		if i == m.sessionIdx && m.focus == FocusSidebar {
			style = SessionActiveStyle
		} else {
			style = SessionItemStyle
		}

		title := SessionTitleStyle.Render(session.Title)
		meta := SessionMetaStyle.Render(fmt.Sprintf("%d msgs", session.MessageCount))

		b.WriteString(style.Render(lipgloss.JoinVertical(lipgloss.Left, title, meta)))
		b.WriteString("\n")
	}

	return SidebarStyle.Render(b.String())
}

// renderChat renders the chat viewport
func (m *Model) renderChat() string {
	return ChatStyle.Render(m.viewport.View())
}

// renderInput renders the input area - PROMINENT!
func (m *Model) renderInput() string {
	var style lipgloss.Style
	if m.focus == FocusInput {
		style = InputFocusedStyle
	} else {
		style = InputStyle
	}

	prompt := PromptStyle.Render("▶")
	inputContent := prompt + " " + m.textInput.View()

	if m.isStreaming {
		// Animated typing indicator
		dots := strings.Repeat(".", m.typingDots)
		typing := TypingIndicatorStyle.Render("⚡ Thinking" + dots)
		inputContent = lipgloss.JoinVertical(lipgloss.Left, inputContent, "  "+typing)
	}

	return style.Width(m.width - 28).Render(inputContent)
}

// renderStatusBar renders the status bar with focus indicator
func (m *Model) renderStatusBar() string {
	// Model badge - OBVIOUS
	modelBadge := ModelBadgeStyle.Render(" " + m.currentModel + " ")

	// Focus indicator
	var focusText string
	switch m.focus {
	case FocusSidebar:
		focusText = FocusIndicatorStyle.Render(" SIDEBAR ")
	case FocusInput:
		focusText = FocusIndicatorStyle.Render(" INPUT ")
	case FocusChat:
		focusText = FocusIndicatorStyle.Render(" CHAT ")
	}

	// Connection status
	var statusText string
	if m.isStreaming {
		statusText = StatusLoadingStyle.Render("● WORKING")
	} else {
		statusText = StatusConnectedStyle.Render("● READY")
	}

	// Combine
	left := lipgloss.JoinHorizontal(lipgloss.Top, modelBadge, " ", focusText, " ", statusText)
	right := lipgloss.NewStyle().Foreground(colorTextMuted).Render("? for help")

	padding := m.width - lipgloss.Width(left) - lipgloss.Width(right) - 2

	if padding > 0 {
		spacer := lipgloss.NewStyle().Width(padding).Render("")
		return StatusBarStyle.Render(lipgloss.JoinHorizontal(lipgloss.Top, left, spacer, right))
	}

	return StatusBarStyle.Render(lipgloss.JoinHorizontal(lipgloss.Top, left, right))
}

// renderShortcutsBar renders the ALWAYS VISIBLE shortcuts bar
func (m *Model) renderShortcutsBar() string {
	shortcuts := []struct {
		key  string
		desc string
	}{
		{"Enter", "Send"},
		{"Ctrl+N", "New Chat"},
		{"Ctrl+M", "Model"},
		{"Ctrl+Q", "Quit"},
		{"?", "Help"},
	}

	var items []string
	for _, s := range shortcuts {
		key := ShortcutKeyStyle.Render(s.key)
		desc := ShortcutDescStyle.Render(s.desc)
		items = append(items, lipgloss.JoinHorizontal(lipgloss.Top, key, " ", desc))
	}

	content := lipgloss.JoinHorizontal(lipgloss.Top, items...)

	padding := m.width - lipgloss.Width(content) - 2
	if padding > 0 {
		spacer := lipgloss.NewStyle().Width(padding / 2).Render("")
		content = lipgloss.JoinHorizontal(lipgloss.Top, spacer, content)
	}

	return ShortcutBarStyle.Width(m.width).Render(content)
}

// renderHelpModal renders the help modal - SIMPLE
func (m *Model) renderHelpModal(base string) string {
	helpContent := `
  ⌨️  KEYBOARD SHORTCUTS

  Enter      Send message
  Ctrl+N     New session
  Ctrl+M     Change model
  Ctrl+Q     Quit

  ↑/↓        Navigate
  Esc        Close this

  Just type and press Enter!
`

	helpBox := ModalStyle.Render(helpContent)

	// Center the modal
	return lipgloss.Place(
		m.width, m.height,
		lipgloss.Center, lipgloss.Center,
		helpBox,
		lipgloss.WithWhitespaceChars(" "),
		lipgloss.WithWhitespaceForeground(colorBgPrimary),
	)
}

// renderModelSelectorModal renders the model selector modal
func (m *Model) renderModelSelectorModal(base string) string {
	models := m.getAvailableModels()

	var b strings.Builder
	b.WriteString(ModalTitleStyle.Render(" SELECT MODEL "))
	b.WriteString("\n\n")

	for i, model := range models {
		var style lipgloss.Style
		if i == m.modelListIdx {
			style = ModelItemSelectedStyle
		} else {
			style = ModelItemStyle
		}

		line := fmt.Sprintf("  %s  %s", model.Name, model.Description)
		b.WriteString(style.Render(line))
		b.WriteString("\n")
	}

	b.WriteString("\n")
	b.WriteString(SessionMetaStyle.Render("  ↑↓ navigate  ·  Enter select  ·  Esc cancel"))

	content := b.String()
	box := ModalStyle.Render(content)

	return lipgloss.Place(
		m.width, m.height,
		lipgloss.Center, lipgloss.Center,
		box,
		lipgloss.WithWhitespaceChars(" "),
		lipgloss.WithWhitespaceForeground(colorBgPrimary),
	)
}

// Run starts the TUI application
func Run(cfg *config.Config) error {
	m, err := New(cfg)
	if err != nil {
		return err
	}

	p := tea.NewProgram(
		m,
		tea.WithAltScreen(),
		tea.WithMouseCellMotion(),
	)

	_, err = p.Run()
	return err
}
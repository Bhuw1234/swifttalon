// SwiftTalon - Ultra-lightweight personal AI agent
// License: MIT
//
// Copyright (c) 2026 SwiftTalon contributors

// Package hooks provides event-driven automation for agent lifecycle events.
// Hooks allow users to define scripts or commands that run at specific points
// during agent execution.
package hooks

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"

	"github.com/Bhuw1234/swifttalon/pkg/logger"
)

// HookType represents the type of hook event
type HookType string

const (
	// PreTool fires before a tool is executed
	PreTool HookType = "pre_tool"
	// PostTool fires after a tool is executed
	PostTool HookType = "post_tool"
	// PreLLM fires before an LLM call
	PreLLM HookType = "pre_llm"
	// PostLLM fires after an LLM call
	PostLLM HookType = "post_llm"
	// OnError fires when an error occurs
	OnError HookType = "on_error"
	// OnMessage fires when a message is received
	OnMessage HookType = "on_message"
	// OnMessageSent fires when a message is sent
	OnMessageSent HookType = "on_message_sent"
)

// String returns the string representation of HookType
func (h HookType) String() string {
	return string(h)
}

// AllHookTypes returns all supported hook types
func AllHookTypes() []HookType {
	return []HookType{
		PreTool,
		PostTool,
		PreLLM,
		PostLLM,
		OnError,
		OnMessage,
		OnMessageSent,
	}
}

// HookEvent contains context data for a hook event
type HookEvent struct {
	Type      HookType           // Hook type (pre_tool, post_tool, etc.)
	ToolName  string             // Name of tool being executed (for tool hooks)
	ToolInput json.RawMessage   // JSON input to tool
	ToolOutput json.RawMessage  // JSON output from tool (for post_tool)
	Error     string            // Error message if any (for on_error)
	Message   string            // User message (for on_message)
	Response  string            // Agent response (for on_message_sent)
	SessionKey string           // Session identifier
	Channel   string            // Channel where message came from
	ChatID    string            // Chat ID
	Metadata  map[string]interface{} // Additional metadata
}

// HookHandler defines the interface for hook handlers
type HookHandler interface {
	Name() string
	Execute(ctx context.Context, event *HookEvent) error
}

// ScriptHook is a hook that executes an external script
type ScriptHook struct {
	name     string
	Script   string
	Args     []string
	Env      []string
	Dir      string
	Timeout  int // seconds, 0 = no timeout
}

// NewScriptHook creates a new script hook
func NewScriptHook(name, script string, args, env []string, dir string, timeout int) *ScriptHook {
	return &ScriptHook{
		name:    name,
		Script:  script,
		Args:    args,
		Env:     env,
		Dir:     dir,
		Timeout: timeout,
	}
}

// Name returns the hook name
func (h *ScriptHook) Name() string {
	return h.name
}

// Execute runs the hook script with the provided event data
func (h *ScriptHook) Execute(ctx context.Context, event *HookEvent) error {
	// Prepare environment variables
	env := os.Environ()
	
	// Add hook-specific environment variables
	env = append(env, fmt.Sprintf("HOOK_TYPE=%s", event.Type))
	env = append(env, fmt.Sprintf("TOOL_NAME=%s", event.ToolName))
	env = append(env, fmt.Sprintf("SESSION_KEY=%s", event.SessionKey))
	env = append(env, fmt.Sprintf("CHANNEL=%s", event.Channel))
	env = append(env, fmt.Sprintf("CHAT_ID=%s", event.ChatID))
	
	if event.ToolInput != nil {
		env = append(env, fmt.Sprintf("TOOL_INPUT=%s", string(event.ToolInput)))
	}
	if event.ToolOutput != nil {
		env = append(env, fmt.Sprintf("TOOL_OUTPUT=%s", string(event.ToolOutput)))
	}
	if event.Error != "" {
		env = append(env, fmt.Sprintf("ERROR=%s", event.Error))
	}
	if event.Message != "" {
		env = append(env, fmt.Sprintf("MESSAGE=%s", event.Message))
	}
	if event.Response != "" {
		env = append(env, fmt.Sprintf("RESPONSE=%s", event.Response))
	}
	
	// Add custom environment variables from config
	for _, e := range h.Env {
		env = append(env, e)
	}
	
	// Prepare command
	cmd := exec.CommandContext(ctx, h.Script, h.Args...)
	cmd.Env = env
	
	if h.Dir != "" {
		cmd.Dir = h.Dir
	}
	
	// Set up stdout/stderr capture
	output, err := cmd.CombinedOutput()
	
	if len(output) > 0 {
		logger.DebugCF("hooks", "Hook script output: %s",
			map[string]interface{}{
				"hook":  h.Name,
				"output": string(output),
			})
	}
	
	if err != nil {
		logger.WarnCF("hooks", "Hook script failed: %v",
			map[string]interface{}{
				"hook":  h.Name(),
				"error": err.Error(),
			})
		return fmt.Errorf("hook %q failed: %w", h.Name(), err)
	}
	
	return nil
}

// Manager manages all hooks and dispatches events
type Manager struct {
	mu            sync.RWMutex
	hooks         map[HookType][]HookHandler
	scriptsDir    string
	workspaceDir  string
	eventConfigs  map[HookType]EventConfig
	enabled       bool
}

// EventConfig configuration for a specific hook event type
type EventConfig struct {
	Enabled bool   `json:"enabled"`
	Script  string `json:"script,omitempty"`
	Timeout int    `json:"timeout,omitempty"` // seconds
}

// Config holds the hooks configuration
type Config struct {
	Enabled    bool                   `json:"enabled"`
	ScriptsDir string                 `json:"scripts_dir"`
	Events     map[string]EventConfig `json:"events"`
}

// NewManager creates a new hooks manager
func NewManager(workspaceDir string) *Manager {
	return &Manager{
		hooks:        make(map[HookType][]HookHandler),
		workspaceDir: workspaceDir,
		scriptsDir:   "~/.swifttalon/hooks",
		eventConfigs: make(map[HookType]EventConfig),
		enabled:      false,
	}
}

// Configure applies the hooks configuration
func (m *Manager) Configure(cfg Config) {
	m.mu.Lock()
	defer m.mu.Unlock()
	
	m.enabled = cfg.Enabled
	m.scriptsDir = cfg.ScriptsDir
	
	// Expand ~ in scripts dir
	m.scriptsDir = expandHome(m.scriptsDir)
	
	// Clear existing hooks
	m.hooks = make(map[HookType][]HookHandler)
	m.eventConfigs = make(map[HookType]EventConfig)
	
	// Load configured hooks
	for eventTypeStr, eventCfg := range cfg.Events {
		hookType := HookType(eventTypeStr)
		
		// Validate hook type
		valid := false
		for _, t := range AllHookTypes() {
			if t == hookType {
				valid = true
				break
			}
		}
		if !valid {
			logger.WarnCF("hooks", "Invalid hook type: %s",
				map[string]interface{}{"type": eventTypeStr})
			continue
		}
		
		m.eventConfigs[hookType] = eventCfg
		
		// Skip disabled hooks
		if !eventCfg.Enabled {
			continue
		}
		
		// Create script hook if script path is provided
		if eventCfg.Script != "" {
			scriptPath := eventCfg.Script
			
			// Check if path is relative (workspace-relative)
			if !filepath.IsAbs(scriptPath) {
				// Try workspace-relative first
				workspaceScript := filepath.Join(m.workspaceDir, scriptPath)
				if _, err := os.Stat(workspaceScript); err == nil {
					scriptPath = workspaceScript
				} else {
					// Fall back to scripts directory
					scriptPath = filepath.Join(m.scriptsDir, scriptPath)
				}
			}
			
			// Verify script exists
			if _, err := os.Stat(scriptPath); err != nil {
				logger.WarnCF("hooks", "Hook script not found: %s",
					map[string]interface{}{
						"script": scriptPath,
						"type":   eventTypeStr,
						"error":  err.Error(),
					})
				continue
			}
			
			// Create hook name from script filename
			hookName := filepath.Base(scriptPath)
			hookName = strings.TrimSuffix(hookName, filepath.Ext(hookName))
			
			hook := NewScriptHook(
				hookName,
				scriptPath,
				nil,
				nil,
				"",
				eventCfg.Timeout,
			)
			
			m.registerHook(hookType, hook)
			
			logger.InfoCF("hooks", "Registered hook",
				map[string]interface{}{
					"type":  eventTypeStr,
					"script": scriptPath,
				})
		}
	}
	
	logger.InfoCF("hooks", "Hooks configured",
		map[string]interface{}{
			"enabled":   m.enabled,
			"scripts_dir": m.scriptsDir,
			"hook_count": m.countHooks(),
		})
}

// registerHook registers a handler for a specific hook type
func (m *Manager) registerHook(hookType HookType, handler HookHandler) {
	m.hooks[hookType] = append(m.hooks[hookType], handler)
}

// countHooks returns the total number of registered hooks
func (m *Manager) countHooks() int {
	count := 0
	for _, handlers := range m.hooks {
		count += len(handlers)
	}
	return count
}

// IsEnabled returns whether hooks are enabled
func (m *Manager) IsEnabled() bool {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.enabled
}

// Trigger fires a hook event to all registered handlers for that type
func (m *Manager) Trigger(ctx context.Context, event *HookEvent) {
	m.mu.RLock()
	if !m.enabled {
		m.mu.RUnlock()
		return
	}
	
	handlers, ok := m.hooks[event.Type]
	m.mu.RUnlock()
	
	if !ok || len(handlers) == 0 {
		return
	}
	
	logger.DebugCF("hooks", "Triggering hook",
		map[string]interface{}{
			"type":     event.Type,
			"handlers": len(handlers),
		})
	
	// Execute all handlers (best effort, don't fail on error)
	for _, handler := range handlers {
		if err := handler.Execute(ctx, event); err != nil {
			logger.WarnCF("hooks", "Hook handler error",
				map[string]interface{}{
					"hook":  handler.Name(),
					"type":  event.Type,
					"error": err.Error(),
				})
		}
	}
}

// Helper functions for triggering specific events

// TriggerPreTool fires before tool execution
func (m *Manager) TriggerPreTool(ctx context.Context, toolName string, toolInput map[string]interface{}, sessionKey, channel, chatID string) {
	inputJSON, _ := json.Marshal(toolInput)
	m.Trigger(ctx, &HookEvent{
		Type:       PreTool,
		ToolName:   toolName,
		ToolInput:  inputJSON,
		SessionKey: sessionKey,
		Channel:    channel,
		ChatID:     chatID,
	})
}

// TriggerPostTool fires after tool execution
func (m *Manager) TriggerPostTool(ctx context.Context, toolName string, toolInput map[string]interface{}, toolOutput string, sessionKey, channel, chatID string) {
	inputJSON, _ := json.Marshal(toolInput)
	outputJSON := json.RawMessage(toolOutput)
	if toolOutput == "" {
		outputJSON = nil
	}
	m.Trigger(ctx, &HookEvent{
		Type:       PostTool,
		ToolName:   toolName,
		ToolInput:  inputJSON,
		ToolOutput: outputJSON,
		SessionKey: sessionKey,
		Channel:    channel,
		ChatID:     chatID,
	})
}

// TriggerOnError fires when an error occurs
func (m *Manager) TriggerOnError(ctx context.Context, err error, contextData map[string]interface{}) {
	event := &HookEvent{
		Type:      OnError,
		Error:     err.Error(),
		Metadata:  contextData,
	}
	
	if c, ok := contextData["session_key"].(string); ok {
		event.SessionKey = c
	}
	if c, ok := contextData["channel"].(string); ok {
		event.Channel = c
	}
	if c, ok := contextData["chat_id"].(string); ok {
		event.ChatID = c
	}
	
	m.Trigger(ctx, event)
}

// TriggerOnMessage fires when a message is received
func (m *Manager) TriggerOnMessage(ctx context.Context, message, sessionKey, channel, chatID string) {
	m.Trigger(ctx, &HookEvent{
		Type:       OnMessage,
		Message:    message,
		SessionKey: sessionKey,
		Channel:    channel,
		ChatID:     chatID,
	})
}

// TriggerOnMessageSent fires when a message is sent
func (m *Manager) TriggerOnMessageSent(ctx context.Context, response, sessionKey, channel, chatID string) {
	m.Trigger(ctx, &HookEvent{
		Type:       OnMessageSent,
		Response:   response,
		SessionKey: sessionKey,
		Channel:    channel,
		ChatID:     chatID,
	})
}

// TriggerPreLLM fires before an LLM call
func (m *Manager) TriggerPreLLM(ctx context.Context, sessionKey, channel, chatID string) {
	m.Trigger(ctx, &HookEvent{
		Type:       PreLLM,
		SessionKey: sessionKey,
		Channel:    channel,
		ChatID:     chatID,
	})
}

// TriggerPostLLM fires after an LLM call
func (m *Manager) TriggerPostLLM(ctx context.Context, sessionKey, channel, chatID string, responseContent string) {
	m.Trigger(ctx, &HookEvent{
		Type:       PostLLM,
		Response:   responseContent,
		SessionKey: sessionKey,
		Channel:    channel,
		ChatID:     chatID,
	})
}

// GetRegisteredHooks returns information about registered hooks (for debugging)
func (m *Manager) GetRegisteredHooks() map[string][]string {
	m.mu.RLock()
	defer m.mu.RUnlock()
	
	result := make(map[string][]string)
	for hookType, handlers := range m.hooks {
		hookNames := make([]string, len(handlers))
		for i, h := range handlers {
			hookNames[i] = h.Name()
		}
		result[hookType.String()] = hookNames
	}
	return result
}

func expandHome(path string) string {
	if path == "" {
		return path
	}
	if path[0] == '~' {
		home, _ := os.UserHomeDir()
		if len(path) > 1 && path[1] == '/' {
			return home + path[1:]
		}
		return home
	}
	return path
}

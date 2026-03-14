// PicoClaw - Ultra-lightweight personal AI agent
// Context Window Management - Token tracking and truncation
//
// Copyright (c) 2026 PicoClaw contributors

package agent

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/Bhuw1234/swifttalon/pkg/logger"
	"github.com/Bhuw1234/swifttalon/pkg/providers"
	"github.com/Bhuw1234/swifttalon/pkg/skills"
	"github.com/Bhuw1234/swifttalon/pkg/tools"
)

// TruncationStrategy defines how to handle context when it exceeds the limit
type TruncationStrategy string

const (
	// StrategyRemoveOldest removes the oldest messages first
	StrategyRemoveOldest TruncationStrategy = "remove_oldest"
	// StrategySummarize uses AI to summarize old messages instead of removing
	StrategySummarize TruncationStrategy = "summarize"
	// StrategyHybrid keeps a mix of messages and summarizes the rest
	StrategyHybrid TruncationStrategy = "hybrid"
)

// ModelContextLimits defines context window sizes for common models
// These are approximate values - actual limits may vary
var ModelContextLimits = map[string]int{
	// OpenAI models
	"gpt-4o":        128000,
	"gpt-4o-mini":   128000,
	"gpt-4-turbo":   128000,
	"gpt-4":         8192,
	"gpt-3.5-turbo": 16385,

	// Anthropic models
	"claude-3-5-sonnet": 200000,
	"claude-3-opus":     200000,
	"claude-3-sonnet":   180000,
	"claude-3-haiku":    200000,
	"claude-2.1":        200000,
	"claude-2":          100000,

	// Google models
	"gemini-1.5-pro":     200000,
	"gemini-1.5-flash":   100000,
	"gemini-1.5-flash-8": 1000000,
	"gemini-pro":         32768,

	// Meta/Llama models
	"llama-3.1-405b": 128000,
	"llama-3.1-70b":  128000,
	"llama-3.1-8b":   128000,
	"llama-3-70b":    8192,
	"llama-3-8b":     8192,
	"llama-2-70b":    4096,
	"llama-2-13b":    4096,

	// Mistral models
	"mistral-large":  128000,
	"mistral-medium": 128000,
	"mistral-small": 128000,
	"mistral-7b":     8192,

	// Zhipu models
	"glm-4":       128000,
	"glm-4-flash": 128000,
	"glm-4-plus":  128000,
	"glm-3-turbo": 32000,

	// DeepSeek models
	"deepseek-chat":  64000,
	"deepseek-coder": 16000,

	// Groq models (using their hosted models)
	"mixtral-8x7b": 32768,

	// Ollama local models (conservative defaults)
	"llama2":    4096,
	"codellama": 16384,
	"mistral":   8192,
	"phi3":      4096,

	// NVIDIA NIM
	"nvidia/llama-3.1": 128000,

	// GitHub Copilot
	"claude-sonnet-4-20250514": 200000,
	"gpt-4o-2025-05-14":        128000,
}

// ContextBuilder builds the context for the agent loop
type ContextBuilder struct {
	workspace    string
	skillsLoader *skills.SkillsLoader
	memory       *MemoryStore
	tools        *tools.ToolRegistry // Direct reference to tool registry
}

func getGlobalConfigDir() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return ""
	}
	return filepath.Join(home, ".picoclaw")
}

func NewContextBuilder(workspace string) *ContextBuilder {
	// builtin skills: skills directory in current project
	// Use the skills/ directory under the current working directory
	wd, _ := os.Getwd()
	builtinSkillsDir := filepath.Join(wd, "skills")
	globalSkillsDir := filepath.Join(getGlobalConfigDir(), "skills")

	return &ContextBuilder{
		workspace:    workspace,
		skillsLoader: skills.NewSkillsLoader(workspace, globalSkillsDir, builtinSkillsDir),
		memory:       NewMemoryStore(workspace),
	}
}

// SetToolsRegistry sets the tools registry for dynamic tool summary generation.
func (cb *ContextBuilder) SetToolsRegistry(registry *tools.ToolRegistry) {
	cb.tools = registry
}

func (cb *ContextBuilder) getIdentity() string {
	now := time.Now().Format("2006-01-02 15:04 (Monday)")
	workspacePath, _ := filepath.Abs(filepath.Join(cb.workspace))
	runtime := fmt.Sprintf("%s %s, Go %s", runtime.GOOS, runtime.GOARCH, runtime.Version())

	// Build tools section dynamically
	toolsSection := cb.buildToolsSection()

	return fmt.Sprintf(`# picoclaw 🐙

You are picoclaw, a helpful AI assistant.

## Current Time
%s

## Runtime
%s

## Workspace
Your workspace is at: %s
- Memory: %s/memory/MEMORY.md
- Daily Notes: %s/memory/YYYYMM/YYYYMMDD.md
- Skills: %s/skills/{skill-name}/SKILL.md

%s

## Important Rules

1. **ALWAYS use tools** - When you need to perform an action (schedule reminders, send messages, execute commands, etc.), you MUST call the appropriate tool. Do NOT just say you'll do it or pretend to do it.

2. **Be helpful and accurate** - When using tools, briefly explain what you're doing.

3. **Memory** - When remembering something, write to %s/memory/MEMORY.md`,
		now, runtime, workspacePath, workspacePath, workspacePath, workspacePath, toolsSection, workspacePath)
}

func (cb *ContextBuilder) buildToolsSection() string {
	if cb.tools == nil {
		return ""
	}

	summaries := cb.tools.GetSummaries()
	if len(summaries) == 0 {
		return ""
	}

	var sb strings.Builder
	sb.WriteString("## Available Tools\n\n")
	sb.WriteString("**CRITICAL**: You MUST use tools to perform actions. Do NOT pretend to execute commands or schedule tasks.\n\n")
	sb.WriteString("You have access to the following tools:\n\n")
	for _, s := range summaries {
		sb.WriteString(s)
		sb.WriteString("\n")
	}

	return sb.String()
}

func (cb *ContextBuilder) BuildSystemPrompt() string {
	parts := []string{}

	// Core identity section
	parts = append(parts, cb.getIdentity())

	// Bootstrap files
	bootstrapContent := cb.LoadBootstrapFiles()
	if bootstrapContent != "" {
		parts = append(parts, bootstrapContent)
	}

	// Skills - show summary, AI can read full content with read_file tool
	skillsSummary := cb.skillsLoader.BuildSkillsSummary()
	if skillsSummary != "" {
		parts = append(parts, fmt.Sprintf(`# Skills

The following skills extend your capabilities. To use a skill, read its SKILL.md file using the read_file tool.

%s`, skillsSummary))
	}

	// Memory context
	memoryContext := cb.memory.GetMemoryContext()
	if memoryContext != "" {
		parts = append(parts, "# Memory\n\n"+memoryContext)
	}

	// Join with "---" separator
	return strings.Join(parts, "\n\n---\n\n")
}

func (cb *ContextBuilder) LoadBootstrapFiles() string {
	bootstrapFiles := []string{
		"AGENTS.md",
		"SOUL.md",
		"USER.md",
		"IDENTITY.md",
	}

	var result string
	for _, filename := range bootstrapFiles {
		filePath := filepath.Join(cb.workspace, filename)
		if data, err := os.ReadFile(filePath); err == nil {
			result += fmt.Sprintf("## %s\n\n%s\n\n", filename, string(data))
		}
	}

	return result
}

func (cb *ContextBuilder) BuildMessages(history []providers.Message, summary string, currentMessage string, media []string, channel, chatID string) []providers.Message {
	messages := []providers.Message{}

	systemPrompt := cb.BuildSystemPrompt()

	// Add Current Session info if provided
	if channel != "" && chatID != "" {
		systemPrompt += fmt.Sprintf("\n\n## Current Session\nChannel: %s\nChat ID: %s", channel, chatID)
	}

	// Log system prompt summary for debugging (debug mode only)
	logger.DebugCF("agent", "System prompt built",
		map[string]interface{}{
			"total_chars":   len(systemPrompt),
			"total_lines":   strings.Count(systemPrompt, "\n") + 1,
			"section_count": strings.Count(systemPrompt, "\n\n---\n\n") + 1,
		})

	// Log preview of system prompt (avoid logging huge content)
	preview := systemPrompt
	if len(preview) > 500 {
		preview = preview[:500] + "... (truncated)"
	}
	logger.DebugCF("agent", "System prompt preview",
		map[string]interface{}{
			"preview": preview,
		})

	if summary != "" {
		systemPrompt += "\n\n## Summary of Previous Conversation\n\n" + summary
	}

	//This fix prevents the session memory from LLM failure due to elimination of toolu_IDs required from LLM
	// --- INICIO DEL FIX ---
	//Diegox-17
	for len(history) > 0 && (history[0].Role == "tool") {
		logger.DebugCF("agent", "Removing orphaned tool message from history to prevent LLM error",
			map[string]interface{}{"role": history[0].Role})
		history = history[1:]
	}
	//Diegox-17
	// --- FIN DEL FIX ---

	messages = append(messages, providers.Message{
		Role:    "system",
		Content: systemPrompt,
	})

	messages = append(messages, history...)

	messages = append(messages, providers.Message{
		Role:    "user",
		Content: currentMessage,
	})

	return messages
}

func (cb *ContextBuilder) AddToolResult(messages []providers.Message, toolCallID, toolName, result string) []providers.Message {
	messages = append(messages, providers.Message{
		Role:       "tool",
		Content:    result,
		ToolCallID: toolCallID,
	})
	return messages
}

func (cb *ContextBuilder) AddAssistantMessage(messages []providers.Message, content string, toolCalls []map[string]interface{}) []providers.Message {
	msg := providers.Message{
		Role:    "assistant",
		Content: content,
	}
	// Always add assistant message, whether or not it has tool calls
	messages = append(messages, msg)
	return messages
}

func (cb *ContextBuilder) loadSkills() string {
	allSkills := cb.skillsLoader.ListSkills()
	if len(allSkills) == 0 {
		return ""
	}

	var skillNames []string
	for _, s := range allSkills {
		skillNames = append(skillNames, s.Name)
	}

	content := cb.skillsLoader.LoadSkillsForContext(skillNames)
	if content == "" {
		return ""
	}

	return "# Skill Definitions\n\n" + content
}

// GetSkillsInfo returns information about loaded skills.
func (cb *ContextBuilder) GetSkillsInfo() map[string]interface{} {
	allSkills := cb.skillsLoader.ListSkills()
	skillNames := make([]string, 0, len(allSkills))
	for _, s := range allSkills {
		skillNames = append(skillNames, s.Name)
	}
	return map[string]interface{}{
		"total":     len(allSkills),
		"available": len(allSkills),
		"names":     skillNames,
	}
}

// ContextManager handles token counting and context truncation
type ContextManager struct {
	maxContextTokens int
	strategy         TruncationStrategy
	model            string
	charPerToken     float64 // Characters per token ratio for estimation
}

// NewContextManager creates a new ContextManager with the specified limits
func NewContextManager(maxContextTokens int, strategy string, model string) *ContextManager {
	// If maxContextTokens is 0, try to get from model defaults
	if maxContextTokens == 0 {
		maxContextTokens = GetModelContextLimit(model)
	}

	// Default to 100K if still 0
	if maxContextTokens == 0 {
		maxContextTokens = 100000
	}

	// Parse strategy
	truncationStrategy := StrategyRemoveOldest
	switch strings.ToLower(strategy) {
	case "summarize":
		truncationStrategy = StrategySummarize
	case "hybrid":
		truncationStrategy = StrategyHybrid
	}

	// Determine char per token ratio based on model
	// Some models use more tokens for the same text (e.g., multilingual models)
	charPerToken := getModelCharPerToken(model)

	return &ContextManager{
		maxContextTokens: maxContextTokens,
		strategy:         truncationStrategy,
		model:            model,
		charPerToken:     charPerToken,
	}
}

// GetModelContextLimit returns the context window limit for a given model
func GetModelContextLimit(model string) int {
	// Try exact match first
	if limit, ok := ModelContextLimits[model]; ok {
		return limit
	}

	// Try prefix matching (e.g., "glm-4.7" matches "glm-4")
	modelLower := strings.ToLower(model)
	for known, limit := range ModelContextLimits {
		if strings.HasPrefix(modelLower, strings.ToLower(known)) {
			return limit
		}
	}

	return 0
}

// getModelCharPerToken returns the characters per token ratio for a model
// This affects token estimation accuracy
func getModelCharPerToken(model string) float64 {
	modelLower := strings.ToLower(model)

	// Chinese/ multilingual models typically have lower ratio
	if strings.Contains(modelLower, "glm") ||
		strings.Contains(modelLower, "claude") ||
		strings.Contains(modelLower, "qwen") ||
		strings.Contains(modelLower, "baichuan") {
		// These models often encode more characters per token for CJK text
		return 2.0
	}

	// Default ratio (approximately 4 chars per token for English)
	return 4.0
}

// EstimateTokens estimates the number of tokens in a message
func (cm *ContextManager) EstimateTokens(msg providers.Message) int {
	// Count content characters
	contentLen := utf8.RuneCountInString(msg.Content)

	// Add overhead for role
	roleLen := len(msg.Role) + 2 // ": " separator

	// Add overhead for tool calls if present
	toolCallOverhead := 0
	for _, tc := range msg.ToolCalls {
		toolCallOverhead += len(tc.Name) + 20 // Function name + overhead
		if tc.Function != nil {
			toolCallOverhead += utf8.RuneCountInString(tc.Function.Arguments)
		}
	}

	totalChars := contentLen + roleLen + toolCallOverhead

	// Estimate tokens using the model's char per token ratio
	return int(float64(totalChars) / cm.charPerToken)
}

// EstimateMessagesTokens estimates total tokens in a message list
func (cm *ContextManager) EstimateMessagesTokens(messages []providers.Message) int {
	total := 0
	for _, msg := range messages {
		total += cm.EstimateTokens(msg)
	}
	return total
}

// GetTokenCount returns the current context usage
func (cm *ContextManager) GetTokenCount(messages []providers.Message) int {
	return cm.EstimateMessagesTokens(messages)
}

// GetMaxTokens returns the maximum context window size
func (cm *ContextManager) GetMaxTokens() int {
	return cm.maxContextTokens
}

// GetStrategy returns the current truncation strategy
func (cm *ContextManager) GetStrategy() TruncationStrategy {
	return cm.strategy
}

// NeedsTruncation checks if the messages exceed the context limit
// Returns true if tokens exceed the limit, also returns the token count
func (cm *ContextManager) NeedsTruncation(messages []providers.Message) (bool, int) {
	tokenCount := cm.EstimateMessagesTokens(messages)
	return tokenCount > cm.maxContextTokens, tokenCount
}

// GetSafeTokenLimit returns the token limit with a safety margin (80%)
// This预留 space for response generation
func (cm *ContextManager) GetSafeTokenLimit() int {
	return cm.maxContextTokens * 80 / 100
}

// TruncateMessages truncates messages based on the configured strategy
// Returns the truncated messages and info about what was truncated
func (cm *ContextManager) TruncateMessages(messages []providers.Message) ([]providers.Message, *TruncationInfo, error) {
	if len(messages) == 0 {
		return messages, &TruncationInfo{}, nil
	}

	tokenCount := cm.EstimateMessagesTokens(messages)
	if tokenCount <= cm.maxContextTokens {
		return messages, &TruncationInfo{
			OriginalTokens: tokenCount,
			KeptTokens:     tokenCount,
			Strategy:       cm.strategy,
		}, nil
	}

	logger.InfoCF("context", "Context limit exceeded, applying truncation",
		map[string]interface{}{
			"tokens":     tokenCount,
			"max_tokens": cm.maxContextTokens,
			"msg_count":  len(messages),
			"strategy":   cm.strategy,
		})

	switch cm.strategy {
	case StrategyRemoveOldest:
		return cm.truncateRemoveOldest(messages)
	case StrategySummarize:
		// Summarize is handled by the agent loop's summarization
		// Here we do a simple truncation as fallback
		return cm.truncateRemoveOldest(messages)
	case StrategyHybrid:
		return cm.truncateHybrid(messages)
	default:
		return cm.truncateRemoveOldest(messages)
	}
}

// TruncationInfo contains information about a truncation operation
type TruncationInfo struct {
	OriginalTokens   int
	KeptTokens       int
	RemovedTokens    int
	RemovedMessages  int
	KeptMessages     int
	Strategy         TruncationStrategy
	CompressionRatio float64
}

// truncateRemoveOldest removes the oldest messages to fit within the limit
// Always keeps the first message (system prompt) and the most recent messages
func (cm *ContextManager) truncateRemoveOldest(messages []providers.Message) ([]providers.Message, *TruncationInfo, error) {
	if len(messages) <= 1 {
		return messages, &TruncationInfo{
			OriginalTokens: cm.EstimateMessagesTokens(messages),
			KeptTokens:     cm.EstimateMessagesTokens(messages),
			Strategy:       cm.strategy,
		}, nil
	}

	originalTokens := cm.EstimateMessagesTokens(messages)
	targetTokens := cm.GetSafeTokenLimit()

	// Always keep the system prompt (first message)
	systemMsg := messages[0]
	systemTokens := cm.EstimateTokens(systemMsg)

	// Calculate how many tokens we can use for conversation
	availableTokens := targetTokens - systemTokens
	if availableTokens < 0 {
		availableTokens = targetTokens / 2
	}

	// Keep: system prompt + as many recent messages as fit
	result := []providers.Message{systemMsg}

	// Start from the end and work backwards
	// Always keep at least the last message
	for i := len(messages) - 1; i > 0 && len(result) < len(messages); i-- {
		msg := messages[i]
		msgTokens := cm.EstimateTokens(msg)

		// Check if adding this message would exceed limit
		currentTokens := cm.EstimateMessagesTokens(result)
		if currentTokens+msgTokens > availableTokens {
			// Don't add more messages, but check if we should keep at least 2 messages
			// (user message + assistant response pattern)
			if len(result) < 3 && i < len(messages)-1 {
				// Try to keep at least the last message
				result = append(result, messages[len(messages)-1])
			}
			break
		}

		// Prepend the message (to maintain order)
		result = append([]providers.Message{msg}, result[1:]...)
	}

	keptTokens := cm.EstimateMessagesTokens(result)
	removedCount := len(messages) - len(result)

	logger.InfoCF("context", "Truncated messages (remove_oldest)",
		map[string]interface{}{
			"original_msgs":   len(messages),
			"kept_msgs":       len(result),
			"removed_msgs":    removedCount,
			"original_tokens": originalTokens,
			"kept_tokens":     keptTokens,
			"target_tokens":   targetTokens,
		})

	return result, &TruncationInfo{
		OriginalTokens:   originalTokens,
		KeptTokens:       keptTokens,
		RemovedTokens:    originalTokens - keptTokens,
		RemovedMessages:  removedCount,
		KeptMessages:     len(result),
		Strategy:         cm.strategy,
		CompressionRatio: float64(keptTokens) / float64(originalTokens),
	}, nil
}

// truncateHybrid keeps the most recent messages and marks older ones as summarized
// It keeps more recent context while noting that older context was summarized
func (cm *ContextManager) truncateHybrid(messages []providers.Message) ([]providers.Message, *TruncationInfo, error) {
	if len(messages) <= 1 {
		return messages, &TruncationInfo{
			OriginalTokens: cm.EstimateMessagesTokens(messages),
			KeptTokens:     cm.EstimateMessagesTokens(messages),
			Strategy:       cm.strategy,
		}, nil
	}

	originalTokens := cm.EstimateMessagesTokens(messages)
	targetTokens := cm.GetSafeTokenLimit()

	// Always keep system prompt
	systemMsg := messages[0]
	systemTokens := cm.EstimateTokens(systemMsg)

	// For hybrid: keep more recent messages (75% of available space)
	recentTokens := (targetTokens - systemTokens) * 75 / 100

	result := []providers.Message{systemMsg}

	// Count conversation messages (excluding system)
	conversationMsgs := messages[1:]

	// Add summary placeholder if there are many messages
	if len(conversationMsgs) > 10 {
		summaryMsg := providers.Message{
			Role:    "system",
			Content: "[Previous conversation has been summarized. See summary above.]",
		}
		result = append(result, summaryMsg)
	}

	// Add recent messages
	for i := len(conversationMsgs) - 1; i >= 0 && len(result) < len(messages); i-- {
		msg := conversationMsgs[i]
		msgTokens := cm.EstimateTokens(msg)

		currentTokens := cm.EstimateMessagesTokens(result)
		if currentTokens+msgTokens > recentTokens {
			continue
		}

		result = append(result, msg)
	}

	keptTokens := cm.EstimateMessagesTokens(result)
	removedCount := len(messages) - len(result)

	logger.InfoCF("context", "Truncated messages (hybrid)",
		map[string]interface{}{
			"original_msgs":   len(messages),
			"kept_msgs":       len(result),
			"removed_msgs":    removedCount,
			"original_tokens": originalTokens,
			"kept_tokens":     keptTokens,
		})

	return result, &TruncationInfo{
		OriginalTokens:   originalTokens,
		KeptTokens:       keptTokens,
		RemovedTokens:    originalTokens - keptTokens,
		RemovedMessages:  removedCount,
		KeptMessages:     len(result),
		Strategy:         cm.strategy,
		CompressionRatio: float64(keptTokens) / float64(originalTokens),
	}, nil
}

// GetContextInfo returns a formatted string with context usage information
func (cm *ContextManager) GetContextInfo(messages []providers.Message) string {
	tokenCount := cm.EstimateMessagesTokens(messages)
	percentage := float64(tokenCount) / float64(cm.maxContextTokens) * 100

	return fmt.Sprintf("Context: %d / %d tokens (%.1f%%)",
		tokenCount, cm.maxContextTokens, percentage)
}

// PrepareMessagesForProvider prepares messages with context management
// This should be called before sending to the LLM provider
func (cm *ContextManager) PrepareMessagesForProvider(messages []providers.Message) ([]providers.Message, *TruncationInfo, error) {
	// First check if we need to truncate
	needsTruncation, tokenCount := cm.NeedsTruncation(messages)

	if !needsTruncation {
		return messages, &TruncationInfo{
			OriginalTokens: tokenCount,
			KeptTokens:     tokenCount,
			Strategy:       cm.strategy,
		}, nil
	}

	// Apply truncation
	return cm.TruncateMessages(messages)
}

// EstimateTokenCount is a simple utility function to estimate tokens
// Uses a default ratio of 4 characters per token
func EstimateTokenCount(text string) int {
	if text == "" {
		return 0
	}
	charCount := utf8.RuneCountInString(text)
	return charCount / 4
}

// EstimateTokenCountForMessages estimates total tokens for a message list
// Using a simple formula
func EstimateTokenCountForMessages(messages []providers.Message) int {
	total := 0
	for _, msg := range messages {
		// Content
		total += EstimateTokenCount(msg.Content)
		// Role overhead
		total += len(msg.Role) / 4
		// Tool calls overhead
		for _, tc := range msg.ToolCalls {
			total += len(tc.Name) / 4
			if tc.Function != nil {
				total += EstimateTokenCount(tc.Function.Arguments)
			}
		}
	}
	return total
}
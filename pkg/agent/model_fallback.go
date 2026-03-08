// PicoClaw - Ultra-lightweight personal AI agent
// Model Fallback System - Automatically tries next model when one fails
//
// Copyright (c) 2026 PicoClaw contributors

package agent

import (
	"context"
	"fmt"
	"strings"
	"sync"

	"github.com/sipeed/picoclaw/pkg/config"
	"github.com/sipeed/picoclaw/pkg/logger"
	"github.com/sipeed/picoclaw/pkg/providers"
)

// FallbackAttempt records a single fallback attempt
type FallbackAttempt struct {
	Model   string
	Error   string
	Attempt int
}

// FallbackProvider wraps an LLMProvider with automatic fallback to backup models.
// When the primary model fails, it automatically tries the next model in the fallback chain.
type FallbackProvider struct {
	mu           sync.RWMutex
	baseProvider providers.LLMProvider
	fallbacks    []string
	currentModel string
	maxRetries   int
	attempts     []FallbackAttempt
}

// NewFallbackProvider creates a new fallback provider wrapping the given base provider
func NewFallbackProvider(base providers.LLMProvider, fallbacks []string) *FallbackProvider {
	fp := &FallbackProvider{
		baseProvider: base,
		fallbacks:    fallbacks,
		maxRetries:   len(fallbacks),
		attempts:     make([]FallbackAttempt, 0),
	}

	// Set initial model from base provider
	fp.currentModel = base.GetDefaultModel()

	logger.InfoCF("model_fallback", "Fallback provider initialized",
		map[string]interface{}{
			"primary_model": fp.currentModel,
			"fallbacks":    fallbacks,
			"fallback_len": len(fallbacks),
		})

	return fp
}

// Chat attempts to chat with the model, automatically trying fallbacks on failure
func (fp *FallbackProvider) Chat(ctx context.Context, messages []providers.Message, tools []providers.ToolDefinition, model string, options map[string]interface{}) (*providers.LLMResponse, error) {
	fp.mu.Lock()
	fp.attempts = fp.attempts[:0] // Clear previous attempts
	fp.mu.Unlock()

	// Build the list of models to try: primary first, then fallbacks
	modelsToTry := fp.buildModelList(model)

	logger.InfoCF("model_fallback", "Starting chat with fallback chain",
		map[string]interface{}{
			"primary_model":    model,
			"total_models":    len(modelsToTry),
			"fallback_models": fp.fallbacks,
		})

	for i, modelToTry := range modelsToTry {
		// Update current model
		fp.mu.Lock()
		fp.currentModel = modelToTry
		fp.mu.Unlock()

		logger.InfoCF("model_fallback", "Attempting model",
			map[string]interface{}{
				"model":      modelToTry,
				"attempt":    i + 1,
				"total":      len(modelsToTry),
				"is_fallback": i > 0,
			})

		// Try the chat call
		response, err := fp.baseProvider.Chat(ctx, messages, tools, modelToTry, options)

		if err == nil {
			// Success! Log and return
			if i > 0 {
				logger.InfoCF("model_fallback", "Fallback successful",
					map[string]interface{}{
						"model_used":    modelToTry,
						"failed_models": fp.getAttemptedModels(),
					})
			}
			return response, nil
		}

		// Failure - record the attempt and try next model
		errMsg := err.Error()

		// Check if this is a "retryable" error (not user abort, etc.)
		if !fp.isRetryableError(err) {
			logger.WarnCF("model_fallback", "Non-retryable error, skipping fallback",
				map[string]interface{}{
					"model": modelToTry,
					"error": errMsg,
				})
			return nil, err
		}

		// Record the failed attempt
		fp.mu.Lock()
		fp.attempts = append(fp.attempts, FallbackAttempt{
			Model:   modelToTry,
			Error:   errMsg,
			Attempt: i + 1,
		})
		fp.mu.Unlock()

		logger.WarnCF("model_fallback", "Model failed, trying next",
			map[string]interface{}{
				"model":        modelToTry,
				"error":        errMsg,
				"attempt":      i + 1,
				"remaining":    len(modelsToTry) - i - 1,
				"failed_count": i + 1,
			})
	}

	// All models failed
	summary := fp.getFailureSummary()
	return nil, fmt.Errorf("all models failed (%d attempted): %s", len(modelsToTry), summary)
}

// GetDefaultModel returns the current model being used
func (fp *FallbackProvider) GetDefaultModel() string {
	fp.mu.RLock()
	defer fp.mu.RUnlock()
	return fp.currentModel
}

// GetAttempts returns the list of attempts made
func (fp *FallbackProvider) GetAttempts() []FallbackAttempt {
	fp.mu.RLock()
	defer fp.mu.RUnlock()
	// Return a copy
	result := make([]FallbackAttempt, len(fp.attempts))
	copy(result, fp.attempts)
	return result
}

// GetFallbacks returns the configured fallback models
func (fp *FallbackProvider) GetFallbacks() []string {
	fp.mu.RLock()
	defer fp.mu.RUnlock()
	result := make([]string, len(fp.fallbacks))
	copy(result, fp.fallbacks)
	return result
}

// buildModelList builds the complete list of models to try (primary + fallbacks)
func (fp *FallbackProvider) buildModelList(primaryModel string) []string {
	fp.mu.RLock()
	defer fp.mu.RUnlock()

	models := []string{primaryModel}

	// Add fallbacks, avoiding duplicates
	seen := map[string]bool{primaryModel: true}
	for _, fb := range fp.fallbacks {
		if !seen[fb] {
			models = append(models, fb)
			seen[fb] = true
		}
	}

	return models
}

// isRetryableError determines if an error should trigger a fallback
func (fp *FallbackProvider) isRetryableError(err error) bool {
	if err == nil {
		return false
	}

	errMsg := strings.ToLower(err.Error())

	// Common retryable error patterns
	retryablePatterns := []string{
		"rate limit",
		"rate_limit",
		"too many requests",
		"429",
		"timeout",
		"timeout",
		"connection",
		"network",
		"temporary",
		"unavailable",
		"service unavailable",
		"internal error",
		"internal_error",
		"500",
		"502",
		"503",
		"504",
		"context deadline",
		"i/o timeout",
		"EOF",
		"broken pipe",
		"reset by peer",
		"no such host",
		"connection refused",
		"connection reset",
		"token",
		"context window",
		"max tokens",
		"length",
		"invalidparameter",
		"quota",
		"billing",
		"insufficient",
	}

	for _, pattern := range retryablePatterns {
		if strings.Contains(errMsg, pattern) {
			return true
		}
	}

	return false
}

// getAttemptedModels returns the list of models that have been attempted
func (fp *FallbackProvider) getAttemptedModels() []string {
	fp.mu.RLock()
	defer fp.mu.RUnlock()

	models := make([]string, 0, len(fp.attempts))
	for _, a := range fp.attempts {
		models = append(models, a.Model)
	}
	return models
}

// getFailureSummary returns a human-readable summary of failures
func (fp *FallbackProvider) getFailureSummary() string {
	fp.mu.RLock()
	defer fp.mu.RUnlock()

	if len(fp.attempts) == 0 {
		return "unknown"
	}

	var parts []string
	for _, a := range fp.attempts {
		parts = append(parts, fmt.Sprintf("%s: %s", a.Model, a.Error))
	}

	return strings.Join(parts, " | ")
}

// Ensure FallbackProvider implements LLMProvider interface
var _ providers.LLMProvider = (*FallbackProvider)(nil)

// CreateProviderWithFallback creates a provider with fallback support if configured
func CreateProviderWithFallback(cfg *config.Config, baseProvider providers.LLMProvider) providers.LLMProvider {
	fallbacks := cfg.Agents.Defaults.ModelFallbacks

	// Filter out empty fallbacks
	actualFallbacks := make([]string, 0)
	for _, fb := range fallbacks {
		if strings.TrimSpace(fb) != "" {
			actualFallbacks = append(actualFallbacks, strings.TrimSpace(fb))
		}
	}

	if len(actualFallbacks) == 0 {
		// No fallbacks configured, return base provider as-is
		return baseProvider
	}

	// Wrap with fallback provider
	return NewFallbackProvider(baseProvider, actualFallbacks)
}

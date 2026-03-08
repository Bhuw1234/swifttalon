// PicoClaw - Ultra-lightweight personal AI agent
// Auth Profile Provider - Multi-key provider with automatic rotation
//
// Copyright (c) 2026 PicoClaw contributors

package providers

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/sipeed/picoclaw/pkg/auth"
	"github.com/sipeed/picoclaw/pkg/config"
)

// AuthProfileProvider wraps an LLMProvider with profile rotation
type AuthProfileProvider struct {
	mu            sync.RWMutex
	providerName  string
	apiBase       string
	proxy         string
	store         *auth.AuthProfileStore
	currentIndex  int
	profiles      []*auth.AuthProfile
	wrapped       LLMProvider
	maxRetries    int
	retryDelay    time.Duration
}

// NewAuthProfileProvider creates a new auth profile provider
func NewAuthProfileProvider(
	providerName string,
	cfg *config.ProviderConfig,
	store *auth.AuthProfileStore,
) (*AuthProfileProvider, error) {
	if store == nil {
		store = &auth.AuthProfileStore{}
	}

	p := &AuthProfileProvider{
		providerName: providerName,
		apiBase:     cfg.APIBase,
		proxy:       cfg.Proxy,
		store:       store,
		currentIndex: 0,
		maxRetries:   3,
		retryDelay:   1 * time.Second,
	}

	// Load profiles from config and store
	if err := p.loadProfiles(cfg); err != nil {
		return nil, fmt.Errorf("loading profiles: %w", err)
	}

	// Try to create provider with first available profile
	if err := p.initializeProvider(); err != nil {
		return nil, fmt.Errorf("initializing provider: %w", err)
	}

	return p, nil
}

// loadProfiles loads profiles from config and merges with store
func (p *AuthProfileProvider) loadProfiles(cfg *config.ProviderConfig) error {
	// First, add profiles from config to store
	for i, profileCfg := range cfg.Profiles {
		profileID := fmt.Sprintf("%s:%s", p.providerName, profileCfg.Name)
		if profileCfg.Name == "" {
			profileID = fmt.Sprintf("%s:config-%d", p.providerName, i)
		}

		profile := &auth.AuthProfile{
			ID:       profileID,
			Name:     profileCfg.Name,
			Provider: p.providerName,
			Type:     auth.CredentialTypeAPIKey,
			APIKey:   profileCfg.APIKey,
			Weight:   profileCfg.Weight,
			Disabled: profileCfg.Disabled,
		}
		p.store.AddProfile(profile)
	}

	// Also add legacy single key as default profile if no profiles exist
	if len(cfg.Profiles) == 0 && cfg.APIKey != "" {
		profileID := fmt.Sprintf("%s:default", p.providerName)
		profile := &auth.AuthProfile{
			ID:       profileID,
			Name:     "default",
			Provider: p.providerName,
			Type:     auth.CredentialTypeAPIKey,
			APIKey:   cfg.APIKey,
			Weight:   10,
		}
		p.store.AddProfile(profile)
	}

	// Get all available profiles for this provider
	p.profiles = p.store.GetProfilesForProvider(p.providerName)
	if len(p.profiles) == 0 {
		return fmt.Errorf("no profiles available for provider: %s", p.providerName)
	}

	return nil
}

// initializeProvider creates the wrapped provider with the current profile
func (p *AuthProfileProvider) initializeProvider() error {
	if len(p.profiles) == 0 {
		return fmt.Errorf("no profiles available")
	}

	if p.currentIndex >= len(p.profiles) {
		p.currentIndex = 0
	}

	profile := p.profiles[p.currentIndex]
	apiKey := profile.APIKey

	// Create HTTP provider with current profile's key
	wrapped, err := p.createWrappedProvider(apiKey)
	if err != nil {
		return fmt.Errorf("creating wrapped provider: %w", err)
	}

	p.wrapped = wrapped
	return nil
}

// createWrappedProvider creates a new HTTP provider with the given API key
func (p *AuthProfileProvider) createWrappedProvider(apiKey string) (LLMProvider, error) {
	var apiBase string

	// Resolve API base based on provider
	switch strings.ToLower(p.providerName) {
	case "openai":
		apiBase = p.apiBase
		if apiBase == "" {
			apiBase = "https://api.openai.com/v1"
		}
	case "openrouter":
		apiBase = p.apiBase
		if apiBase == "" {
			apiBase = "https://openrouter.ai/api/v1"
		}
	case "anthropic":
		apiBase = p.apiBase
		if apiBase == "" {
			apiBase = "https://api.anthropic.com/v1"
		}
	case "groq":
		apiBase = p.apiBase
		if apiBase == "" {
			apiBase = "https://api.groq.com/openai/v1"
		}
	case "gemini", "google":
		apiBase = p.apiBase
		if apiBase == "" {
			apiBase = "https://generativelanguage.googleapis.com/v1beta"
		}
	case "zhipu", "glm":
		apiBase = p.apiBase
		if apiBase == "" {
			apiBase = "https://open.bigmodel.cn/api/paas/v4"
		}
	case "moonshot":
		apiBase = p.apiBase
		if apiBase == "" {
			apiBase = "https://api.moonshot.cn/v1"
		}
	case "deepseek":
		apiBase = p.apiBase
		if apiBase == "" {
			apiBase = "https://api.deepseek.com/v1"
		}
	case "vllm":
		apiBase = p.apiBase
	case "ollama":
		apiBase = p.apiBase
		if apiBase == "" {
			apiBase = "http://localhost:11434/v1"
		}
	case "nvidia":
		apiBase = p.apiBase
		if apiBase == "" {
			apiBase = "https://integrate.api.nvidia.com/v1"
		}
	case "shengsuanyun":
		apiBase = p.apiBase
		if apiBase == "" {
			apiBase = "https://router.shengsuanyun.com/api/v1"
		}
	default:
		apiBase = p.apiBase
	}

	if apiKey == "" {
		return nil, fmt.Errorf("no API key available for profile")
	}

	return NewHTTPProvider(apiKey, apiBase, p.proxy), nil
}

// Chat implements the LLMProvider interface with automatic profile rotation
func (p *AuthProfileProvider) Chat(ctx context.Context, messages []Message, tools []ToolDefinition, model string, options map[string]interface{}) (*LLMResponse, error) {
	var lastErr error

	for attempt := 0; attempt < p.maxRetries; attempt++ {
		// Ensure we have a valid provider
		p.mu.RLock()
		wrapped := p.wrapped
		currentIndex := p.currentIndex
		p.mu.RUnlock()

		if wrapped == nil {
			p.mu.Lock()
			if err := p.initializeProvider(); err != nil {
				p.mu.Unlock()
				return nil, fmt.Errorf("no provider available: %w", err)
			}
			wrapped = p.wrapped
			p.mu.Unlock()
		}

		// Try to chat with current profile
		response, err := wrapped.Chat(ctx, messages, tools, model, options)
		if err == nil {
			// Success - mark profile as good and return
			profile := p.profiles[currentIndex]
			p.store.MarkProfileGood(p.providerName, profile.ID)
			p.store.MarkProfileUsed(profile.ID)
			_ = p.store.Save()
			return response, nil
		}

		// Handle error - determine if we should switch profiles
		lastErr = err
		if shouldSwitchProfile(err) {
			p.switchToNextProfile(err)
		}

		// Wait before retry
		if attempt < p.maxRetries-1 {
			time.Sleep(p.retryDelay * time.Duration(attempt+1))
		}
	}

	return nil, fmt.Errorf("all profiles exhausted, last error: %w", lastErr)
}

// shouldSwitchProfile determines if an error should trigger profile rotation
func shouldSwitchProfile(err error) bool {
	if err == nil {
		return false
	}

	errMsg := strings.ToLower(err.Error())

	// Rate limit errors
	if strings.Contains(errMsg, "rate limit") ||
		strings.Contains(errMsg, "429") ||
		strings.Contains(errMsg, "too many requests") {
		return true
	}

	// Auth errors
	if strings.Contains(errMsg, "unauthorized") ||
		strings.Contains(errMsg, "401") ||
		strings.Contains(errMsg, "invalid api key") ||
		strings.Contains(errMsg, "authentication failed") ||
		strings.Contains(errMsg, "forbidden") ||
		strings.Contains(errMsg, "403") {
		return true
	}

	// Billing errors
	if strings.Contains(errMsg, "billing") ||
		strings.Contains(errMsg, "insufficient quota") ||
		strings.Contains(errMsg, "credit limit") ||
		strings.Contains(errMsg, "exceeded") {
		return true
	}

	return false
}

// getFailureReason determines the failure reason from the error
func getFailureReason(err error) auth.ProfileFailureReason {
	if err == nil {
		return auth.FailureReasonUnknown
	}

	errMsg := strings.ToLower(err.Error())

	if strings.Contains(errMsg, "rate limit") ||
		strings.Contains(errMsg, "429") ||
		strings.Contains(errMsg, "too many requests") {
		return auth.FailureReasonRateLimit
	}

	if strings.Contains(errMsg, "unauthorized") ||
		strings.Contains(errMsg, "401") ||
		strings.Contains(errMsg, "invalid api key") ||
		strings.Contains(errMsg, "authentication failed") ||
		strings.Contains(errMsg, "forbidden") ||
		strings.Contains(errMsg, "403") {
		return auth.FailureReasonAuth
	}

	if strings.Contains(errMsg, "billing") ||
		strings.Contains(errMsg, "insufficient quota") ||
		strings.Contains(errMsg, "credit limit") ||
		strings.Contains(errMsg, "exceeded") {
		return auth.FailureReasonBilling
	}

	if strings.Contains(errMsg, "timeout") ||
		strings.Contains(errMsg, "timed out") {
		return auth.FailureReasonTimeout
	}

	return auth.FailureReasonUnknown
}

// switchToNextProfile rotates to the next available profile
func (p *AuthProfileProvider) switchToNextProfile(err error) {
	p.mu.Lock()
	defer p.mu.Unlock()

	if len(p.profiles) <= 1 {
		// No other profiles to switch to
		return
	}

	// Mark current profile as failed
	oldIndex := p.currentIndex
	profile := p.profiles[oldIndex]
	reason := getFailureReason(err)
	p.store.MarkProfileFailure(profile.ID, reason)

	// Get ordered list of profiles
	order := p.store.ResolveAuthProfileOrder(p.providerName, nil)

	// Find next available profile
	for _, profileID := range order {
		if profileID == profile.ID {
			continue
		}

		// Find the index of this profile
		for i, prof := range p.profiles {
			if prof.ID == profileID {
				p.currentIndex = i

				// Reinitialize provider with new profile
				if err := p.initializeProvider(); err != nil {
					continue
				}

				_ = p.store.Save()
				return
			}
		}
	}

	// If we get here, no valid profiles found - try to save anyway
	_ = p.store.Save()
}

// GetDefaultModel returns the default model for this provider
func (p *AuthProfileProvider) GetDefaultModel() string {
	p.mu.RLock()
	defer p.mu.RUnlock()

	if p.wrapped != nil {
		return p.wrapped.GetDefaultModel()
	}
	return ""
}

// GetCurrentProfile returns the currently active profile
func (p *AuthProfileProvider) GetCurrentProfile() *auth.AuthProfile {
	p.mu.RLock()
	defer p.mu.RUnlock()

	if p.currentIndex < len(p.profiles) {
		return p.profiles[p.currentIndex]
	}
	return nil
}

// GetAllProfiles returns all available profiles
func (p *AuthProfileProvider) GetAllProfiles() []*auth.AuthProfile {
	p.mu.RLock()
	defer p.mu.RUnlock()

	profiles := make([]*auth.AuthProfile, len(p.profiles))
	copy(profiles, p.profiles)
	return profiles
}

// ForceProfile switches to a specific profile by name
func (p *AuthProfileProvider) ForceProfile(profileName string) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	for i, profile := range p.profiles {
		if profile.Name == profileName {
			p.currentIndex = i
			if err := p.initializeProvider(); err != nil {
				return fmt.Errorf("switching to profile %s: %w", profileName, err)
			}
			p.store.MarkProfileGood(p.providerName, profile.ID)
			_ = p.store.Save()
			return nil
		}
	}

	return fmt.Errorf("profile not found: %s", profileName)
}

// ResetProfileCooldowns clears all profile cooldowns
func (p *AuthProfileProvider) ResetProfileCooldowns() error {
	p.mu.Lock()
	defer p.mu.Unlock()

	for _, profile := range p.profiles {
		p.store.ClearProfileCooldown(profile.ID)
	}

	return p.store.Save()
}

// CreateAuthProfileProvider creates an auth profile provider from config
func CreateAuthProfileProvider(providerName string, cfg *config.Config) (LLMProvider, error) {
	store, err := auth.LoadAuthProfileStore()
	if err != nil {
		return nil, fmt.Errorf("loading auth profile store: %w", err)
	}

	// Get provider config
	var providerCfg *config.ProviderConfig
	switch strings.ToLower(providerName) {
	case "openai":
		providerCfg = &cfg.Providers.OpenAI
	case "openrouter":
		providerCfg = &cfg.Providers.OpenRouter
	case "anthropic":
		providerCfg = &cfg.Providers.Anthropic
	case "groq":
		providerCfg = &cfg.Providers.Groq
	case "gemini", "google":
		providerCfg = &cfg.Providers.Gemini
	case "zhipu", "glm":
		providerCfg = &cfg.Providers.Zhipu
	case "moonshot":
		providerCfg = &cfg.Providers.Moonshot
	case "deepseek":
		providerCfg = &cfg.Providers.DeepSeek
	case "vllm":
		providerCfg = &cfg.Providers.VLLM
	case "ollama":
		providerCfg = &cfg.Providers.Ollama
	case "nvidia":
		providerCfg = &cfg.Providers.Nvidia
	case "shengsuanyun":
		providerCfg = &cfg.Providers.ShengSuanYun
	default:
		return nil, fmt.Errorf("unknown provider: %s", providerName)
	}

	// Check if profiles are configured
	if len(providerCfg.Profiles) == 0 && providerCfg.APIKey == "" {
		return nil, fmt.Errorf("no API key or profiles configured for provider: %s", providerName)
	}

	return NewAuthProfileProvider(providerName, providerCfg, store)
}

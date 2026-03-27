// SwiftTalon - Ultra-lightweight personal AI agent
// Auth Profiles System - Multi-key management with automatic rotation
//
// Copyright (c) 2026 SwiftTalon contributors

package auth

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"sync"
	"time"
)

// ProfileFailureReason represents why a profile failed
type ProfileFailureReason string

const (
	FailureReasonAuth      ProfileFailureReason = "auth"
	FailureReasonRateLimit ProfileFailureReason = "rate_limit"
	FailureReasonBilling   ProfileFailureReason = "billing"
	FailureReasonTimeout   ProfileFailureReason = "timeout"
	FailureReasonUnknown   ProfileFailureReason = "unknown"
)

// CredentialType represents the type of credential
type CredentialType string

const (
	CredentialTypeAPIKey CredentialType = "api_key"
	CredentialTypeToken  CredentialType = "token"
	CredentialTypeOAuth  CredentialType = "oauth"
)

// AuthProfile represents a single authentication profile
type AuthProfile struct {
	ID        string          `json:"id"`
	Name      string          `json:"name"`
	Provider  string          `json:"provider"`
	Type      CredentialType `json:"type"`
	APIKey    string          `json:"api_key,omitempty"`
	Token     string          `json:"token,omitempty"`
	Email     string          `json:"email,omitempty"`
	Weight    int             `json:"weight"` // For load balancing, higher = more preferred
	Disabled  bool            `json:"disabled"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
}

// ProfileUsageStats tracks usage statistics for a profile
type ProfileUsageStats struct {
	LastUsed      time.Time            `json:"last_used"`
	ErrorCount    int                  `json:"error_count"`
	FailureCounts map[string]int       `json:"failure_counts,omitempty"`
	LastFailureAt time.Time           `json:"last_failure_at"`
	CooldownUntil time.Time           `json:"cooldown_until,omitempty"`
	DisabledUntil time.Time           `json:"disabled_until,omitempty"`
	DisabledReason ProfileFailureReason `json:"disabled_reason,omitempty"`
}

// AuthProfileStore stores all auth profiles and their state
type AuthProfileStore struct {
	mu       sync.RWMutex
	Version  int                      `json:"version"`
	Profiles map[string]*AuthProfile  `json:"profiles"`
	Stats    map[string]*ProfileUsageStats `json:"stats,omitempty"`
	LastGood map[string]string        `json:"last_good,omitempty"`
	Order    map[string][]string      `json:"order,omitempty"`
}

const (
	AuthStoreVersion = 1
	profileDirName   = "auth-profiles.json"
)

// DefaultProfileCooldowns contains default cooldown durations
var DefaultProfileCooldowns = map[ProfileFailureReason]time.Duration{
	FailureReasonAuth:      5 * time.Minute,
	FailureReasonRateLimit: 1 * time.Minute,
	FailureReasonBilling:   5 * time.Hour,
	FailureReasonTimeout:   30 * time.Second,
	FailureReasonUnknown:   1 * time.Minute,
}

// ProfileCooldownConfig contains configurable cooldown settings
type ProfileCooldownConfig struct {
	BillingBackoffHours int `json:"billing_backoff_hours"`
	BillingMaxHours    int `json:"billing_max_hours"`
	FailureWindowHours int `json:"failure_window_hours"`
}

var defaultCooldownConfig = ProfileCooldownConfig{
	BillingBackoffHours: 5,
	BillingMaxHours:    24,
	FailureWindowHours: 24,
}

// ResolveProfileUnusableUntil returns the time until which a profile is unusable
func (s *ProfileUsageStats) ResolveProfileUnusableUntil() time.Time {
	if s.DisabledUntil.After(time.Now()) {
		return s.DisabledUntil
	}
	if s.CooldownUntil.After(time.Now()) {
		return s.CooldownUntil
	}
	return time.Time{}
}

// IsProfileInCooldown checks if a profile is currently in cooldown
func (s *ProfileUsageStats) IsProfileInCooldown() bool {
	now := time.Now()
	return (!s.CooldownUntil.IsZero() && s.CooldownUntil.After(now)) ||
		(!s.DisabledUntil.IsZero() && s.DisabledUntil.After(now))
}

// NewAuthProfileStore creates a new auth profile store
func NewAuthProfileStore() *AuthProfileStore {
	return &AuthProfileStore{
		Version:  AuthStoreVersion,
		Profiles: make(map[string]*AuthProfile),
		Stats:    make(map[string]*ProfileUsageStats),
		LastGood: make(map[string]string),
		Order:    make(map[string][]string),
	}
}

// ProfileStorePath returns the path to the auth profile store file
func ProfileStorePath() string {
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".swifttalon", profileDirName)
}

// LoadAuthProfileStore loads the auth profile store from disk
func LoadAuthProfileStore() (*AuthProfileStore, error) {
	path := ProfileStorePath()
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return NewAuthProfileStore(), nil
		}
		return nil, fmt.Errorf("reading auth profile store: %w", err)
	}

	var store AuthProfileStore
	if err := json.Unmarshal(data, &store); err != nil {
		return nil, fmt.Errorf("parsing auth profile store: %w", err)
	}

	if store.Profiles == nil {
		store.Profiles = make(map[string]*AuthProfile)
	}
	if store.Stats == nil {
		store.Stats = make(map[string]*ProfileUsageStats)
	}
	if store.LastGood == nil {
		store.LastGood = make(map[string]string)
	}
	if store.Order == nil {
		store.Order = make(map[string][]string)
	}

	return &store, nil
}

// Save saves the auth profile store to disk
func (s *AuthProfileStore) Save() error {
	path := ProfileStorePath()
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("creating auth profile directory: %w", err)
	}

	data, err := json.MarshalIndent(s, "", "  ")
	if err != nil {
		return fmt.Errorf("marshaling auth profile store: %w", err)
	}

	return os.WriteFile(path, data, 0600)
}



// RLock acquires a read lock on the store
func (s *AuthProfileStore) RLock() {
	s.mu.RLock()
}

// RUnlock releases the read lock on the store
func (s *AuthProfileStore) RUnlock() {
	s.mu.RUnlock()
}

// AddProfile adds or updates an auth profile
func (s *AuthProfileStore) AddProfile(profile *AuthProfile) {
	s.mu.Lock()
	defer s.mu.Unlock()

	profile.UpdatedAt = time.Now()
	if profile.CreatedAt.IsZero() {
		profile.CreatedAt = time.Now()
	}
	s.Profiles[profile.ID] = profile
}

// RemoveProfile removes an auth profile by ID
func (s *AuthProfileStore) RemoveProfile(profileID string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	delete(s.Profiles, profileID)
	delete(s.Stats, profileID)

	// Clean up LastGood references
	for provider, lastGood := range s.LastGood {
		if lastGood == profileID {
			delete(s.LastGood, provider)
		}
	}

	// Clean up Order references
	for provider, order := range s.Order {
		newOrder := make([]string, 0)
		for _, id := range order {
			if id != profileID {
				newOrder = append(newOrder, id)
			}
		}
		if len(newOrder) == 0 {
			delete(s.Order, provider)
		} else {
			s.Order[provider] = newOrder
		}
	}
}

// GetProfile returns a profile by ID
func (s *AuthProfileStore) GetProfile(profileID string) (*AuthProfile, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	profile, ok := s.Profiles[profileID]
	return profile, ok
}

// GetProfilesForProvider returns all profiles for a given provider
func (s *AuthProfileStore) GetProfilesForProvider(provider string) []*AuthProfile {
	s.mu.RLock()
	defer s.mu.RUnlock()

	provider = normalizeProviderID(provider)
	profiles := make([]*AuthProfile, 0)

	for _, profile := range s.Profiles {
		if normalizeProviderID(profile.Provider) == provider && !profile.Disabled {
			profiles = append(profiles, profile)
		}
	}

	return profiles
}

// SetProfileOrder sets the explicit order for a provider's profiles
func (s *AuthProfileStore) SetProfileOrder(provider string, order []string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.Order[normalizeProviderID(provider)] = order
}

// GetProfileOrder returns the explicit order for a provider
func (s *AuthProfileStore) GetProfileOrder(provider string) []string {
	s.mu.RLock()
	defer s.mu.RUnlock()

	return s.Order[normalizeProviderID(provider)]
}

// MarkProfileGood marks a profile as working correctly
func (s *AuthProfileStore) MarkProfileGood(provider, profileID string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.LastGood[normalizeProviderID(provider)] = profileID

	// Reset stats for this profile
	if stats, ok := s.Stats[profileID]; ok {
		stats.ErrorCount = 0
		stats.FailureCounts = nil
		stats.CooldownUntil = time.Time{}
		stats.DisabledUntil = time.Time{}
		stats.DisabledReason = ""
	}
}

// MarkProfileUsed marks a profile as recently used
func (s *AuthProfileStore) MarkProfileUsed(profileID string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, ok := s.Stats[profileID]; !ok {
		s.Stats[profileID] = &ProfileUsageStats{}
	}
	s.Stats[profileID].LastUsed = time.Now()
}

// CalculateCooldownDuration calculates the cooldown duration based on error count
func CalculateCooldownDuration(errorCount int, reason ProfileFailureReason) time.Duration {
	baseDuration := DefaultProfileCooldowns[reason]
	if baseDuration == 0 {
		baseDuration = DefaultProfileCooldowns[FailureReasonUnknown]
	}

	// Exponential backoff: 1x, 5x, 25x, max 1 hour
	multiplier := 1.0
	for i := 1; i < errorCount && i <= 3; i++ {
		multiplier *= 5
	}

	duration := baseDuration * time.Duration(multiplier)
	if duration > time.Hour {
		duration = time.Hour
	}

	return duration
}

// CalculateBillingCooldownDuration calculates the billing-specific cooldown duration
func CalculateBillingCooldownDuration(billingErrorCount int) time.Duration {
	baseMs := int64(5 * time.Hour.Milliseconds())
	maxMs := int64(24 * time.Hour.Milliseconds())

	baseMs = maxInt64(60000, baseMs) // At least 1 minute
	maxMs = maxInt64(baseMs, maxMs)

	exponent := minInt(billingErrorCount-1, 10)
	raw := float64(baseMs) * pow(2, float64(exponent))

	cooldownMs := minFloat64(float64(maxMs), raw)
	return time.Duration(cooldownMs)
}

// MarkProfileFailure marks a profile as failed with a specific reason
func (s *AuthProfileStore) MarkProfileFailure(profileID string, reason ProfileFailureReason) {
	s.mu.Lock()
	defer s.mu.Unlock()

	now := time.Now()

	if _, ok := s.Stats[profileID]; !ok {
		s.Stats[profileID] = &ProfileUsageStats{}
	}

	stats := s.Stats[profileID]
	stats.ErrorCount++
	stats.LastFailureAt = now

	if stats.FailureCounts == nil {
		stats.FailureCounts = make(map[string]int)
	}
	stats.FailureCounts[string(reason)]++

	// Apply cooldown based on reason
	if reason == FailureReasonBilling {
		billingCount := stats.FailureCounts[string(FailureReasonBilling)]
		stats.DisabledUntil = now.Add(CalculateBillingCooldownDuration(billingCount))
		stats.DisabledReason = reason
	} else {
		stats.CooldownUntil = now.Add(CalculateCooldownDuration(stats.ErrorCount, reason))
	}
}

// ClearProfileCooldown clears the cooldown for a profile
func (s *AuthProfileStore) ClearProfileCooldown(profileID string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if stats, ok := s.Stats[profileID]; ok {
		stats.CooldownUntil = time.Time{}
		stats.DisabledUntil = time.Time{}
		stats.DisabledReason = ""
		stats.ErrorCount = 0
	}
}

// ResolveAuthProfileOrder resolves the ordered list of profile IDs to use for a provider
// It applies round-robin ordering with cooldown awareness
func (s *AuthProfileStore) ResolveAuthProfileOrder(provider string, explicitOrder []string) []string {
	s.mu.RLock()
	defer s.mu.RUnlock()

	provider = normalizeProviderID(provider)

	// Determine base order: explicit > stored order > available profiles
	var baseOrder []string

	if len(explicitOrder) > 0 {
		baseOrder = explicitOrder
	} else if order, ok := s.Order[provider]; ok && len(order) > 0 {
		baseOrder = order
	} else {
		// Get all profiles for this provider and sort by weight
		for _, p := range s.Profiles {
			if normalizeProviderID(p.Provider) == provider && !p.Disabled {
				baseOrder = append(baseOrder, p.ID)
			}
		}
		// Sort by weight (higher first), then by name
		sort.Slice(baseOrder, func(i, j int) bool {
			pi := s.Profiles[baseOrder[i]]
			pj := s.Profiles[baseOrder[j]]
			if pi.Weight != pj.Weight {
				return pi.Weight > pj.Weight
			}
			return pi.Name < pj.Name
		})
	}

	if len(baseOrder) == 0 {
		return nil
	}

	// Filter out invalid profiles and partition into available/in-cooldown
	var available []string
	var inCooldown []string

	for _, profileID := range baseOrder {
		profile, ok := s.Profiles[profileID]
		if !ok {
			continue
		}
		if normalizeProviderID(profile.Provider) != provider {
			continue
		}
		if profile.Disabled {
			continue
		}
		if profile.Type == CredentialTypeAPIKey && profile.APIKey == "" {
			continue
		}
		if profile.Type == CredentialTypeToken && profile.Token == "" {
			continue
		}

		stats := s.Stats[profileID]
		if stats != nil && stats.IsProfileInCooldown() {
			inCooldown = append(inCooldown, profileID)
		} else {
			available = append(available, profileID)
		}
	}

	// Sort available by lastUsed (oldest first for round-robin)
	sort.Slice(available, func(i, j int) bool {
		statsI := s.Stats[available[i]]
		statsJ := s.Stats[available[j]]
		timeI := time.Time{}
		timeJ := time.Time{}
		if statsI != nil {
			timeI = statsI.LastUsed
		}
		if statsJ != nil {
			timeJ = statsJ.LastUsed
		}
		return timeI.Before(timeJ)
	})

	// Sort in-cooldown by when they become available (soonest first)
	sort.Slice(inCooldown, func(i, j int) bool {
		statsI := s.Stats[inCooldown[i]]
		statsJ := s.Stats[inCooldown[j]]
		unusableI := time.Time{}
		unusableJ := time.Time{}
		if statsI != nil {
			unusableI = statsI.ResolveProfileUnusableUntil()
		}
		if statsJ != nil {
			unusableJ = statsJ.ResolveProfileUnusableUntil()
		}
		return unusableI.Before(unusableJ)
	})

	// Return available first, then in-cooldown
	result := append(available, inCooldown...)

	// If there's a lastGood profile and it's available, prioritize it
	lastGood, ok := s.LastGood[provider]
	if ok {
		for i, pid := range result {
			if pid == lastGood && i > 0 {
				// Move to front
				result = append([]string{pid}, append(result[:i], result[i+1:]...)...)
				break
			}
		}
	}

	return result
}

// normalizeProviderID normalizes a provider ID for comparison
func normalizeProviderID(provider string) string {
	// Convert to lowercase and handle common aliases
	provider = toLower(provider)
	switch provider {
	case "openai", "gpt", "chatgpt":
		return "openai"
	case "anthropic", "claude":
		return "anthropic"
	case "google", "gemini":
		return "gemini"
	case "moonshot", "kimi":
		return "moonshot"
	default:
		return provider
	}
}

func toLower(s string) string {
	result := make([]byte, len(s))
	for i := 0; i < len(s); i++ {
		c := s[i]
		if c >= 'A' && c <= 'Z' {
			c += 'a' - 'A'
		}
		result[i] = c
	}
	return string(result)
}

func minInt(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func maxInt(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func minInt64(a, b int64) int64 {
	if a < b {
		return a
	}
	return b
}

func maxInt64(a, b int64) int64 {
	if a > b {
		return a
	}
	return b
}

func minFloat64(a, b float64) float64 {
	if a < b {
		return a
	}
	return b
}

func pow(base, exp float64) float64 {
	result := 1.0
	for i := 0; i < int(exp); i++ {
		result *= base
	}
	return result
}

// UnlockFunc is returned by Lock() to be called to release the lock
type UnlockFunc func()

// Lock acquires a write lock and returns an unlock function
func (s *AuthProfileStore) Lock() func() {
	s.mu.Lock()
	return s.mu.Unlock
}

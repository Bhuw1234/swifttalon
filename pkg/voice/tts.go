package voice

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/sipeed/picoclaw/pkg/logger"
	"github.com/sipeed/picoclaw/pkg/utils"
)

// TTSProvider defines the interface for Text-to-Speech providers
type TTSProvider interface {
	// Name returns the provider name
	Name() string
	// Convert converts text to speech and returns audio data
	Convert(ctx context.Context, text string, opts *TTSOptions) ([]byte, error)
	// IsAvailable checks if the provider is configured and available
	IsAvailable() bool
	// GetVoices returns available voices for this provider
	GetVoices() []Voice
}

// Voice represents a TTS voice
type Voice struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Language    string `json:"language,omitempty"`
	Gender      string `json:"gender,omitempty"`
	Provider    string `json:"provider"`
}

// TTSOptions holds options for TTS conversion
type TTSOptions struct {
	Voice     string  `json:"voice"`
	Model     string  `json:"model,omitempty"`
	Speed     float64 `json:"speed,omitempty"`
	Language  string  `json:"language,omitempty"`
	OutputFile string  `json:"output_file,omitempty"`
}

// TTSCacheEntry represents a cached TTS audio entry
type TTSCacheEntry struct {
	AudioData   []byte   `json:"-"`
	TextHash    string   `json:"text_hash"`
	Provider    string   `json:"provider"`
	Voice       string   `json:"voice"`
	Model       string   `json:"model"`
	Speed       float64  `json:"speed"`
	Language    string   `json:"language"`
	CreatedAt   int64    `json:"created_at"`
	AudioBase64 string   `json:"audio_base64,omitempty"`
	AudioFile   string   `json:"audio_file,omitempty"`
}

// TTSSystem is the main TTS system that manages providers and caching
type TTSSystem struct {
	mu         sync.RWMutex
	providers  map[string]TTSProvider
	cache      map[string]*TTSCacheEntry
	cacheDir   string
	httpClient *http.Client
	config     *TTSConfig
}

// TTSConfig holds TTS configuration
type TTSConfig struct {
	Enabled      bool              `json:"enabled" env:"PICOCLAW_VOICE_ENABLED"`
	Provider     string            `json:"provider" env:"PICOCLAW_VOICE_PROVIDER"`
	Voice        string            `json:"voice" env:"PICOCLAW_VOICE_VOICE"`
	Model        string            `json:"model" env:"PICOCLAW_VOICE_MODEL"`
	Speed        float64           `json:"speed" env:"PICOCLAW_VOICE_SPEED"`
	CacheEnabled bool              `json:"cache_enabled" env:"PICOCLAW_VOICE_CACHE_ENABLED"`
	CacheDir     string            `json:"cache_dir" env:"PICOCLAW_VOICE_CACHE_DIR"`
	OpenAI       OpenAITTSConfig   `json:"openai,omitempty"`
	ElevenLabs   ElevenLabsConfig  `json:"elevenlabs,omitempty"`
}

// OpenAITTSConfig holds OpenAI TTS specific configuration
type OpenAITTSConfig struct {
	APIKey   string `json:"api_key" env:"PICOCLAW_VOICE_OPENAI_API_KEY"`
	APIBase  string `json:"api_base" env:"PICOCLAW_VOICE_OPENAI_API_BASE"`
	Model    string `json:"model" env:"PICOCLAW_VOICE_OPENAI_MODEL"`
	Voice    string `json:"voice" env:"PICOCLAW_VOICE_OPENAI_VOICE"`
	Speed    string `json:"speed" env:"PICOCLAW_VOICE_OPENAI_SPEED"`
	Response string `json:"response" env:"PICOCLAW_VOICE_OPENAI_RESPONSE"`
}

// ElevenLabsConfig holds ElevenLabs specific configuration
type ElevenLabsConfig struct {
	APIKey       string              `json:"api_key" env:"PICOCLAW_VOICE_ELEVENLABS_API_KEY"`
	BaseURL      string              `json:"base_url" env:"PICOCLAW_VOICE_ELEVENLABS_BASE_URL"`
	VoiceID      string              `json:"voice_id" env:"PICOCLAW_VOICE_ELEVENLABS_VOICE_ID"`
	ModelID      string              `json:"model_id" env:"PICOCLAW_VOICE_ELEVENLABS_MODEL_ID"`
	LanguageCode string              `json:"language_code" env:"PICOCLAW_VOICE_ELEVENLABS_LANGUAGE_CODE"`
	Seed         int                 `json:"seed" env:"PICOCLAW_VOICE_ELEVENLABS_SEED"`
	VoiceSettings *ElevenLabsVoiceSettings `json:"voice_settings,omitempty"`
}

// ElevenLabsVoiceSettings holds ElevenLabs voice settings
type ElevenLabsVoiceSettings struct {
	Stability            float64 `json:"stability"`
	SimilarityBoost      float64 `json:"similarity_boost"`
	Style                float64 `json:"style"`
	UseSpeakerBoost      bool    `json:"use_speaker_boost"`
	Speed                float64 `json:"speed"`
}

// TTSResult holds the result of a TTS conversion
type TTSResult struct {
	AudioData   []byte `json:"audio_data,omitempty"`
	AudioBase64 string `json:"audio_base64,omitempty"`
	AudioFile   string `json:"audio_file,omitempty"`
	Provider    string `json:"provider"`
	Voice       string `json:"voice"`
	Model       string `json:"model"`
	Duration    int    `json:"duration,omitempty"` // Estimated duration in seconds
	Cached      bool   `json:"cached"`
}

// DefaultTTSConfig returns the default TTS configuration
func DefaultTTSConfig() *TTSConfig {
	return &TTSConfig{
		Enabled:      false,
		Provider:     "openai",
		Voice:        "alloy",
		Model:        "tts-1",
		Speed:        1.0,
		CacheEnabled: true,
		CacheDir:     "~/.picoclaw/cache/tts",
		OpenAI: OpenAITTSConfig{
			APIKey:   "",
			APIBase:  "https://api.openai.com/v1",
			Model:    "tts-1",
			Voice:    "alloy",
			Speed:    "1.0",
			Response: "mp3",
		},
		ElevenLabs: ElevenLabsConfig{
			APIKey:       "",
			BaseURL:      "https://api.elevenlabs.io/v1",
			VoiceID:      "21m00Tcm4TlvDq8ikWAM", // Rachel voice
			ModelID:      "eleven_multilingual_v2",
			LanguageCode: "en",
			Seed:         0,
			VoiceSettings: &ElevenLabsVoiceSettings{
				Stability:            0.5,
				SimilarityBoost:      0.75,
				Style:                0.0,
				UseSpeakerBoost:      true,
				Speed:                1.0,
			},
		},
	}
}

// AvailableOpenAIVoices returns the list of available OpenAI voices
var AvailableOpenAIVoices = []Voice{
	{ID: "alloy", Name: "Alloy", Language: "en", Gender: "neutral", Provider: "openai"},
	{ID: "echo", Name: "Echo", Language: "en", Gender: "male", Provider: "openai"},
	{ID: "fable", Name: "Fable", Language: "en", Gender: "male", Provider: "openai"},
	{ID: "onyx", Name: "Onyx", Language: "en", Gender: "male", Provider: "openai"},
	{ID: "nova", Name: "Nova", Language: "en", Gender: "female", Provider: "openai"},
	{ID: "shimmer", Name: "Shimmer", Language: "en", Gender: "female", Provider: "openai"},
	{ID: "ash", Name: "Ash", Language: "en", Gender: "neutral", Provider: "openai"},
	{ID: "ballad", Name: "Ballad", Language: "en", Gender: "neutral", Provider: "openai"},
	{ID: "coral", Name: "Coral", Language: "en", Gender: "female", Provider: "openai"},
	{ID: "sage", Name: "Sage", Language: "en", Gender: "neutral", Provider: "openai"},
}

// OpenAITTSProvider implements TTSProvider for OpenAI
type OpenAITTSProvider struct {
	apiKey     string
	apiBase    string
	model      string
	voice      string
	speed      string
	response   string
	httpClient *http.Client
}

// NewOpenAITTSProvider creates a new OpenAI TTS provider
func NewOpenAITTSProvider(cfg OpenAITTSConfig) *OpenAITTSProvider {
	logger.DebugCF("voice.tts", "Creating OpenAI TTS provider", map[string]interface{}{
		"has_api_key": cfg.APIKey != "",
		"model":       cfg.Model,
		"voice":       cfg.Voice,
	})

	apiBase := cfg.APIBase
	if apiBase == "" {
		apiBase = "https://api.openai.com/v1"
	}

	speed := cfg.Speed
	if speed == "" {
		speed = "1.0"
	}

	response := cfg.Response
	if response == "" {
		response = "mp3"
	}

	return &OpenAITTSProvider{
		apiKey:  cfg.APIKey,
		apiBase: apiBase,
		model:   cfg.Model,
		voice:   cfg.Voice,
		speed:   speed,
		response: response,
		httpClient: &http.Client{
			Timeout: 60 * time.Second,
		},
	}
}

// Name returns the provider name
func (p *OpenAITTSProvider) Name() string {
	return "openai"
}

// Convert converts text to speech using OpenAI
func (p *OpenAITTSProvider) Convert(ctx context.Context, text string, opts *TTSOptions) ([]byte, error) {
	logger.InfoCF("voice.tts", "Converting text to speech with OpenAI", map[string]interface{}{
		"text_length":    len(text),
		"voice":          opts.Voice,
		"model":          opts.Model,
	})

	voice := opts.Voice
	if voice == "" {
		voice = p.voice
	}

	model := opts.Model
	if model == "" {
		model = p.model
	}

	speed := p.speed
	if opts.Speed != 0 && opts.Speed != 1.0 {
		speed = fmt.Sprintf("%.1f", opts.Speed)
	}

	url := p.apiBase + "/audio/speech"

	reqBody := map[string]interface{}{
		"model":   model,
		"voice":   voice,
		"input":   text,
		"speed":   speed,
		"response_format": p.response,
	}

	jsonBody, err := json.Marshal(reqBody)
	if err != nil {
		logger.ErrorCF("voice.tts", "Failed to marshal request body", map[string]interface{}{"error": err})
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(jsonBody))
	if err != nil {
		logger.ErrorCF("voice.tts", "Failed to create request", map[string]interface{}{"error": err})
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+p.apiKey)

	logger.DebugCF("voice.tts", "Sending request to OpenAI TTS API", map[string]interface{}{
		"url":     url,
		"voice":   voice,
		"model":   model,
		"text_len": len(text),
	})

	resp, err := p.httpClient.Do(req)
	if err != nil {
		logger.ErrorCF("voice.tts", "Failed to send request", map[string]interface{}{"error": err})
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		logger.ErrorCF("voice.tts", "Failed to read response", map[string]interface{}{"error": err})
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		logger.ErrorCF("voice.tts", "OpenAI API error", map[string]interface{}{
			"status_code": resp.StatusCode,
			"response":    string(body),
		})
		return nil, fmt.Errorf("API error (status %d): %s", resp.StatusCode, string(body))
	}

	logger.InfoCF("voice.tts", "OpenAI TTS conversion completed", map[string]interface{}{
		"audio_size": len(body),
		"provider":   "openai",
	})

	return body, nil
}

// IsAvailable checks if OpenAI TTS is available
func (p *OpenAITTSProvider) IsAvailable() bool {
	available := p.apiKey != ""
	logger.DebugCF("voice.tts", "Checking OpenAI TTS availability", map[string]interface{}{"available": available})
	return available
}

// GetVoices returns available OpenAI voices
func (p *OpenAITTSProvider) GetVoices() []Voice {
	return AvailableOpenAIVoices
}

// ElevenLabsProvider implements TTSProvider for ElevenLabs
type ElevenLabsProvider struct {
	apiKey       string
	baseURL      string
	voiceID      string
	modelID      string
	languageCode string
	seed         int
	settings     *ElevenLabsVoiceSettings
	httpClient   *http.Client
}

// NewElevenLabsProvider creates a new ElevenLabs TTS provider
func NewElevenLabsProvider(cfg ElevenLabsConfig) *ElevenLabsProvider {
	logger.DebugCF("voice.tts", "Creating ElevenLabs TTS provider", map[string]interface{}{
		"has_api_key":  cfg.APIKey != "",
		"voice_id":     cfg.VoiceID,
		"model_id":     cfg.ModelID,
		"language_code": cfg.LanguageCode,
	})

	baseURL := cfg.BaseURL
	if baseURL == "" {
		baseURL = "https://api.elevenlabs.io/v1"
	}

	settings := cfg.VoiceSettings
	if settings == nil {
		settings = &ElevenLabsVoiceSettings{
			Stability:        0.5,
			SimilarityBoost:   0.75,
			Style:             0.0,
			UseSpeakerBoost:   true,
			Speed:             1.0,
		}
	}

	return &ElevenLabsProvider{
		apiKey:       cfg.APIKey,
		baseURL:      baseURL,
		voiceID:      cfg.VoiceID,
		modelID:      cfg.ModelID,
		languageCode: cfg.LanguageCode,
		seed:         cfg.Seed,
		settings:     settings,
		httpClient: &http.Client{
			Timeout: 60 * time.Second,
		},
	}
}

// Name returns the provider name
func (p *ElevenLabsProvider) Name() string {
	return "elevenlabs"
}

// Convert converts text to speech using ElevenLabs
func (p *ElevenLabsProvider) Convert(ctx context.Context, text string, opts *TTSOptions) ([]byte, error) {
	logger.InfoCF("voice.tts", "Converting text to speech with ElevenLabs", map[string]interface{}{
		"text_length": len(text),
		"voice_id":    opts.Voice,
		"model_id":    opts.Model,
	})

	voiceID := opts.Voice
	if voiceID == "" {
		voiceID = p.voiceID
	}

	modelID := opts.Model
	if modelID == "" {
		modelID = p.modelID
	}

	language := opts.Language
	if language == "" {
		language = p.languageCode
	}

	speed := p.settings.Speed
	if opts.Speed != 0 && opts.Speed != 1.0 {
		speed = opts.Speed
	}

	url := fmt.Sprintf("%s/text-to-speech/%s", p.baseURL, voiceID)

	reqBody := map[string]interface{}{
		"text":              text,
		"model_id":          modelID,
		"language_code":     language,
		"voice_settings": map[string]interface{}{
			"stability":            p.settings.Stability,
			"similarity_boost":     p.settings.SimilarityBoost,
			"style":                p.settings.Style,
			"use_speaker_boost":    p.settings.UseSpeakerBoost,
			"speed":                speed,
		},
	}

	// Add seed if specified
	if p.seed > 0 {
		reqBody["seed"] = p.seed
	}

	jsonBody, err := json.Marshal(reqBody)
	if err != nil {
		logger.ErrorCF("voice.tts", "Failed to marshal request body", map[string]interface{}{"error": err})
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(jsonBody))
	if err != nil {
		logger.ErrorCF("voice.tts", "Failed to create request", map[string]interface{}{"error": err})
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "audio/mpeg")
	req.Header.Set("xi-api-key", p.apiKey)

	logger.DebugCF("voice.tts", "Sending request to ElevenLabs TTS API", map[string]interface{}{
		"url":       url,
		"voice_id":  voiceID,
		"model_id":  modelID,
		"text_len":  len(text),
		"language":  language,
	})

	resp, err := p.httpClient.Do(req)
	if err != nil {
		logger.ErrorCF("voice.tts", "Failed to send request", map[string]interface{}{"error": err})
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		logger.ErrorCF("voice.tts", "Failed to read response", map[string]interface{}{"error": err})
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		logger.ErrorCF("voice.tts", "ElevenLabs API error", map[string]interface{}{
			"status_code": resp.StatusCode,
			"response":    string(body),
		})
		return nil, fmt.Errorf("API error (status %d): %s", resp.StatusCode, string(body))
	}

	logger.InfoCF("voice.tts", "ElevenLabs TTS conversion completed", map[string]interface{}{
		"audio_size": len(body),
		"provider":   "elevenlabs",
	})

	return body, nil
}

// IsAvailable checks if ElevenLabs TTS is available
func (p *ElevenLabsProvider) IsAvailable() bool {
	available := p.apiKey != ""
	logger.DebugCF("voice.tts", "Checking ElevenLabs TTS availability", map[string]interface{}{"available": available})
	return available
}

// GetVoices returns available ElevenLabs voices (placeholder - would need API call)
func (p *ElevenLabsProvider) GetVoices() []Voice {
	// In production, this would fetch from the ElevenLabs API
	// For now, return the configured voice
	return []Voice{
		{ID: p.voiceID, Name: "ElevenLabs Voice", Language: p.languageCode, Provider: "elevenlabs"},
	}
}

// NewTTSSystem creates a new TTS system
func NewTTSSystem(cfg *TTSConfig) *TTSSystem {
	logger.InfoCF("voice.tts", "Initializing TTS system", map[string]interface{}{
		"enabled":    cfg.Enabled,
		"provider":  cfg.Provider,
		"cache_dir":  cfg.CacheDir,
	})

	system := &TTSSystem{
		providers: make(map[string]TTSProvider),
		cache:     make(map[string]*TTSCacheEntry),
		cacheDir:  expandHome(cfg.CacheDir),
		config:    cfg,
		httpClient: &http.Client{
			Timeout: 60 * time.Second,
		},
	}

	// Register OpenAI provider
	openaiProvider := NewOpenAITTSProvider(cfg.OpenAI)
	system.providers["openai"] = openaiProvider

	// Register ElevenLabs provider
	elevenlabsProvider := NewElevenLabsProvider(cfg.ElevenLabs)
	system.providers["elevenlabs"] = elevenlabsProvider

	// Load cache if enabled
	if cfg.CacheEnabled {
		system.loadCache()
	}

	return system
}

// expandHome expands ~ to home directory
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

// Convert converts text to speech using the configured provider
func (s *TTSSystem) Convert(ctx context.Context, text string, opts *TTSOptions) (*TTSResult, error) {
	if !s.config.Enabled {
		logger.WarnCF("voice.tts", "TTS is not enabled", nil)
		return nil, fmt.Errorf("TTS is not enabled")
	}

	// Determine provider
	providerName := s.config.Provider
	if opts != nil && opts.Voice != "" {
		// Try to determine provider from voice prefix (e.g., "elevenlabs:voice_id")
		if strings.Contains(opts.Voice, ":") {
			providerName = strings.Split(opts.Voice, ":")[0]
		}
	}

	provider, ok := s.providers[providerName]
	if !ok {
		logger.ErrorCF("voice.tts", "Unknown TTS provider", map[string]interface{}{"provider": providerName})
		return nil, fmt.Errorf("unknown TTS provider: %s", providerName)
	}

	if !provider.IsAvailable() {
		logger.ErrorCF("voice.tts", "TTS provider not available", map[string]interface{}{"provider": providerName})
		return nil, fmt.Errorf("TTS provider not available: %s", providerName)
	}

	// Apply default options
	if opts == nil {
		opts = &TTSOptions{}
	}
	if opts.Voice == "" {
		opts.Voice = s.config.Voice
	}
	if opts.Model == "" {
		opts.Model = s.config.Model
	}
	if opts.Speed == 0 {
		opts.Speed = s.config.Speed
	}

	// Check cache
	cacheKey := s.generateCacheKey(text, providerName, opts)
	if s.config.CacheEnabled {
		if cached := s.getFromCache(cacheKey); cached != nil {
			logger.InfoCF("voice.tts", "Using cached TTS audio", map[string]interface{}{
				"cache_key": cacheKey,
				"provider":  providerName,
			})
			return &TTSResult{
				AudioData:   cached.AudioData,
				AudioBase64: cached.AudioBase64,
				AudioFile:   cached.AudioFile,
				Provider:    cached.Provider,
				Voice:       cached.Voice,
				Model:       cached.Model,
				Cached:      true,
			}, nil
		}
	}

	// Convert text to speech
	audioData, err := provider.Convert(ctx, text, opts)
	if err != nil {
		logger.ErrorCF("voice.tts", "TTS conversion failed", map[string]interface{}{
			"provider": providerName,
			"error":    err,
		})
		return nil, fmt.Errorf("TTS conversion failed: %w", err)
	}

	// Save to file if specified
	audioFile := ""
	if opts.OutputFile != "" {
		audioFile = s.saveAudioFile(opts.OutputFile, audioData)
	}

	// Cache the result
	if s.config.CacheEnabled {
		s.saveToCache(cacheKey, audioData, providerName, opts, audioFile)
	}

	// Encode to base64
	audioBase64 := base64.StdEncoding.EncodeToString(audioData)

	return &TTSResult{
		AudioData:   audioData,
		AudioBase64: audioBase64,
		AudioFile:   audioFile,
		Provider:    providerName,
		Voice:       opts.Voice,
		Model:       opts.Model,
		Cached:      false,
	}, nil
}

// ConvertToFile converts text to speech and saves to a file
func (s *TTSSystem) ConvertToFile(ctx context.Context, text string, outputPath string, opts *TTSOptions) (string, error) {
	if opts == nil {
		opts = &TTSOptions{}
	}
	opts.OutputFile = outputPath

	result, err := s.Convert(ctx, text, opts)
	if err != nil {
		return "", err
	}

	return result.AudioFile, nil
}

// GetProvider returns the provider by name
func (s *TTSSystem) GetProvider(name string) (TTSProvider, bool) {
	provider, ok := s.providers[name]
	return provider, ok
}

// GetAvailableProviders returns available and configured providers
func (s *TTSSystem) GetAvailableProviders() []string {
	var available []string
	for name, provider := range s.providers {
		if provider.IsAvailable() {
			available = append(available, name)
		}
	}
	return available
}

// GetVoices returns available voices for a provider
func (s *TTSSystem) GetVoices(providerName string) []Voice {
	provider, ok := s.providers[providerName]
	if !ok {
		return nil
	}
	return provider.GetVoices()
}

// IsEnabled returns whether TTS is enabled
func (s *TTSSystem) IsEnabled() bool {
	return s.config.Enabled
}

// GetConfig returns the TTS configuration
func (s *TTSSystem) GetConfig() *TTSConfig {
	return s.config
}

// generateCacheKey generates a cache key for the TTS request
func (s *TTSSystem) generateCacheKey(text string, provider string, opts *TTSOptions) string {
	data := fmt.Sprintf("%s|%s|%s|%s|%f|%s",
		text,
		provider,
		opts.Voice,
		opts.Model,
		opts.Speed,
		opts.Language,
	)
	hash := sha256.Sum256([]byte(data))
	return fmt.Sprintf("%x", hash)
}

// getFromCache retrieves cached audio data
func (s *TTSSystem) getFromCache(key string) *TTSCacheEntry {
	s.mu.RLock()
	defer s.mu.RUnlock()

	entry, ok := s.cache[key]
	if !ok {
		return nil
	}

	// Check if cache entry is expired (7 days)
	age := time.Now().Unix() - entry.CreatedAt
	if age > 7*24*60*60 {
		logger.DebugCF("voice.tts", "Cache entry expired", map[string]interface{}{"key": key, "age_days": age/86400})
		return nil
	}

	// Decode base64 if needed
	if entry.AudioData == nil && entry.AudioBase64 != "" {
		data, err := base64.StdEncoding.DecodeString(entry.AudioBase64)
		if err != nil {
			logger.ErrorCF("voice.tts", "Failed to decode cached audio", map[string]interface{}{"error": err})
			return nil
		}
		entry.AudioData = data
	}

	// Load from file if needed
	if entry.AudioData == nil && entry.AudioFile != "" {
		data, err := os.ReadFile(entry.AudioFile)
		if err != nil {
			logger.ErrorCF("voice.tts", "Failed to read cached audio file", map[string]interface{}{"error": err, "file": entry.AudioFile})
			return nil
		}
		entry.AudioData = data
	}

	return entry
}

// saveToCache saves audio data to cache
func (s *TTSSystem) saveToCache(key string, audioData []byte, provider string, opts *TTSOptions, audioFile string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	entry := &TTSCacheEntry{
		TextHash:  key,
		Provider:  provider,
		Voice:     opts.Voice,
		Model:     opts.Model,
		Speed:     opts.Speed,
		Language:  opts.Language,
		CreatedAt: time.Now().Unix(),
		AudioData: audioData,
		AudioFile: audioFile,
	}

	// Store base64 in memory (audio data may be large)
	entry.AudioBase64 = base64.StdEncoding.EncodeToString(audioData)
	entry.AudioData = nil // Clear raw data to save memory

	s.cache[key] = entry

	logger.DebugCF("voice.tts", "Cached TTS audio", map[string]interface{}{
		"key":         key,
		"provider":    provider,
		"size_bytes":  len(audioData),
	})
}

// saveAudioFile saves audio data to a file
func (s *TTSSystem) saveAudioFile(outputPath string, audioData []byte) string {
	// Expand home directory
	outputPath = expandHome(outputPath)

	// Create directory if it doesn't exist
	dir := filepath.Dir(outputPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		logger.ErrorCF("voice.tts", "Failed to create audio directory", map[string]interface{}{
			"dir":   dir,
			"error": err,
		})
		return ""
	}

	// Write file
	if err := os.WriteFile(outputPath, audioData, 0644); err != nil {
		logger.ErrorCF("voice.tts", "Failed to write audio file", map[string]interface{}{
			"path":  outputPath,
			"error": err,
		})
		return ""
	}

	logger.InfoCF("voice.tts", "Saved audio file", map[string]interface{}{
		"path": outputPath,
		"size": len(audioData),
	})

	return outputPath
}

// loadCache loads cache from disk
func (s *TTSSystem) loadCache() {
	// Create cache directory if it doesn't exist
	if err := os.MkdirAll(s.cacheDir, 0755); err != nil {
		logger.ErrorCF("voice.tts", "Failed to create cache directory", map[string]interface{}{
			"dir":   s.cacheDir,
			"error": err,
		})
		return
	}

	// Load cache index file (stores metadata about cached files)
	cacheIndexPath := filepath.Join(s.cacheDir, "index.json")
	data, err := os.ReadFile(cacheIndexPath)
	if err != nil {
		if os.IsNotExist(err) {
			logger.DebugCF("voice.tts", "No cache index found, starting fresh", nil)
			return
		}
		logger.ErrorCF("voice.tts", "Failed to read cache index", map[string]interface{}{"error": err})
		return
	}

	var entries map[string]*TTSCacheEntry
	if err := json.Unmarshal(data, &entries); err != nil {
		logger.ErrorCF("voice.tts", "Failed to parse cache index", map[string]interface{}{"error": err})
		return
	}

	s.cache = entries
	logger.InfoCF("voice.tts", "Loaded TTS cache", map[string]interface{}{"entries": len(entries)})
}

// saveCache saves cache index to disk
func (s *TTSSystem) saveCache() {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Save only metadata (not audio data) to index file
	entries := make(map[string]*TTSCacheEntry)
	for k, v := range s.cache {
		entryCopy := *v
		entryCopy.AudioData = nil // Don't save raw data
		entries[k] = &entryCopy
	}

	data, err := json.MarshalIndent(entries, "", "  ")
	if err != nil {
		logger.ErrorCF("voice.tts", "Failed to marshal cache index", map[string]interface{}{"error": err})
		return
	}

	cacheIndexPath := filepath.Join(s.cacheDir, "index.json")
	if err := os.WriteFile(cacheIndexPath, data, 0644); err != nil {
		logger.ErrorCF("voice.tts", "Failed to write cache index", map[string]interface{}{"error": err})
		return
	}

	logger.DebugCF("voice.tts", "Saved TTS cache index", map[string]interface{}{"entries": len(entries)})
}

// ClearCache clears the TTS cache
func (s *TTSSystem) ClearCache() {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.cache = make(map[string]*TTSCacheEntry)

	// Remove cache index file
	cacheIndexPath := filepath.Join(s.cacheDir, "index.json")
	os.Remove(cacheIndexPath)

	logger.InfoCF("voice.tts", "Cleared TTS cache", nil)
}

// GetCacheStats returns cache statistics
func (s *TTSSystem) GetCacheStats() map[string]interface{} {
	s.mu.RLock()
	defer s.mu.RUnlock()

	return map[string]interface{}{
		"entries":      len(s.cache),
		"cache_dir":    s.cacheDir,
		"cache_enabled": s.config.CacheEnabled,
	}
}

// TruncateText truncates text to fit TTS limits
func TruncateText(text string, maxLength int) string {
	if len(text) <= maxLength {
		return text
	}
	return utils.Truncate(text, maxLength) + "..."
}

// SwiftTalon - Ultra-lightweight personal AI agent
// License: MIT
//
// Copyright (c) 2026 SwiftTalon contributors

package providers

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/Bhuw1234/swifttalon/pkg/auth"
	"github.com/Bhuw1234/swifttalon/pkg/config"
)

// CopilotProvider implements LLMProvider for GitHub Copilot API
type CopilotProvider struct {
	apiKey     string
	apiBase    string
	httpClient *http.Client
}

// NewCopilotProvider creates a new GitHub Copilot provider
func NewCopilotProvider(apiKey, apiBase, proxy string) *CopilotProvider {
	client := &http.Client{
		Timeout: 120 * time.Second,
	}

	if proxy != "" {
		proxyURL, err := url.Parse(proxy)
		if err == nil {
			client.Transport = &http.Transport{
				Proxy: http.ProxyURL(proxyURL),
			}
		}
	}

	// Default API base for GitHub Copilot
	if apiBase == "" {
		apiBase = "https://api.github.com/integrations/github-copilot/v1"
	}

	return &CopilotProvider{
		apiKey:     apiKey,
		apiBase:    strings.TrimRight(apiBase, "/"),
		httpClient: client,
	}
}

// Chat sends a chat request to GitHub Copilot
func (p *CopilotProvider) Chat(ctx context.Context, messages []Message, tools []ToolDefinition, model string, options map[string]interface{}) (*LLMResponse, error) {
	if p.apiBase == "" {
		return nil, fmt.Errorf("API base not configured")
	}

	// Use default model if not specified
	if model == "" {
		model = p.GetDefaultModel()
	}

	// Build request body for GitHub Copilot API
	requestBody := map[string]interface{}{
		"model":    model,
		"messages": messages,
		"stream":   false,
	}

	// Add tools if provided
	if len(tools) > 0 {
		requestBody["tools"] = tools
		requestBody["tool_choice"] = "auto"
	}

	// Handle optional parameters
	if maxTokens, ok := options["max_tokens"].(int); ok {
		requestBody["max_tokens"] = maxTokens
	}

	if temperature, ok := options["temperature"].(float64); ok {
		requestBody["temperature"] = temperature
	}

	if topP, ok := options["top_p"].(float64); ok {
		requestBody["top_p"] = topP
	}

	jsonData, err := json.Marshal(requestBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", p.apiBase+"/chat/completions", bytes.NewReader(jsonData))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Set headers for GitHub Copilot API
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/vnd.github+json")
	req.Header.Set("X-GitHub-Api-Version", "2022-11-28")

	// Set authorization header
	if p.apiKey != "" {
		req.Header.Set("Authorization", "Bearer "+p.apiKey)
	}

	resp, err := p.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		return nil, fmt.Errorf("API request failed:\n  Status: %d\n  Body:   %s", resp.StatusCode, string(body))
	}

	return p.parseResponse(body)
}

// parseResponse parses the GitHub Copilot API response
func (p *CopilotProvider) parseResponse(body []byte) (*LLMResponse, error) {
	var apiResponse struct {
		Choices []struct {
			Message struct {
				Content   string `json:"content"`
				Role      string `json:"role"`
				ToolCalls []struct {
					ID       string `json:"id"`
					Type     string `json:"type"`
					Function *struct {
						Name      string `json:"name"`
						Arguments string `json:"arguments"`
					} `json:"function"`
				} `json:"tool_calls"`
			} `json:"message"`
			FinishReason string `json:"finish_reason"`
		} `json:"choices"`
		Usage *UsageInfo `json:"usage"`
	}

	if err := json.Unmarshal(body, &apiResponse); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	if len(apiResponse.Choices) == 0 {
		return &LLMResponse{
			Content:      "",
			FinishReason: "stop",
		}, nil
	}

	choice := apiResponse.Choices[0]

	toolCalls := make([]ToolCall, 0, len(choice.Message.ToolCalls))
	for _, tc := range choice.Message.ToolCalls {
		arguments := make(map[string]interface{})
		name := ""

		if tc.Type == "function" && tc.Function != nil {
			name = tc.Function.Name
			if tc.Function.Arguments != "" {
				if err := json.Unmarshal([]byte(tc.Function.Arguments), &arguments); err != nil {
					arguments["raw"] = tc.Function.Arguments
				}
			}
		} else if tc.Function != nil {
			// Legacy format without type field
			name = tc.Function.Name
			if tc.Function.Arguments != "" {
				if err := json.Unmarshal([]byte(tc.Function.Arguments), &arguments); err != nil {
					arguments["raw"] = tc.Function.Arguments
				}
			}
		}

		toolCalls = append(toolCalls, ToolCall{
			ID:        tc.ID,
			Name:      name,
			Arguments: arguments,
		})
	}

	return &LLMResponse{
		Content:      choice.Message.Content,
		ToolCalls:    toolCalls,
		FinishReason: choice.FinishReason,
		Usage:        apiResponse.Usage,
	}, nil
}

// GetDefaultModel returns the default model for GitHub Copilot
func (p *CopilotProvider) GetDefaultModel() string {
	return "gpt-4.1"
}

// createCopilotAuthProvider creates a Copilot provider using auth credentials
func createCopilotAuthProvider(cfg *config.Config) (LLMProvider, error) {
	// Try to get GitHub token from auth store
	cred, err := getGitHubCopilotCredential()
	if err != nil {
		return nil, fmt.Errorf("loading auth credentials: %w", err)
	}

	apiBase := cfg.Providers.GitHubCopilot.APIBase
	if apiBase == "" {
		apiBase = "https://api.github.com/integrations/github-copilot/v1"
	}

	proxy := cfg.Providers.GitHubCopilot.Proxy

	return NewCopilotProvider(cred.AccessToken, apiBase, proxy), nil
}

// getGitHubCopilotCredential attempts to get credentials from auth store
func getGitHubCopilotCredential() (*auth.AuthCredential, error) {
	// First try github_copilot provider
	if cred, err := auth.GetCredential("github_copilot"); err == nil && cred != nil {
		return cred, nil
	}
	// Fall back to github provider
	if cred, err := auth.GetCredential("github"); err == nil && cred != nil {
		return cred, nil
	}
	return nil, fmt.Errorf("no credentials for github-copilot. Run: swifttalon auth login --provider github-copilot")
}

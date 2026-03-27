package providers

import (
	"context"
	"fmt"

	json "encoding/json"

	copilot "github.com/github/copilot-sdk/go"
)

// approveAllPermissions is a permission handler that approves all permission requests.
// This is required by the SDK but we implement it locally for v0.1.23 compatibility.
func approveAllPermissions(_ copilot.PermissionRequest, _ copilot.PermissionInvocation) (copilot.PermissionRequestResult, error) {
	return copilot.PermissionRequestResult{Kind: "approved"}, nil
}

type GitHubCopilotProvider struct {
	uri         string
	connectMode string // `stdio` or `grpc`

	client  *copilot.Client
	session *copilot.Session
}

func NewGitHubCopilotProvider(uri string, connectMode string, model string) (*GitHubCopilotProvider, error) {
	if connectMode == "" {
		connectMode = "stdio" // Default to stdio mode (recommended)
	}

	var client *copilot.Client
	var session *copilot.Session
	var err error

	switch connectMode {
	case "stdio":
		// stdio mode: spawn a CLI process and communicate via stdin/stdout
		// This is the recommended mode for most use cases.
		// The SDK will spawn a `copilot` process (or use COPILOT_CLI_PATH env var).
		opts := &copilot.ClientOptions{
			LogLevel: "error", // Reduce noise in logs
		}
		// If uri is provided, treat it as CLIPath (path to copilot executable)
		if uri != "" {
			opts.CLIPath = uri
		}
		client = copilot.NewClient(opts)

		if err = client.Start(context.Background()); err != nil {
			return nil, fmt.Errorf("failed to start GitHub Copilot CLI in stdio mode: %w", err)
		}

		session, err = client.CreateSession(context.Background(), &copilot.SessionConfig{
			Model:               model,
			OnPermissionRequest: approveAllPermissions, // Required by SDK
			Hooks:               &copilot.SessionHooks{},
		})
		if err != nil {
			client.Stop()
			return nil, fmt.Errorf("failed to create GitHub Copilot session: %w", err)
		}

	case "grpc":
		// grpc mode: connect to an existing CLI server via gRPC.
		// The uri should be the server address (e.g., "localhost:8080").
		// The client will NOT spawn a CLI process in this mode.
		if uri == "" {
			return nil, fmt.Errorf("uri is required for grpc mode (e.g., 'localhost:8080')")
		}
		client = copilot.NewClient(&copilot.ClientOptions{
			CLIUrl: uri,
		})

		if err = client.Start(context.Background()); err != nil {
			return nil, fmt.Errorf("failed to connect to GitHub Copilot server at %s: %w (see https://github.com/github/copilot-sdk/blob/main/docs/getting-started.md#connecting-to-an-external-cli-server)", uri, err)
		}

		session, err = client.CreateSession(context.Background(), &copilot.SessionConfig{
			Model:               model,
			OnPermissionRequest: approveAllPermissions, // Required by SDK
			Hooks:               &copilot.SessionHooks{},
		})
		if err != nil {
			client.Stop()
			return nil, fmt.Errorf("failed to create GitHub Copilot session: %w", err)
		}

	default:
		return nil, fmt.Errorf("unsupported connect mode: %s (supported: 'stdio', 'grpc')", connectMode)
	}

	return &GitHubCopilotProvider{
		uri:         uri,
		connectMode: connectMode,
		client:      client,
		session:     session,
	}, nil
}

// Close stops the GitHub Copilot client and releases resources.
// It should be called when the provider is no longer needed.
func (p *GitHubCopilotProvider) Close() {
	if p.session != nil {
		// Destroy releases session resources (v0.1.23 uses Destroy, newer versions use Disconnect)
		p.session.Destroy()
	}
	if p.client != nil {
		p.client.Stop()
	}
}

// Chat sends a chat request to GitHub Copilot
func (p *GitHubCopilotProvider) Chat(ctx context.Context, messages []Message, tools []ToolDefinition, model string, options map[string]interface{}) (*LLMResponse, error) {
	type tempMessage struct {
		Role    string `json:"role"`
		Content string `json:"content"`
	}
	out := make([]tempMessage, 0, len(messages))

	for _, msg := range messages {
		out = append(out, tempMessage{
			Role:    msg.Role,
			Content: msg.Content,
		})
	}

	fullcontent, _ := json.Marshal(out)

	content, _ := p.session.Send(ctx, copilot.MessageOptions{
		Prompt: string(fullcontent),
	})

	return &LLMResponse{
		FinishReason: "stop",
		Content:      content,
	}, nil

}

func (p *GitHubCopilotProvider) GetDefaultModel() string {

	return "gpt-4.1"
}

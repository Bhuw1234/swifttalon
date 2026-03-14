package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"html"
	"io"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"time"

	"github.com/Bhuw1234/swifttalon/pkg/providers"
)

const (
	// Default user agent to mimic browser requests
	defaultUserAgent = "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36"

	// Default values
	defaultMaxLength    = 20000
	defaultSummaryLength = 1000

	// Timeout settings
	defaultFetchTimeout    = 30 * time.Second
	defaultSummaryTimeout  = 60 * time.Second
	maxRedirects           = 5
)

// LinkTool fetches URLs and extracts content with optional AI summarization.
type LinkTool struct {
	httpClient  *http.Client
	llmProvider providers.LLMProvider
	maxLength   int
}

// LinkToolOptions configures the LinkTool.
type LinkToolOptions struct {
	// MaxLength is the maximum number of characters to extract from a page.
	// Defaults to 20000.
	MaxLength int

	// HTTPClient allows custom HTTP client configuration.
	// If nil, a default client with sensible timeouts will be used.
	HTTPClient *http.Client
}

// NewLinkTool creates a new LinkTool instance.
// The llmProvider is optional - if provided, AI summarization will be available.
func NewLinkTool(llmProvider providers.LLMProvider, opts LinkToolOptions) *LinkTool {
	maxLength := opts.MaxLength
	if maxLength <= 0 {
		maxLength = defaultMaxLength
	}

	client := opts.HTTPClient
	if client == nil {
		client = &http.Client{
			Timeout: defaultFetchTimeout,
			CheckRedirect: func(req *http.Request, via []*http.Request) error {
				if len(via) >= maxRedirects {
					return fmt.Errorf("stopped after %d redirects", maxRedirects)
				}
				return nil
			},
		}
	}

	return &LinkTool{
		httpClient:  client,
		llmProvider: llmProvider,
		maxLength:   maxLength,
	}
}

// Name returns the tool name.
func (t *LinkTool) Name() string {
	return "link"
}

// Description returns the tool description.
func (t *LinkTool) Description() string {
	return "Fetch a URL and extract its content. Supports HTML parsing, link extraction, image extraction, and optional AI summarization."
}

// Parameters returns the tool parameter schema.
func (t *LinkTool) Parameters() map[string]interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"url": map[string]interface{}{
				"type":        "string",
				"description": "URL to fetch and analyze",
			},
			"summarize": map[string]interface{}{
				"type":        "boolean",
				"description": "Use AI to summarize the content (requires LLM provider)",
				"default":     false,
			},
			"max_length": map[string]interface{}{
				"type":        "integer",
				"description": "Maximum content length in characters",
				"default":     defaultMaxLength,
			},
			"extract_images": map[string]interface{}{
				"type":        "boolean",
				"description": "Extract all image URLs from the page",
				"default":     false,
			},
			"extract_links": map[string]interface{}{
				"type":        "boolean",
				"description": "Extract all links from the page",
				"default":     false,
			},
		},
		"required": []string{"url"},
	}
}

// Execute performs the link understanding operation.
func (t *LinkTool) Execute(ctx context.Context, args map[string]interface{}) *ToolResult {
	// Extract parameters
	urlStr, ok := args["url"].(string)
	if !ok || urlStr == "" {
		return ErrorResult("url is required")
	}

	// Validate URL
	parsedURL, err := url.Parse(urlStr)
	if err != nil {
		return ErrorResult(fmt.Sprintf("invalid URL: %v", err))
	}

	if parsedURL.Scheme != "http" && parsedURL.Scheme != "https" {
		return ErrorResult("only http/https URLs are allowed")
	}

	if parsedURL.Host == "" {
		return ErrorResult("missing domain in URL")
	}

	// Extract optional parameters
	maxLength := t.maxLength
	if ml, ok := args["max_length"].(float64); ok && int(ml) > 0 {
		maxLength = int(ml)
	}

	extractImages := false
	if ei, ok := args["extract_images"].(bool); ok {
		extractImages = ei
	}

	extractLinks := false
	if el, ok := args["extract_links"].(bool); ok {
		extractLinks = el
	}

	summarize := false
	if s, ok := args["summarize"].(bool); ok {
		summarize = s
	}

	// Fetch the URL
	fetchResult, err := t.fetchURL(ctx, urlStr)
	if err != nil {
		return ErrorResult(fmt.Sprintf("failed to fetch URL: %v", err))
	}

	// Build response
	response := LinkResponse{
		URL:        urlStr,
		StatusCode: fetchResult.StatusCode,
		ContentType: fetchResult.ContentType,
		Title:      fetchResult.Title,
	}

	// Process based on content type
	if strings.Contains(fetchResult.ContentType, "application/json") || strings.Contains(fetchResult.ContentType, "application/ld+json") {
		response.Extractor = "json"
		response.Content = string(fetchResult.Body)
	} else if strings.Contains(fetchResult.ContentType, "text/html") {
		response.Extractor = "html"

		// Extract main content
		extractedText := t.extractText(string(fetchResult.Body))

		// Extract images if requested
		if extractImages {
			response.Images = t.extractImages(string(fetchResult.Body), urlStr)
		}

		// Extract links if requested
		if extractLinks {
			response.Links = t.extractLinks(string(fetchResult.Body), urlStr)
		}

		response.Content = extractedText
	} else {
		response.Extractor = "raw"
		response.Content = string(fetchResult.Body)
	}

	// Truncate content if needed
	if len(response.Content) > maxLength {
		response.Content = response.Content[:maxLength]
		response.Truncated = true
	}

	response.ContentLength = len(response.Content)

	// Summarize if requested
	if summarize {
		if t.llmProvider == nil {
			return ErrorResult("summarization requires an LLM provider to be configured")
		}

		summary, err := t.summarizeContent(ctx, response.Content, response.Title)
		if err != nil {
			return ErrorResult(fmt.Sprintf("summarization failed: %v", err))
		}
		response.Summary = summary
	}

	// Format output
	output, err := json.MarshalIndent(response, "", "  ")
	if err != nil {
		return ErrorResult(fmt.Sprintf("failed to format result: %v", err))
	}

	return &ToolResult{
		ForLLM:  fmt.Sprintf("Fetched and analyzed URL: %s", urlStr),
		ForUser: string(output),
	}
}

// fetchResult contains the raw fetch response data.
type fetchResult struct {
	StatusCode  int
	ContentType string
	Body        []byte
	Title       string
}

// fetchURL performs the HTTP request.
func (t *LinkTool) fetchURL(ctx context.Context, urlStr string) (*fetchResult, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", urlStr, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("User-Agent", defaultUserAgent)
	req.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,*/*;q=0.8")
	req.Header.Set("Accept-Language", "en-US,en;q=0.5")

	resp, err := t.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	contentType := resp.Header.Get("Content-Type")
	if contentType == "" {
		contentType = "text/html" // Default assumption
	}

	// Extract title from HTML
	title := ""
	if strings.Contains(contentType, "text/html") {
		title = extractTitle(string(body))
	}

	return &fetchResult{
		StatusCode:  resp.StatusCode,
		ContentType: contentType,
		Body:        body,
		Title:       title,
	}, nil
}

// extractTitle extracts the page title from HTML.
func extractTitle(htmlContent string) string {
	// Try <title> tag first
	re := regexp.MustCompile(`<title[^>]*>([\s\S]*?)</title>`)
	matches := re.FindStringSubmatch(htmlContent)
	if len(matches) > 1 {
		return strings.TrimSpace(stripTags(matches[1]))
	}

	// Try <h1> as fallback
	re = regexp.MustCompile(`<h1[^>]*>([\s\S]*?)</h1>`)
	matches = re.FindStringSubmatch(htmlContent)
	if len(matches) > 1 {
		return strings.TrimSpace(stripTags(matches[1]))
	}

	return ""
}

// extractText extracts readable text from HTML, removing scripts, styles, and navigation.
func (t *LinkTool) extractText(htmlContent string) string {
	// Step 1: Remove script and style tags with their content
	result := htmlContent

	// Remove script tags
	re := regexp.MustCompile(`<script[\s\S]*?</script>`)
	result = re.ReplaceAllString(result, "")

	// Remove style tags
	re = regexp.MustCompile(`<style[\s\S]*?</style>`)
	result = re.ReplaceAllString(result, "")

	// Remove noscript tags
	re = regexp.MustCompile(`<noscript[\s\S]*?</noscript>`)
	result = re.ReplaceAllString(result, "")

	// Remove SVG elements
	re = regexp.MustCompile(`<svg[\s\S]*?</svg>`)
	result = re.ReplaceAllString(result, "")

	// Remove HTML comments
	re = regexp.MustCompile(`<!--[\s\S]*?-->`)
	result = re.ReplaceAllString(result, "")

	// Step 2: Remove common navigation and footer elements
	navPatterns := []string{
		`<nav[\s\S]*?</nav>`,
		`<header[\s\S]*?</header>`,
		`<footer[\s\S]*?</footer>`,
		`<aside[\s\S]*?</aside>`,
		`<iframe[\s\S]*?</iframe>`,
		`<form[\s\S]*?</form>`,
	}

	for _, pattern := range navPatterns {
		re = regexp.MustCompile(pattern)
		result = re.ReplaceAllString(result, "")
	}

	// Step 3: Replace block elements with newlines for better formatting
	blockElements := []string{"</p>", "</div>", "</li>", "</h1>", "</h2>", "</h3>", "</h4>", "</h5>", "</h6>", "</tr>", "</br>", "<br>"}
	for _, elem := range blockElements {
		result = strings.ReplaceAll(result, elem, "\n")
	}

	// Step 4: Remove all remaining HTML tags
	re = regexp.MustCompile(`<[^>]+>`)
	result = re.ReplaceAllString(result, "")

	// Step 5: Decode HTML entities
	result = html.UnescapeString(result)

	// Step 6: Clean up whitespace
	result = strings.TrimSpace(result)

	// Replace multiple spaces/newlines with single space or newline
	re = regexp.MustCompile(`[ \t]+`)
	result = re.ReplaceAllString(result, " ")

	re = regexp.MustCompile(`\n\s*\n`)
	result = re.ReplaceAllString(result, "\n")

	// Split into lines and remove empty ones
	lines := strings.Split(result, "\n")
	var cleanLines []string
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line != "" {
			cleanLines = append(cleanLines, line)
		}
	}

	return strings.Join(cleanLines, "\n")
}

// extractImages extracts image URLs from HTML.
func (t *LinkTool) extractImages(htmlContent, baseURL string) []string {
	var images []string
	seen := make(map[string]bool)

	// Match img src and srcset
	patterns := []string{
		`<img[^>]+src=["']([^"']+)["']`,
		`<img[^>]+srcset=["']([^"']+)`,
	}

	for _, pattern := range patterns {
		re := regexp.MustCompile(pattern)
		matches := re.FindAllStringSubmatch(htmlContent, -1)

		for _, match := range matches {
			if len(match) < 2 {
				continue
			}

			// Handle srcset (contains multiple URLs)
			srcset := match[1]
			if strings.Contains(srcset, ",") {
				parts := strings.Split(srcset, ",")
				for _, part := range parts {
					imgURL := strings.TrimSpace(strings.Split(part, " ")[0])
					if imgURL != "" && !seen[imgURL] {
						seen[imgURL] = true
						images = append(images, resolveURL(imgURL, baseURL))
					}
				}
			} else {
				imgURL := match[1]
				if imgURL != "" && !seen[imgURL] {
					seen[imgURL] = true
					images = append(images, resolveURL(imgURL, baseURL))
				}
			}
		}
	}

	// Limit to reasonable number
	if len(images) > 20 {
		images = images[:20]
	}

	return images
}

// extractLinks extracts all links from HTML.
func (t *LinkTool) extractLinks(htmlContent, baseURL string) []string {
	var links []string
	seen := make(map[string]bool)

	re := regexp.MustCompile(`<a[^>]+href=["']([^"']+)["'][^>]*>`)
	matches := re.FindAllStringSubmatch(htmlContent, -1)

	for _, match := range matches {
		if len(match) < 2 {
			continue
		}

		linkURL := match[1]
		// Skip empty, javascript, mailto, and anchor links
		if linkURL == "" ||
			strings.HasPrefix(linkURL, "javascript:") ||
			strings.HasPrefix(linkURL, "mailto:") ||
			strings.HasPrefix(linkURL, "#") ||
			strings.HasPrefix(linkURL, "tel:") {
			continue
		}

		resolvedURL := resolveURL(linkURL, baseURL)
		if !seen[resolvedURL] {
			seen[resolvedURL] = true
			links = append(links, resolvedURL)
		}
	}

	// Limit to reasonable number
	if len(links) > 50 {
		links = links[:50]
	}

	return links
}

// resolveURL resolves a potentially relative URL against a base URL.
func resolveURL(linkURL, baseURL string) string {
	// If already absolute, return as-is
	if strings.HasPrefix(linkURL, "http://") || strings.HasPrefix(linkURL, "https://") {
		return linkURL
	}

	base, err := url.Parse(baseURL)
	if err != nil {
		return linkURL
	}

	// Handle absolute paths
	if strings.HasPrefix(linkURL, "//") {
		return base.Scheme + ":" + linkURL
	}

	// Resolve relative URL
	resolved := base.ResolveReference(&url.URL{Path: linkURL})
	return resolved.String()
}

// summarizeContent uses AI to summarize the extracted content.
func (t *LinkTool) summarizeContent(ctx context.Context, content, title string) (string, error) {
	if t.llmProvider == nil {
		return "", fmt.Errorf("no LLM provider configured")
	}

	// Truncate content for summary
	summaryContent := content
	if len(summaryContent) > defaultSummaryLength*2 {
		summaryContent = summaryContent[:defaultSummaryLength*2]
	}

	prompt := fmt.Sprintf(`Please summarize the following web page content in a concise way (2-3 sentences). 
Focus on the main points and key information.

Page Title: %s

Content:
%s

Summary:`, title, summaryContent)

	// Use a timeout for the summary request
	summaryCtx, cancel := context.WithTimeout(ctx, defaultSummaryTimeout)
	defer cancel()

	messages := []providers.Message{
		{
			Role:    "user",
			Content: prompt,
		},
	}

	response, err := t.llmProvider.Chat(summaryCtx, messages, nil, "", nil)
	if err != nil {
		return "", fmt.Errorf("LLM request failed: %w", err)
	}

	if response == nil || response.Content == "" {
		return "", fmt.Errorf("empty response from LLM")
	}

	return strings.TrimSpace(response.Content), nil
}

// LinkResponse represents the structured response from the LinkTool.
type LinkResponse struct {
	URL           string   `json:"url"`
	StatusCode    int      `json:"status_code"`
	ContentType   string   `json:"content_type"`
	Title         string   `json:"title,omitempty"`
	Extractor     string   `json:"extractor"`
	Content       string   `json:"content"`
	ContentLength int      `json:"content_length"`
	Truncated     bool     `json:"truncated"`
	Summary       string   `json:"summary,omitempty"`
	Images        []string `json:"images,omitempty"`
	Links         []string `json:"links,omitempty"`
}

// LinkToolWithDefaultProvider creates a LinkTool with a default HTTP client.
// This is a convenience function for simple use cases.
func LinkToolWithDefaultProvider(llmProvider providers.LLMProvider) *LinkTool {
	return NewLinkTool(llmProvider, LinkToolOptions{})
}

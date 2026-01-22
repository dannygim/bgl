package backlog

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/dannygim/bgl/internal/auth"
	"github.com/dannygim/bgl/internal/config"
)

// Client is a Backlog API client with automatic token management.
type Client struct {
	cfg        *config.Config
	httpClient *http.Client
}

// NewClient creates a new Backlog API client.
// It checks token expiration and refreshes if needed.
func NewClient() (*Client, error) {
	cfg, err := config.Load()
	if err != nil {
		return nil, fmt.Errorf("failed to load config: %w", err)
	}

	if cfg.AccessToken == "" {
		return nil, fmt.Errorf("not logged in. Please run 'bgl auth login' first")
	}

	// Check if token is expired and refresh if needed
	if cfg.ExpiresAt > 0 && time.Now().UnixMilli() >= cfg.ExpiresAt {
		if err := auth.RefreshToken(); err != nil {
			return nil, fmt.Errorf("failed to refresh token: %w", err)
		}
		// Reload config after refresh
		cfg, err = config.Load()
		if err != nil {
			return nil, fmt.Errorf("failed to reload config: %w", err)
		}
	}

	return &Client{
		cfg:        cfg,
		httpClient: &http.Client{Timeout: 30 * time.Second},
	}, nil
}

// doRequest performs an HTTP request with authentication and error handling.
func (c *Client) doRequest(method, path string) ([]byte, error) {
	url := fmt.Sprintf("https://%s%s", c.cfg.Space, path)

	req, err := http.NewRequest(method, url, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", "Bearer "+c.cfg.AccessToken)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	// Handle authentication errors
	if resp.StatusCode == http.StatusUnauthorized {
		wwwAuth := resp.Header.Get("WWW-Authenticate")
		if strings.Contains(wwwAuth, "The access token expired") {
			// Token expired - try to refresh
			if err := auth.RefreshToken(); err != nil {
				return nil, fmt.Errorf("access token expired and refresh failed: %w. Please run 'bgl auth login'", err)
			}
			// Reload config and retry
			cfg, err := config.Load()
			if err != nil {
				return nil, fmt.Errorf("failed to reload config: %w", err)
			}
			c.cfg = cfg
			return c.doRequest(method, path)
		}
		if strings.Contains(wwwAuth, "The access token is invalid") {
			return nil, fmt.Errorf("access token is invalid. Please run 'bgl auth login'")
		}
		return nil, fmt.Errorf("authentication failed (status %d). Please run 'bgl auth login'", resp.StatusCode)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API request failed with status %d: %s", resp.StatusCode, string(body))
	}

	return body, nil
}

// GetIssue retrieves an issue by its key or ID.
func (c *Client) GetIssue(issueKeyOrID string) ([]byte, error) {
	return c.doRequest("GET", "/api/v2/issues/"+issueKeyOrID)
}

// Issue represents a Backlog issue.
type Issue struct {
	Summary     string    `json:"summary"`
	Description string    `json:"description"`
	Assignee    *Assignee `json:"assignee"`
	Status      *Status   `json:"status"`
}

// Assignee represents the assignee of an issue.
type Assignee struct {
	Name        string `json:"name"`
	MailAddress string `json:"mailAddress"`
}

// Status represents the status of an issue.
type Status struct {
	Name string `json:"name"`
}

// ParseIssue parses the JSON response into an Issue struct.
func ParseIssue(data []byte) (*Issue, error) {
	var issue Issue
	if err := json.Unmarshal(data, &issue); err != nil {
		return nil, fmt.Errorf("failed to parse issue: %w", err)
	}
	return &issue, nil
}

// FormatIssueMarkdown formats the issue as Markdown.
func FormatIssueMarkdown(issue *Issue) string {
	var sb strings.Builder

	fmt.Fprintf(&sb, "## Summary\n%s\n\n", issue.Summary)

	sb.WriteString("## Assignee\n")
	if issue.Assignee != nil {
		fmt.Fprintf(&sb, "%s<%s>\n\n", issue.Assignee.Name, issue.Assignee.MailAddress)
	} else {
		sb.WriteString("(unassigned)\n\n")
	}

	sb.WriteString("## Status\n")
	if issue.Status != nil {
		fmt.Fprintf(&sb, "%s\n\n", issue.Status.Name)
	} else {
		sb.WriteString("(unknown)\n\n")
	}

	sb.WriteString("## Description\n")
	if issue.Description != "" {
		sb.WriteString(issue.Description)
	} else {
		sb.WriteString("(no description)")
	}
	sb.WriteString("\n")

	return sb.String()
}

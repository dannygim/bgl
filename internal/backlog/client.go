package backlog

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
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

// GetComments retrieves comments for an issue.
func (c *Client) GetComments(issueKeyOrID string) ([]byte, error) {
	return c.doRequest("GET", "/api/v2/issues/"+issueKeyOrID+"/comments")
}

// GetComment retrieves a specific comment by ID.
func (c *Client) GetComment(issueKeyOrID string, commentID string) ([]byte, error) {
	return c.doRequest("GET", "/api/v2/issues/"+issueKeyOrID+"/comments/"+commentID)
}

// doPostRequest performs an HTTP POST request with form data.
func (c *Client) doPostRequest(path string, data url.Values) ([]byte, error) {
	apiURL := fmt.Sprintf("https://%s%s", c.cfg.Space, path)

	req, err := http.NewRequest("POST", apiURL, strings.NewReader(data.Encode()))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", "Bearer "+c.cfg.AccessToken)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

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
			return c.doPostRequest(path, data)
		}
		if strings.Contains(wwwAuth, "The access token is invalid") {
			return nil, fmt.Errorf("access token is invalid. Please run 'bgl auth login'")
		}
		return nil, fmt.Errorf("authentication failed (status %d). Please run 'bgl auth login'", resp.StatusCode)
	}

	if resp.StatusCode != http.StatusCreated && resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API request failed with status %d: %s", resp.StatusCode, string(body))
	}

	return body, nil
}

// AddComment adds a comment to an issue.
func (c *Client) AddComment(issueKeyOrID string, content string) ([]byte, error) {
	data := url.Values{}
	data.Set("content", content)
	return c.doPostRequest("/api/v2/issues/"+issueKeyOrID+"/comments", data)
}

// GetSpace returns the space domain from the client config.
func (c *Client) GetSpace() string {
	return c.cfg.Space
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

// Comment represents a Backlog comment.
type Comment struct {
	ID          int          `json:"id"`
	Content     string       `json:"content"`
	CreatedUser *CommentUser `json:"createdUser"`
	Created     string       `json:"created"`
}

// CommentUser represents the user who created a comment.
type CommentUser struct {
	Name        string `json:"name"`
	MailAddress string `json:"mailAddress"`
}

// ParseComment parses the JSON response into a Comment struct.
func ParseComment(data []byte) (*Comment, error) {
	var comment Comment
	if err := json.Unmarshal(data, &comment); err != nil {
		return nil, fmt.Errorf("failed to parse comment: %w", err)
	}
	return &comment, nil
}

// ParseComments parses the JSON response into a slice of Comment structs.
func ParseComments(data []byte) ([]Comment, error) {
	var comments []Comment
	if err := json.Unmarshal(data, &comments); err != nil {
		return nil, fmt.Errorf("failed to parse comments: %w", err)
	}
	return comments, nil
}

// FormatCommentMarkdown formats a single comment as Markdown.
func FormatCommentMarkdown(comment *Comment) string {
	var sb strings.Builder

	fmt.Fprintf(&sb, "**Comment Id:** %d\n\n", comment.ID)

	sb.WriteString("**User:** ")
	if comment.CreatedUser != nil {
		fmt.Fprintf(&sb, "%s<%s>\n\n", comment.CreatedUser.Name, comment.CreatedUser.MailAddress)
	} else {
		sb.WriteString("(unknown)\n\n")
	}

	fmt.Fprintf(&sb, "**Datetime:** %s\n\n", comment.Created)

	sb.WriteString("**Content:**\n")
	if comment.Content != "" {
		sb.WriteString(comment.Content)
	} else {
		sb.WriteString("(no content)")
	}
	sb.WriteString("\n")

	return sb.String()
}

// FormatCommentsMarkdown formats a list of comments as Markdown.
func FormatCommentsMarkdown(comments []Comment) string {
	var sb strings.Builder

	for i, comment := range comments {
		sb.WriteString(FormatCommentMarkdown(&comment))
		if i < len(comments)-1 {
			sb.WriteString("\n---\n\n")
		}
	}

	return sb.String()
}

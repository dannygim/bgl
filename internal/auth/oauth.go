package auth

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/dannygim/bgl/internal/config"
)

const (
	callbackPort = 18765
)

// TokenResponse represents the OAuth token response from Backlog.
type TokenResponse struct {
	AccessToken  string `json:"access_token"`
	TokenType    string `json:"token_type"`
	ExpiresIn    int    `json:"expires_in"`
	RefreshToken string `json:"refresh_token"`
}

// ValidateSpace validates the space format.
func ValidateSpace(space string) error {
	if !strings.HasSuffix(space, ".backlog.com") && !strings.HasSuffix(space, ".backlog.jp") {
		return fmt.Errorf("invalid space format: must be <your-space-key>.backlog.com or <your-space-key>.backlog.jp")
	}
	return nil
}

// getBacklogBaseURL returns the Backlog base URL for the given space.
func getBacklogBaseURL(space string) string {
	return "https://" + space
}

// generateState creates a random state string for CSRF protection.
func generateState() (string, error) {
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}

// Login performs the OAuth 2.0 login flow.
func Login(space string) error {
	if err := ValidateSpace(space); err != nil {
		return err
	}

	if config.ClientID == "" || config.ClientSecret == "" {
		return fmt.Errorf("ClientID and ClientSecret are not set. Please build with proper ldflags")
	}

	baseURL := getBacklogBaseURL(space)
	redirectURI := fmt.Sprintf("http://localhost:%d", callbackPort)

	state, err := generateState()
	if err != nil {
		return fmt.Errorf("failed to generate state: %w", err)
	}

	authURL := fmt.Sprintf("%s/OAuth2AccessRequest.action?response_type=code&client_id=%s&redirect_uri=%s&state=%s",
		baseURL,
		url.QueryEscape(config.ClientID),
		url.QueryEscape(redirectURI),
		url.QueryEscape(state),
	)

	codeChan := make(chan string, 1)
	errChan := make(chan error, 1)

	server := &http.Server{
		Addr: fmt.Sprintf(":%d", callbackPort),
	}

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		receivedState := r.URL.Query().Get("state")
		if receivedState != state {
			errChan <- fmt.Errorf("state mismatch: expected %s, got %s", state, receivedState)
			http.Error(w, "State mismatch", http.StatusBadRequest)
			return
		}

		code := r.URL.Query().Get("code")
		if code == "" {
			errChan <- fmt.Errorf("no authorization code received")
			http.Error(w, "No authorization code", http.StatusBadRequest)
			return
		}

		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		fmt.Fprint(w, "<html><body><h1>Login successful!</h1><p>You can close this window.</p></body></html>")
		codeChan <- code
	})

	listener, err := net.Listen("tcp", fmt.Sprintf(":%d", callbackPort))
	if err != nil {
		return fmt.Errorf("failed to start callback server: %w", err)
	}

	go func() {
		if err := server.Serve(listener); err != http.ErrServerClosed {
			errChan <- err
		}
	}()

	fmt.Println("Opening browser for authentication...")
	fmt.Printf("If browser doesn't open automatically, please visit:\n%s\n\n", authURL)

	if err := openBrowser(authURL); err != nil {
		fmt.Printf("Failed to open browser: %v\n", err)
	}

	fmt.Println("Waiting for authentication...")

	var code string
	select {
	case code = <-codeChan:
	case err := <-errChan:
		server.Shutdown(context.Background())
		return err
	case <-time.After(5 * time.Minute):
		server.Shutdown(context.Background())
		return fmt.Errorf("authentication timeout")
	}

	server.Shutdown(context.Background())

	token, err := exchangeCode(baseURL, code, redirectURI)
	if err != nil {
		return fmt.Errorf("failed to exchange code: %w", err)
	}

	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	cfg.Space = space
	cfg.AccessToken = token.AccessToken
	cfg.RefreshToken = token.RefreshToken

	if err := cfg.Save(); err != nil {
		return fmt.Errorf("failed to save config: %w", err)
	}

	fmt.Println("Login successful! Tokens saved to config.")
	return nil
}

// exchangeCode exchanges the authorization code for tokens.
func exchangeCode(baseURL, code, redirectURI string) (*TokenResponse, error) {
	tokenURL := baseURL + "/api/v2/oauth2/token"

	data := url.Values{}
	data.Set("grant_type", "authorization_code")
	data.Set("code", code)
	data.Set("redirect_uri", redirectURI)
	data.Set("client_id", config.ClientID)
	data.Set("client_secret", config.ClientSecret)

	resp, err := http.PostForm(tokenURL, data)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("token request failed with status: %d", resp.StatusCode)
	}

	var token TokenResponse
	if err := json.NewDecoder(resp.Body).Decode(&token); err != nil {
		return nil, err
	}

	return &token, nil
}

// RefreshToken refreshes the access token using the refresh token.
func RefreshToken() error {
	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	if cfg.RefreshToken == "" {
		return fmt.Errorf("no refresh token found. Please run 'bgl auth login' first")
	}

	baseURL := getBacklogBaseURL(cfg.Space)
	tokenURL := baseURL + "/api/v2/oauth2/token"

	data := url.Values{}
	data.Set("grant_type", "refresh_token")
	data.Set("client_id", config.ClientID)
	data.Set("client_secret", config.ClientSecret)
	data.Set("refresh_token", cfg.RefreshToken)

	resp, err := http.PostForm(tokenURL, data)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("token refresh failed with status: %d", resp.StatusCode)
	}

	var token TokenResponse
	if err := json.NewDecoder(resp.Body).Decode(&token); err != nil {
		return err
	}

	cfg.AccessToken = token.AccessToken
	cfg.RefreshToken = token.RefreshToken

	if err := cfg.Save(); err != nil {
		return fmt.Errorf("failed to save config: %w", err)
	}

	return nil
}

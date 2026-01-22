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

	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
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

// inputModel is the bubbletea model for text input.
type inputModel struct {
	textInput textinput.Model
	err       error
	done      bool
	cancelled bool
}

func newInputModel() inputModel {
	ti := textinput.New()
	ti.Placeholder = "myspace.backlog.com"
	ti.Focus()
	ti.CharLimit = 100
	ti.Width = 40
	return inputModel{
		textInput: ti,
	}
}

func (m inputModel) Init() tea.Cmd {
	return textinput.Blink
}

func (m inputModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "enter":
			value := m.textInput.Value()
			if err := ValidateSpace(value); err != nil {
				m.err = err
				return m, nil
			}
			m.done = true
			return m, tea.Quit
		case "ctrl+c", "esc":
			m.cancelled = true
			return m, tea.Quit
		}
	}

	m.textInput, cmd = m.textInput.Update(msg)
	m.err = nil
	return m, cmd
}

func (m inputModel) View() string {
	var s strings.Builder
	s.WriteString("Enter your Backlog space:\n\n")
	s.WriteString(m.textInput.View())
	s.WriteString("\n\n")
	if m.err != nil {
		errStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("9"))
		s.WriteString(errStyle.Render(m.err.Error()))
		s.WriteString("\n")
	}
	s.WriteString("(press Enter to confirm, Esc to cancel)\n")
	return s.String()
}

// authResult represents the result of the authentication process.
type authResult struct {
	code string
	err  error
}

// spinnerModel is the bubbletea model for the spinner.
type spinnerModel struct {
	spinner    spinner.Model
	message    string
	done       bool
	code       string
	err        error
	resultChan <-chan authResult
}

func newSpinnerModel(message string, resultChan <-chan authResult) spinnerModel {
	s := spinner.New()
	s.Spinner = spinner.Dot
	s.Style = lipgloss.NewStyle().Foreground(lipgloss.Color("205"))
	return spinnerModel{
		spinner:    s,
		message:    message,
		resultChan: resultChan,
	}
}

func (m spinnerModel) Init() tea.Cmd {
	return tea.Batch(m.spinner.Tick, m.waitForResult())
}

type resultMsg authResult

func (m spinnerModel) waitForResult() tea.Cmd {
	return func() tea.Msg {
		result := <-m.resultChan
		return resultMsg(result)
	}
}

func (m spinnerModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		if msg.String() == "ctrl+c" || msg.String() == "q" {
			m.err = fmt.Errorf("cancelled by user")
			return m, tea.Quit
		}
	case resultMsg:
		m.done = true
		m.code = msg.code
		m.err = msg.err
		return m, tea.Quit
	case spinner.TickMsg:
		var cmd tea.Cmd
		m.spinner, cmd = m.spinner.Update(msg)
		return m, cmd
	}
	return m, nil
}

func (m spinnerModel) View() string {
	if m.done {
		return ""
	}
	return fmt.Sprintf("%s %s\n", m.spinner.View(), m.message)
}

// Login performs the OAuth 2.0 login flow.
func Login() error {
	// Get space from user input
	im := newInputModel()
	p := tea.NewProgram(im)
	finalModel, err := p.Run()
	if err != nil {
		return fmt.Errorf("input error: %w", err)
	}

	m := finalModel.(inputModel)
	if m.cancelled {
		return fmt.Errorf("cancelled by user")
	}

	space := m.textInput.Value()

	if config.ClientID == "" || config.ClientSecret == "" {
		return fmt.Errorf("OAuth client credentials are not configured. Please build with the required configuration flags")
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

	resultChan := make(chan authResult, 1)

	server := &http.Server{
		Addr: fmt.Sprintf(":%d", callbackPort),
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		receivedState := r.URL.Query().Get("state")
		if receivedState != state {
			resultChan <- authResult{err: fmt.Errorf("state mismatch: expected %s, got %s", state, receivedState)}
			http.Error(w, "State mismatch", http.StatusBadRequest)
			return
		}

		code := r.URL.Query().Get("code")
		if code == "" {
			resultChan <- authResult{err: fmt.Errorf("no authorization code received")}
			http.Error(w, "No authorization code", http.StatusBadRequest)
			return
		}

		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		fmt.Fprint(w, "<html><body><h1>Login successful!</h1><p>You can close this window.</p></body></html>")
		resultChan <- authResult{code: code}
	})
	server.Handler = mux

	listener, err := net.Listen("tcp", fmt.Sprintf(":%d", callbackPort))
	if err != nil {
		return fmt.Errorf("failed to start callback server: %w", err)
	}

	go func() {
		if err := server.Serve(listener); err != http.ErrServerClosed {
			resultChan <- authResult{err: err}
		}
	}()

	go func() {
		time.Sleep(5 * time.Minute)
		select {
		case resultChan <- authResult{err: fmt.Errorf("authentication timeout")}:
		default:
		}
	}()

	fmt.Println("\nOpening browser for authentication...")
	fmt.Printf("If browser doesn't open automatically, please visit:\n%s\n\n", authURL)

	if err := openBrowser(authURL); err != nil {
		fmt.Printf("Failed to open browser: %v\n", err)
	}

	// Ensure the HTTP server is shut down on all exit paths after this point.
	shutdownCtx, shutdownCancel := context.WithCancel(context.Background())
	defer shutdownCancel()
	go func() {
		<-shutdownCtx.Done()
		_ = server.Shutdown(context.Background())
	}()

	sp := newSpinnerModel("Waiting for authentication...", resultChan)
	p = tea.NewProgram(sp)
	finalSpinnerModel, err := p.Run()
	if err != nil {
		return fmt.Errorf("spinner error: %w", err)
	}
	sm := finalSpinnerModel.(spinnerModel)
	if sm.err != nil {
		return sm.err
	}

	token, err := exchangeCode(baseURL, sm.code, redirectURI)
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

// Logout removes the stored access token and refresh token.
func Logout() error {
	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	if cfg.AccessToken == "" && cfg.RefreshToken == "" {
		return fmt.Errorf("not logged in")
	}

	cfg.AccessToken = ""
	cfg.RefreshToken = ""

	if err := cfg.Save(); err != nil {
		return fmt.Errorf("failed to save config: %w", err)
	}

	fmt.Println("Logged out successfully.")
	return nil
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

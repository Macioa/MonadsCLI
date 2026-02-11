package lucid

import (
	"bytes"
	"context"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"os/exec"
	"runtime"
	"strings"
	"time"

	"golang.org/x/oauth2"
)

const (
	DefaultAuthURL    = "https://lucid.app/oauth2/authorize"
	AccountAuthURL    = "https://lucid.app/oauth2/authorizeAccount"
	DefaultTokenURL   = "https://api.lucid.co/oauth2/token"
	DefaultListenHost = "localhost:8787"
)

type LoginOptions struct {
	ClientID      string
	ClientSecret  string
	RedirectURL   string
	Scopes        []string
	AuthURL       string
	TokenURL      string
	State         string
	Prompt        string
	OpenBrowser   bool
	ListenTimeout time.Duration
	Output        io.Writer
}

func Login(ctx context.Context, opts LoginOptions) (*oauth2.Token, error) {
	if strings.TrimSpace(opts.ClientID) == "" {
		return nil, errors.New("lucid login: missing client id")
	}
	if strings.TrimSpace(opts.ClientSecret) == "" {
		return nil, errors.New("lucid login: missing client secret")
	}
	redirectURL := strings.TrimSpace(opts.RedirectURL)
	if redirectURL == "" {
		redirectURL = fmt.Sprintf("http://%s/lucid/callback", DefaultListenHost)
	}
	authURL := strings.TrimSpace(opts.AuthURL)
	if authURL == "" {
		authURL = DefaultAuthURL
	}
	tokenURL := strings.TrimSpace(opts.TokenURL)
	if tokenURL == "" {
		tokenURL = DefaultTokenURL
	}

	state := strings.TrimSpace(opts.State)
	if state == "" {
		var err error
		state, err = randomState()
		if err != nil {
			return nil, err
		}
	}

	scopes := opts.Scopes
	if len(scopes) == 0 {
		// Lucid requires at least one scope. Default: view folders and documents.
		scopes = []string{
			"user.profile",
			"folder:readonly",                      // view folders and list contents
			"lucidchart.document.content:readonly", // view and download documents
		}
	}

	config := oauth2.Config{
		ClientID:     opts.ClientID,
		ClientSecret: opts.ClientSecret,
		RedirectURL:  redirectURL,
		Scopes:       scopes,
		Endpoint: oauth2.Endpoint{
			AuthURL:   authURL,
			TokenURL:  tokenURL,
			AuthStyle: oauth2.AuthStyleInParams, // Lucid expects client_id/client_secret in POST body
		},
	}

	authOptions := []oauth2.AuthCodeOption{}
	if strings.TrimSpace(opts.Prompt) != "" {
		authOptions = append(authOptions, oauth2.SetAuthURLParam("prompt", opts.Prompt))
	}
	authLink := config.AuthCodeURL(state, authOptions...)

	if opts.Output != nil {
		fmt.Fprintln(opts.Output, "Using redirect_uri:", redirectURL)
		fmt.Fprintln(opts.Output, "Open the following URL to authorize Lucidchart:")
		fmt.Fprintln(opts.Output, authLink)
	}
	if opts.OpenBrowser {
		_ = openBrowser(authLink)
	}

	timeout := opts.ListenTimeout
	if timeout == 0 {
		timeout = 2 * time.Minute
	}

	code, err := waitForCode(ctx, redirectURL, state, timeout)
	if err != nil {
		return nil, err
	}

	token, err := exchangeToken(ctx, tokenURL, strings.TrimSpace(opts.ClientID), strings.TrimSpace(opts.ClientSecret), redirectURL, code)
	if err != nil {
		return nil, fmt.Errorf("lucid login: exchange token: %w", err)
	}

	return token, nil
}

// exchangeToken exchanges the auth code for a token. Uses JSON body per Lucid docs.
func exchangeToken(ctx context.Context, tokenURL, clientID, clientSecret, redirectURL, code string) (*oauth2.Token, error) {
	payload := struct {
		GrantType    string `json:"grant_type"`
		Code         string `json:"code"`
		ClientID     string `json:"client_id"`
		ClientSecret string `json:"client_secret"`
		RedirectURI  string `json:"redirect_uri"`
	}{
		GrantType:    "authorization_code",
		Code:         code,
		ClientID:     clientID,
		ClientSecret: clientSecret,
		RedirectURI:  redirectURL,
	}
	bodyBytes, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, tokenURL, bytes.NewReader(bodyBytes))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	dec := json.NewDecoder(resp.Body)
	var raw struct {
		AccessToken  string `json:"access_token"`
		TokenType    string `json:"token_type"`
		RefreshToken string `json:"refresh_token"`
		ExpiresIn    int64  `json:"expires_in"`
		Error        string `json:"error"`
		ErrorDesc    string `json:"error_description"`
	}
	if err := dec.Decode(&raw); err != nil {
		return nil, err
	}
	if raw.Error != "" {
		err := fmt.Errorf("oauth2: %q %s", raw.Error, raw.ErrorDesc)
		if raw.Error == "invalid_client" {
			err = fmt.Errorf("%w — In Lucid Developer Portal (lucid.app/developer) check: 1) Redirect URI list includes exactly: http://localhost:8787/lucid/callback  2) This OAuth client is enabled  3) Client ID and secret are from the same client; if you reset the secret, update .env and run make build", err)
		}
		return nil, err
	}
	if raw.AccessToken == "" {
		return nil, errors.New("oauth2: no access_token in response")
	}

	tok := &oauth2.Token{
		AccessToken:  raw.AccessToken,
		TokenType:    raw.TokenType,
		RefreshToken: raw.RefreshToken,
	}
	if raw.ExpiresIn > 0 {
		tok.Expiry = time.Now().Add(time.Duration(raw.ExpiresIn) * time.Second)
	}
	return tok, nil
}

func waitForCode(ctx context.Context, redirectURL, state string, timeout time.Duration) (string, error) {
	parsed, err := url.Parse(redirectURL)
	if err != nil {
		return "", fmt.Errorf("lucid login: invalid redirect url: %w", err)
	}
	if parsed.Host == "" {
		return "", errors.New("lucid login: redirect url missing host")
	}
	if !isLocalhost(parsed.Hostname()) {
		return "", errors.New("lucid login: redirect url must use localhost for local callback")
	}

	path := parsed.Path
	if path == "" {
		path = "/"
	}

	codeCh := make(chan string, 1)
	errCh := make(chan error, 1)

	mux := http.NewServeMux()
	mux.HandleFunc(path, func(w http.ResponseWriter, r *http.Request) {
		query := r.URL.Query()
		if errText := strings.TrimSpace(query.Get("error")); errText != "" {
			errCh <- fmt.Errorf("lucid login: authorization error: %s", errText)
			http.Error(w, "Authorization failed. You can close this window.", http.StatusBadRequest)
			return
		}
		code := strings.TrimSpace(query.Get("code"))
		if code == "" {
			errCh <- errors.New("lucid login: missing code in callback")
			http.Error(w, "Missing authorization code. You can close this window.", http.StatusBadRequest)
			return
		}
		if state != "" && strings.TrimSpace(query.Get("state")) != state {
			errCh <- errors.New("lucid login: state mismatch")
			http.Error(w, "State mismatch. You can close this window.", http.StatusBadRequest)
			return
		}

		fmt.Fprintln(w, "Lucidchart authorization complete. You can close this window.")
		codeCh <- code
	})

	listener, err := net.Listen("tcp", parsed.Host)
	if err != nil {
		return "", fmt.Errorf("lucid login: listen on %s: %w", parsed.Host, err)
	}
	defer listener.Close()

	server := &http.Server{
		Handler: mux,
	}

	serverErr := make(chan error, 1)
	go func() {
		serverErr <- server.Serve(listener)
	}()

	defer func() {
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()
		_ = server.Shutdown(shutdownCtx)
	}()

	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	select {
	case code := <-codeCh:
		return code, nil
	case err := <-errCh:
		return "", err
	case err := <-serverErr:
		if errors.Is(err, http.ErrServerClosed) {
			return "", errors.New("lucid login: server closed before callback")
		}
		return "", fmt.Errorf("lucid login: callback server error: %w", err)
	case <-ctx.Done():
		return "", errors.New("lucid login: timed out waiting for callback")
	}
}

func randomState() (string, error) {
	raw := make([]byte, 32)
	if _, err := rand.Read(raw); err != nil {
		return "", fmt.Errorf("lucid login: generate state: %w", err)
	}
	return base64.RawURLEncoding.EncodeToString(raw), nil
}

func isLocalhost(host string) bool {
	if host == "" {
		return false
	}
	host = strings.ToLower(strings.TrimSpace(host))
	return host == "localhost" || host == "127.0.0.1" || host == "::1"
}

func openBrowser(link string) error {
	switch runtime.GOOS {
	case "darwin":
		return exec.Command("open", link).Start()
	case "windows":
		return exec.Command("rundll32", "url.dll,FileProtocolHandler", link).Start()
	default:
		return exec.Command("xdg-open", link).Start()
	}
}

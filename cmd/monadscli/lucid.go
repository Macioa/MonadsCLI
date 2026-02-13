package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/ryanmontgomery/MonadsCLI/internal/cli"
	"github.com/ryanmontgomery/MonadsCLI/internal/envembed"
	"github.com/ryanmontgomery/MonadsCLI/internal/lucid"
	"github.com/ryanmontgomery/MonadsCLI/internal/settings"
)

func lucidCommand() cli.Command {
	return cli.Command{
		Name:        "lucid",
		Description: "Lucidchart OAuth helpers",
		Run: func(fs *flag.FlagSet) error {
			args := fs.Args()
			if len(args) == 0 {
				return fmt.Errorf("missing lucid subcommand")
			}
			switch args[0] {
			case "login":
				return lucidLogin(args[1:])
			case "document":
				return lucidDocument(args[1:])
			default:
				return fmt.Errorf("unknown lucid subcommand: %s", args[0])
			}
		},
	}
}

func lucidLogin(args []string) error {
	fs := flag.NewFlagSet("lucid login", flag.ContinueOnError)
	fs.SetOutput(os.Stdout)

	var clientID string
	var clientSecret string
	var redirectURL string
	var authURL string
	var tokenURL string
	var prompt string
	var state string
	var timeout time.Duration
	var openBrowser bool
	var account bool
	var scopes cli.StringList

	fs.StringVar(&clientID, "client-id", "", "Lucid OAuth client ID (or LUCID_OAUTH_CLIENT_ID)")
	fs.StringVar(&clientSecret, "client-secret", "", "Lucid OAuth client secret (or LUCID_OAUTH_CLIENT_SECRET)")
	fs.StringVar(&redirectURL, "redirect-url", "", "Redirect URL (or LUCID_OAUTH_REDIRECT_URL)")
	fs.StringVar(&authURL, "auth-url", "", "Auth URL override (or LUCID_OAUTH_AUTH_URL)")
	fs.StringVar(&tokenURL, "token-url", "", "Token URL override (or LUCID_OAUTH_TOKEN_URL)")
	fs.StringVar(&prompt, "prompt", "", "Prompt parameter to pass to Lucid")
	fs.StringVar(&state, "state", "", "OAuth2 state override")
	fs.DurationVar(&timeout, "timeout", 2*time.Minute, "Timeout waiting for callback")
	fs.BoolVar(&openBrowser, "open-browser", true, "Open authorization URL in a browser")
	fs.BoolVar(&account, "account", false, "Use the account authorization endpoint")
	fs.Var(&scopes, "scope", "OAuth scope (repeatable). Can also use LUCID_OAUTH_SCOPES.")

	if err := fs.Parse(args); err != nil {
		if errors.Is(err, flag.ErrHelp) {
			return nil
		}
		return err
	}

	if clientID == "" {
		clientID = strings.TrimSpace(os.Getenv("LUCID_OAUTH_CLIENT_ID"))
	}
	if clientID == "" {
		clientID = envembed.LucidOAuthClientID
	}
	if clientSecret == "" {
		clientSecret = strings.TrimSpace(os.Getenv("LUCID_OAUTH_CLIENT_SECRET"))
	}
	if clientSecret == "" {
		clientSecret = envembed.LucidOAuthClientSecret
	}
	if redirectURL == "" {
		redirectURL = strings.TrimSpace(os.Getenv("LUCID_OAUTH_REDIRECT_URL"))
	}
	if redirectURL == "" {
		redirectURL = envembed.LucidOAuthRedirectURL
	}
	if authURL == "" {
		authURL = strings.TrimSpace(os.Getenv("LUCID_OAUTH_AUTH_URL"))
	}
	if authURL == "" {
		authURL = envembed.LucidOAuthAuthURL
	}
	if tokenURL == "" {
		tokenURL = strings.TrimSpace(os.Getenv("LUCID_OAUTH_TOKEN_URL"))
	}
	if tokenURL == "" {
		tokenURL = envembed.LucidOAuthTokenURL
	}
	if prompt == "" {
		prompt = strings.TrimSpace(os.Getenv("LUCID_OAUTH_PROMPT"))
	}
	if prompt == "" {
		prompt = envembed.LucidOAuthPrompt
	}
	if state == "" {
		state = strings.TrimSpace(os.Getenv("LUCID_OAUTH_STATE"))
	}
	if state == "" {
		state = envembed.LucidOAuthState
	}
	if len(scopes) == 0 {
		envScopes := strings.TrimSpace(os.Getenv("LUCID_OAUTH_SCOPES"))
		if envScopes != "" {
			for _, scope := range strings.Split(envScopes, ",") {
				if trimmed := strings.TrimSpace(scope); trimmed != "" {
					scopes = append(scopes, trimmed)
				}
			}
		}
	}
	if len(scopes) == 0 && envembed.LucidOAuthScopes != "" {
		for _, scope := range strings.Split(envembed.LucidOAuthScopes, ",") {
			if trimmed := strings.TrimSpace(scope); trimmed != "" {
				scopes = append(scopes, trimmed)
			}
		}
	}
	if authURL == "" && account {
		authURL = lucid.AccountAuthURL
	}

	token, err := lucid.Login(context.Background(), lucid.LoginOptions{
		ClientID:      clientID,
		ClientSecret:  clientSecret,
		RedirectURL:   redirectURL,
		Scopes:        []string(scopes),
		AuthURL:       authURL,
		TokenURL:      tokenURL,
		State:         state,
		Prompt:        prompt,
		OpenBrowser:   openBrowser,
		ListenTimeout: timeout,
		Output:        os.Stdout,
	})
	if err != nil {
		return err
	}

	fmt.Fprintln(os.Stdout)
	fmt.Fprintln(os.Stdout, "Lucidchart access token:")
	fmt.Fprintln(os.Stdout, token.AccessToken)
	if token.RefreshToken != "" {
		fmt.Fprintln(os.Stdout)
		fmt.Fprintln(os.Stdout, "Lucidchart refresh token:")
		fmt.Fprintln(os.Stdout, token.RefreshToken)
	}

	return nil
}

func lucidDocument(args []string) error {
	fs := flag.NewFlagSet("lucid document", flag.ContinueOnError)
	fs.SetOutput(os.Stdout)

	var documentID string
	fs.StringVar(&documentID, "id", "", "Lucidchart document ID")

	if err := fs.Parse(args); err != nil {
		if errors.Is(err, flag.ErrHelp) {
			return nil
		}
		return err
	}
	if documentID == "" {
		return fmt.Errorf("missing --id (document ID)")
	}

	s, err := settings.ToEnv()
	if err != nil {
		return fmt.Errorf("load settings: %w", err)
	}
	apiKey := strings.TrimSpace(s["LUCIDCHART_API_KEY"])
	if apiKey == "" {
		return fmt.Errorf("LUCIDCHART_API_KEY not set in settings (use monadscli settings set LUCIDCHART_API_KEY=...)")
	}

	body, err := lucid.GetDocument(context.Background(), documentID, apiKey)
	if err != nil {
		return err
	}

	fmt.Fprintln(os.Stdout, string(body))
	return nil
}

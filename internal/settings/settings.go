package settings

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"

	"github.com/ryanmontgomery/MonadsCLI/types"
)

const (
	settingsFileName = "monadscli.settings.enc"
	settingsKeyName  = "monadscli.settings.key"
	settingsKeyEnv   = "MONADSCLI_SETTINGS_KEY"
)

// Default values for behavior settings (must match readme/settings.md).
var defaultValues = map[string]string{
	"DEFAULT_CLI":          "CURSOR",
	"DEFAULT_TIMEOUT":      "600",
	"DEFAULT_RETRY_CLI":    "CURSOR",
	"DEFAULT_RETRY_COUNT":  "3",
	"DEFAULT_VALIDATE_CLI": "CURSOR",
	"LOG_DIR":              "./_monad_logs/",
	"WRITE_LOG_SHORT":      "true",
	"WRITE_LOG_LONG":       "true",
}

// Settings defines a map of env keys to values.
type Settings map[string]string

// Get returns settings formatted as .env content read from the encrypted file.
// Stored values are merged with defaultValues so unset behavior keys appear in the output.
func Get() ([]byte, error) {
	settings, err := loadSettings()
	if err != nil {
		return nil, err
	}
	effective := applyDefaults(settings)
	return []byte(formatEnv(effective)), nil
}

// Set writes the provided settings to the encrypted settings file.
func Set(settings Settings) error {
	if settings == nil {
		settings = Settings{}
	}

	payload, err := json.Marshal(settings)
	if err != nil {
		return err
	}

	return writeEncrypted(payload)
}

// FromEnv retrieves settings from environment variables and writes them to file.
func FromEnv() (Settings, error) {
	settings := Settings{}
	for _, key := range settingsKeys() {
		if value, ok := os.LookupEnv(key); ok {
			settings[key] = value
		}
	}

	if err := Set(settings); err != nil {
		return nil, err
	}

	return settings, nil
}

// FromFile reads .env-style settings from a file and writes them to file.
func FromFile(path string) (Settings, error) {
	parsed, err := parseEnvFile(path)
	if err != nil {
		return nil, err
	}

	filtered := Settings{}
	allowed := settingsKeySet()
	for key, value := range parsed {
		if _, ok := allowed[key]; ok {
			filtered[key] = value
		}
	}

	if err := Set(filtered); err != nil {
		return nil, err
	}

	return filtered, nil
}

// FromJSON reads JSON settings from a string and writes them to file.
func FromJSON(payload string) (Settings, error) {
	var settings Settings
	if err := json.Unmarshal([]byte(payload), &settings); err != nil {
		return nil, err
	}

	if err := Set(settings); err != nil {
		return nil, err
	}

	if settings == nil {
		settings = Settings{}
	}

	return settings, nil
}

// ToEnv loads settings and applies them to the current environment.
// Stored values are merged with defaultValues so unset behavior keys get the documented default.
func ToEnv() (Settings, error) {
	settings, err := loadSettings()
	if err != nil {
		return nil, err
	}
	effective := applyDefaults(settings)
	for key, value := range effective {
		if err := os.Setenv(key, value); err != nil {
			return nil, err
		}
	}
	return effective, nil
}

// CLILoginStatus returns which supported CLIs have credentials configured.
// It uses only the encrypted settings file (loadSettings): a CLI is logged in
// if at least one of its KeyENV keys is present in settings and non-empty.
// The current process environment is not consulted.
func CLILoginStatus() (map[string]bool, error) {
	settings, err := loadSettings()
	if err != nil {
		return nil, err
	}
	out := make(map[string]bool)
	for _, cli := range types.AllCLIs {
		loggedIn := false
		if strings.TrimSpace(cli.KeyENV) != "" {
			for _, key := range strings.Split(cli.KeyENV, ",") {
				name := strings.TrimSpace(key)
				if name == "" {
					continue
				}
				if v := strings.TrimSpace(settings[name]); v != "" {
					loggedIn = true
					break
				}
			}
		}
		out[cli.Name] = loggedIn
	}
	return out, nil
}

// ToFile writes settings to a .env-style file at the given path.
// Stored values are merged with defaultValues so the file contains effective values.
func ToFile(path string) (Settings, error) {
	settings, err := loadSettings()
	if err != nil {
		return nil, err
	}
	effective := applyDefaults(settings)
	if err := writeEnvFile(path, effective); err != nil {
		return nil, err
	}
	return effective, nil
}

func loadSettings() (Settings, error) {
	payload, err := readEncrypted()
	if err != nil {
		return nil, err
	}
	if len(payload) == 0 {
		return Settings{}, nil
	}

	var settings Settings
	if err := json.Unmarshal(payload, &settings); err != nil {
		return nil, err
	}

	if settings == nil {
		settings = Settings{}
	}

	return settings, nil
}

func readEncrypted() ([]byte, error) {
	path, err := settingsPath()
	if err != nil {
		return nil, err
	}

	payload, err := os.ReadFile(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil, nil
		}
		return nil, err
	}

	return decrypt(payload)
}

func writeEncrypted(payload []byte) error {
	path, err := settingsPath()
	if err != nil {
		return err
	}

	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}

	ciphertext, err := encrypt(payload)
	if err != nil {
		return err
	}

	return os.WriteFile(path, ciphertext, 0o600)
}

func settingsPath() (string, error) {
	dir, err := settingsDir()
	if err != nil {
		return "", err
	}

	return filepath.Join(dir, settingsFileName), nil
}

func settingsKeyPath() (string, error) {
	dir, err := settingsDir()
	if err != nil {
		return "", err
	}

	return filepath.Join(dir, settingsKeyName), nil
}

func settingsDir() (string, error) {
	home, homeErr := os.UserHomeDir()
	if homeErr != nil {
		return "", homeErr
	}

	exe, err := os.Executable()
	if err == nil && exe != "" {
		dir := filepath.Dir(exe)
		// Use stable home dir when running via "go run" (temp path varies per run)
		if !strings.Contains(dir, "go-build") {
			return dir, nil
		}
	}

	return filepath.Join(home, ".config", "monadscli"), nil
}

func encrypt(plain []byte) ([]byte, error) {
	key, err := settingsKey()
	if err != nil {
		return nil, err
	}

	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return nil, err
	}

	ciphertext := gcm.Seal(nil, nonce, plain, nil)
	return append(nonce, ciphertext...), nil
}

func decrypt(payload []byte) ([]byte, error) {
	if len(payload) == 0 {
		return nil, nil
	}

	key, err := settingsKey()
	if err != nil {
		return nil, err
	}

	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	nonceSize := gcm.NonceSize()
	if len(payload) < nonceSize {
		return nil, fmt.Errorf("settings payload is too short")
	}

	nonce := payload[:nonceSize]
	ciphertext := payload[nonceSize:]
	return gcm.Open(nil, nonce, ciphertext, nil)
}

func settingsKey() ([]byte, error) {
	if value := strings.TrimSpace(os.Getenv(settingsKeyEnv)); value != "" {
		return deriveKey(value), nil
	}

	path, err := settingsKeyPath()
	if err != nil {
		return nil, err
	}

	contents, err := os.ReadFile(path)
	if err == nil {
		return deriveKey(strings.TrimSpace(string(contents))), nil
	}
	if !errors.Is(err, os.ErrNotExist) {
		return nil, err
	}

	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return nil, err
	}

	raw := make([]byte, 32)
	if _, err := io.ReadFull(rand.Reader, raw); err != nil {
		return nil, err
	}

	encoded := base64.StdEncoding.EncodeToString(raw)
	if err := os.WriteFile(path, []byte(encoded), 0o600); err != nil {
		return nil, err
	}

	return raw, nil
}

func deriveKey(value string) []byte {
	if decoded, err := base64.StdEncoding.DecodeString(value); err == nil {
		if len(decoded) == 16 || len(decoded) == 24 || len(decoded) == 32 {
			return decoded
		}
	}

	sum := sha256.Sum256([]byte(value))
	return sum[:]
}

func settingsKeys() []string {
	keys := make([]string, 0, len(types.AllCLIs))
	seen := map[string]struct{}{}
	for _, key := range extraSettingsKeys {
		name := strings.TrimSpace(key)
		if name == "" {
			continue
		}
		if _, ok := seen[name]; ok {
			continue
		}
		seen[name] = struct{}{}
		keys = append(keys, name)
	}
	for _, cli := range types.AllCLIs {
		if strings.TrimSpace(cli.KeyENV) == "" {
			continue
		}
		for _, key := range strings.Split(cli.KeyENV, ",") {
			name := strings.TrimSpace(key)
			if name == "" {
				continue
			}
			if _, ok := seen[name]; ok {
				continue
			}
			seen[name] = struct{}{}
			keys = append(keys, name)
		}
	}

	sort.Strings(keys)
	return keys
}

var extraSettingsKeys = []string{
	"DEFAULT_CLI",
	"DEFAULT_TIMEOUT",
	"DEFAULT_RETRY_CLI",
	"DEFAULT_RETRY_COUNT",
	"DEFAULT_VALIDATE_CLI",
	"LOG_DIR",
	"WRITE_LOG_SHORT",
	"WRITE_LOG_LONG",
	"LUCIDCHART_API_KEY",
	"LUCID_OAUTH_CLIENT_ID",
	"LUCID_OAUTH_CLIENT_SECRET",
	"LUCID_OAUTH_REDIRECT_URL",
	"LUCID_OAUTH_SCOPES",
	"LUCID_OAUTH_PROMPT",
	"LUCID_OAUTH_AUTH_URL",
	"LUCID_OAUTH_TOKEN_URL",
	"LUCID_OAUTH_STATE",
}

// applyDefaults returns a copy of settings with defaultValues filled in for empty keys.
func applyDefaults(settings Settings) Settings {
	out := make(Settings, len(settings)+len(defaultValues))
	for k, v := range settings {
		out[k] = v
	}
	for k, v := range defaultValues {
		if strings.TrimSpace(out[k]) == "" {
			out[k] = v
		}
	}
	return out
}

// DefaultFor returns the documented default for a key, or "" if the key has no default.
func DefaultFor(key string) string {
	return defaultValues[key]
}

func settingsKeySet() map[string]struct{} {
	keys := settingsKeys()
	set := make(map[string]struct{}, len(keys))
	for _, key := range keys {
		set[key] = struct{}{}
	}
	return set
}

func parseEnvFile(path string) (map[string]string, error) {
	payload, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	lines := strings.Split(string(payload), "\n")
	values := map[string]string{}
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" || strings.HasPrefix(trimmed, "#") {
			continue
		}

		trimmed = strings.TrimSpace(strings.TrimPrefix(trimmed, "export "))
		parts := strings.SplitN(trimmed, "=", 2)
		if len(parts) != 2 {
			continue
		}

		key := strings.TrimSpace(parts[0])
		if key == "" {
			continue
		}

		value := strings.TrimSpace(parts[1])
		values[key] = parseEnvValue(value)
	}

	return values, nil
}

func parseEnvValue(value string) string {
	if value == "" {
		return ""
	}

	if len(value) >= 2 {
		if (value[0] == '"' && value[len(value)-1] == '"') || (value[0] == '\'' && value[len(value)-1] == '\'') {
			if unquoted, err := strconv.Unquote(value); err == nil {
				return unquoted
			}
		}
	}

	return value
}

func writeEnvFile(path string, settings Settings) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}

	payload := formatEnv(settings)
	return os.WriteFile(path, []byte(payload), 0o644)
}

func formatEnvValue(value string) string {
	if value == "" {
		return ""
	}

	needsQuote := strings.ContainsAny(value, " \t\r\n#\"")
	if !needsQuote {
		return value
	}

	return strconv.Quote(value)
}

func formatEnv(settings Settings) string {
	if len(settings) == 0 {
		return ""
	}

	keys := make([]string, 0, len(settings))
	for key := range settings {
		keys = append(keys, key)
	}
	sort.Strings(keys)

	lines := make([]string, 0, len(keys))
	for _, key := range keys {
		lines = append(lines, fmt.Sprintf("%s=%s", key, formatEnvValue(settings[key])))
	}

	return strings.Join(lines, "\n")
}

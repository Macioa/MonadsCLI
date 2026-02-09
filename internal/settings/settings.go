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

// Settings defines a map of env keys to values.
type Settings map[string]string

// Get returns settings formatted as .env content read from the encrypted file.
func Get() ([]byte, error) {
	settings, err := loadSettings()
	if err != nil {
		return nil, err
	}

	return []byte(formatEnv(settings)), nil
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
func ToEnv() (Settings, error) {
	settings, err := loadSettings()
	if err != nil {
		return nil, err
	}

	for key, value := range settings {
		if err := os.Setenv(key, value); err != nil {
			return nil, err
		}
	}

	return settings, nil
}

// ToFile writes settings to a .env-style file at the given path.
func ToFile(path string) (Settings, error) {
	settings, err := loadSettings()
	if err != nil {
		return nil, err
	}

	if err := writeEnvFile(path, settings); err != nil {
		return nil, err
	}

	return settings, nil
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
	exe, err := os.Executable()
	if err == nil && exe != "" {
		return filepath.Dir(exe), nil
	}

	home, homeErr := os.UserHomeDir()
	if homeErr != nil {
		if err != nil {
			return "", err
		}
		return "", homeErr
	}

	return filepath.Join(home, "bin"), nil
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
	"LUCIDCHART_API_KEY",
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

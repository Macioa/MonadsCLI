package settings

import (
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

func TestFromJSONAndGet(t *testing.T) {
	t.Setenv(settingsKeyEnv, "test-key")
	clearSettingsFiles(t)

	input := `{"GEMINI_API_KEY":"abc","CURSOR_API_KEY":"def"}`
	settings, err := FromJSON(input)
	if err != nil {
		t.Fatalf("FromJSON: %v", err)
	}

	expected := Settings{
		"GEMINI_API_KEY": "abc",
		"CURSOR_API_KEY": "def",
	}
	if !reflect.DeepEqual(settings, expected) {
		t.Fatalf("FromJSON settings mismatch: %#v", settings)
	}

	payload, err := Get()
	if err != nil {
		t.Fatalf("Get: %v", err)
	}
	expectedOut := "CURSOR_API_KEY=def\nGEMINI_API_KEY=abc"
	if string(payload) != expectedOut {
		t.Fatalf("Get output mismatch: %q", string(payload))
	}

	if err := os.Unsetenv("GEMINI_API_KEY"); err != nil {
		t.Fatalf("unset env: %v", err)
	}
	if err := os.Unsetenv("CURSOR_API_KEY"); err != nil {
		t.Fatalf("unset env: %v", err)
	}

	if _, err := ToEnv(); err != nil {
		t.Fatalf("ToEnv: %v", err)
	}

	if got := os.Getenv("GEMINI_API_KEY"); got != "abc" {
		t.Fatalf("GEMINI_API_KEY mismatch: %q", got)
	}
	if got := os.Getenv("CURSOR_API_KEY"); got != "def" {
		t.Fatalf("CURSOR_API_KEY mismatch: %q", got)
	}
}

func TestFromEnvRoundTrip(t *testing.T) {
	t.Setenv(settingsKeyEnv, "test-key")
	clearSettingsFiles(t)

	t.Setenv("GEMINI_API_KEY", "env-abc")
	t.Setenv("CURSOR_API_KEY", "env-def")
	t.Setenv("SHOULD_SKIP", "skip")

	settings, err := FromEnv()
	if err != nil {
		t.Fatalf("FromEnv: %v", err)
	}

	expected := Settings{
		"GEMINI_API_KEY": "env-abc",
		"CURSOR_API_KEY": "env-def",
	}
	if !reflect.DeepEqual(settings, expected) {
		t.Fatalf("FromEnv settings mismatch: %#v", settings)
	}

	if err := os.Unsetenv("GEMINI_API_KEY"); err != nil {
		t.Fatalf("unset env: %v", err)
	}
	if err := os.Unsetenv("CURSOR_API_KEY"); err != nil {
		t.Fatalf("unset env: %v", err)
	}

	if _, err := ToEnv(); err != nil {
		t.Fatalf("ToEnv: %v", err)
	}

	if got := os.Getenv("GEMINI_API_KEY"); got != "env-abc" {
		t.Fatalf("GEMINI_API_KEY mismatch: %q", got)
	}
	if got := os.Getenv("CURSOR_API_KEY"); got != "env-def" {
		t.Fatalf("CURSOR_API_KEY mismatch: %q", got)
	}
}

func TestFromFileToFile(t *testing.T) {
	t.Setenv(settingsKeyEnv, "test-key")
	clearSettingsFiles(t)

	dir := t.TempDir()
	inputPath := filepath.Join(dir, "input.env")
	outputPath := filepath.Join(dir, "output.env")

	payload := `# comment
export GEMINI_API_KEY="abc 123"
CURSOR_API_KEY=def
SHOULD_SKIP=skip
`

	if err := os.WriteFile(inputPath, []byte(payload), 0o600); err != nil {
		t.Fatalf("write input: %v", err)
	}

	settings, err := FromFile(inputPath)
	if err != nil {
		t.Fatalf("FromFile: %v", err)
	}

	expected := Settings{
		"GEMINI_API_KEY": "abc 123",
		"CURSOR_API_KEY": "def",
	}
	if !reflect.DeepEqual(settings, expected) {
		t.Fatalf("FromFile settings mismatch: %#v", settings)
	}

	if _, err := ToFile(outputPath); err != nil {
		t.Fatalf("ToFile: %v", err)
	}

	output, err := os.ReadFile(outputPath)
	if err != nil {
		t.Fatalf("read output: %v", err)
	}

	expectedOut := "CURSOR_API_KEY=def\nGEMINI_API_KEY=\"abc 123\""
	if string(output) != expectedOut {
		t.Fatalf("output mismatch: %q", string(output))
	}
}

func clearSettingsFiles(t *testing.T) {
	t.Helper()

	path, err := settingsPath()
	if err != nil {
		t.Fatalf("settings path: %v", err)
	}
	_ = os.Remove(path)

	keyPath, err := settingsKeyPath()
	if err != nil {
		t.Fatalf("settings key path: %v", err)
	}
	_ = os.Remove(keyPath)
}

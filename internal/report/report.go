package report

import (
	"encoding/json"
	"os"
	"path/filepath"

	"github.com/ryanmontgomery/MonadsCLI/internal/runner"
)

func WriteJSON(path string, result runner.Result) error {
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}

	data, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(path, data, 0o644)
}

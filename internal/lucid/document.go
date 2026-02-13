package lucid

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
)

const documentsBaseURL = "https://api.lucid.co/documents"

// GetDocument fetches the JSON contents of a Lucidchart document using the given API key.
// The API key must have the DocumentReadonly grant.
// See https://lucid.readme.io/reference/getdocumentcontent
func GetDocument(ctx context.Context, documentID string, apiKey string) ([]byte, error) {
	if documentID == "" {
		return nil, fmt.Errorf("document ID is required")
	}
	if apiKey == "" {
		return nil, fmt.Errorf("API key is required (set LUCIDCHART_API_KEY in settings)")
	}

	url := documentsBaseURL + "/" + documentID + "/contents"
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+apiKey)
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Lucid-Api-Version", "1")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("lucid API %s: %s", resp.Status, bytes.TrimSpace(body))
	}

	return body, nil
}

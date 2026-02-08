package types

// CLI defines metadata for a command line interface integration.
type CLI struct {
	Name    string `json:"name"`
	KeyURL  string `json:"keyUrl"`
	Command string `json:"command"`
	Prompt  string `json:"prompt"`
	Install string `json:"install"`
}

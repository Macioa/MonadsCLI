package types

// CLI defines metadata for a command line interface integration.
type CLI struct {
	Name    string `json:"name"`
	KeyURL  string `json:"keyUrl"`
	KeyENV  string `json:"keyEnv"`
	Command string `json:"command"`
	Prompt  string `json:"prompt"`
	Install string `json:"install"`
}

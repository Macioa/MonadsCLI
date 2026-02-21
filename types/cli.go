package types

// CLI defines metadata for a command line interface integration.
type CLI struct {
	Name     string `json:"name"`
	KeyURL   string `json:"keyUrl"`
	KeyENV   string `json:"keyEnv"`
	Codename string `json:"codename"` // Uppercase alias for settings (e.g. CURSOR); used when set, else Command
	Command  string `json:"command"`
	Prompt   string `json:"prompt"`
	Install  string `json:"install"`
}

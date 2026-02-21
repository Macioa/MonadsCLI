package types

var GeminiCLI = CLI{
	Name:     "Gemini CLI",
	KeyURL:   "https://aistudio.google.com/app/apikey",
	KeyENV:   "GEMINI_API_KEY",
	Codename: "GEMINI",
	Command:  "gemini",
	Prompt:   "gemini",
	Install:  "npm install -g @google/gemini-cli",
}

var CursorCLI = CLI{
	Name:     "Cursor CLI",
	KeyURL:   "https://cursor.com/dashboard",
	KeyENV:   "CURSOR_API_KEY",
	Codename: "CURSOR",
	Command:  "agent",
	Prompt:   "agent \"<prompt>\"",
	Install:  "curl https://cursor.com/install -fsS | bash",
}

var ClaudeCLI = CLI{
	Name:     "Claude CLI",
	KeyURL:   "https://console.anthropic.com/settings/keys",
	KeyENV:   "ANTHROPIC_API_KEY",
	Codename: "CLAUDE",
	Command:  "claude",
	Prompt:   "claude -p \"<prompt>\"",
	Install:  "npm install -g @anthropic-ai/claude-code",
}

var CopilotCLI = CLI{
	Name:     "GitHub Copilot CLI",
	KeyURL:   "https://github.com/settings/personal-access-tokens/new",
	KeyENV:   "GH_TOKEN",
	Codename: "COPILOT",
	Command:  "copilot",
	Prompt:   "copilot",
	Install:  "npm install -g @github/copilot",
}

var AiderCLI = CLI{
	Name:     "Aider",
	KeyURL:   "https://aider.chat/docs/llms.html",
	KeyENV:   "OPENAI_API_KEY, ANTHROPIC_API_KEY",
	Codename: "AIDER",
	Command:  "aider",
	Prompt:   "aider <file1> <file2>",
	Install:  "python -m pip install -U \"tree-sitter-yaml @ git+https://github.com/tree-sitter-grammars/tree-sitter-yaml.git@v0.7.1\" && python -m pip install -U aider-chat",
}

var QodoCLI = CLI{
	Name:     "Qodo Gen CLI",
	KeyURL:   "https://app.qodo.ai/",
	KeyENV:   "QODO_API_KEY",
	Codename: "QODO",
	Command:  "qodo",
	Prompt:   "qodo chat",
	Install:  "npm install -g @qodo/gen",
}

var AllCLIs = []CLI{
	GeminiCLI,
	CursorCLI,
	ClaudeCLI,
	CopilotCLI,
	AiderCLI,
	QodoCLI,
}

package types

var GeminiCLI = CLI{
	Name:    "Gemini CLI",
	KeyURL:  "https://aistudio.google.com/app/apikey",
	Command: "gemini",
	Prompt:  "gemini",
	Install: "npm install -g @google/gemini-cli",
}

var CursorCLI = CLI{
	Name:    "Cursor CLI",
	KeyURL:  "https://cursor.com/dashboard",
	Command: "agent",
	Prompt:  "agent \"<prompt>\"",
	Install: "curl https://cursor.com/install -fsS | bash",
}

var ClaudeCLI = CLI{
	Name:    "Claude CLI",
	KeyURL:  "https://console.anthropic.com/settings/keys",
	Command: "claude",
	Prompt:  "claude -p \"<prompt>\"",
	Install: "npm install -g @anthropic-ai/claude-code",
}

var CopilotCLI = CLI{
	Name:    "GitHub Copilot CLI",
	KeyURL:  "https://github.com/settings/personal-access-tokens/new",
	Command: "copilot",
	Prompt:  "copilot",
	Install: "npm install -g @github/copilot",
}

var AiderCLI = CLI{
	Name:    "Aider",
	KeyURL:  "https://aider.chat/docs/llms.html",
	Command: "aider",
	Prompt:  "aider <file1> <file2>",
	Install: "python -m pip install -U aider-chat",
}

var QodoCLI = CLI{
	Name:    "Qodo Gen CLI",
	KeyURL:  "https://app.qodo.ai/",
	Command: "qodo",
	Prompt:  "qodo chat",
	Install: "npm install -g @qodo/gen",
}

var AllCLIs = []CLI{
	GeminiCLI,
	CursorCLI,
	ClaudeCLI,
	CopilotCLI,
	AiderCLI,
	QodoCLI,
}

package types

var GeminiCLI = CLI{
	Name:                 "Gemini CLI",
	KeyURL:               "https://aistudio.google.com/app/apikey",
	KeyENV:               "GEMINI_API_KEY",
	Codename:             "GEMINI",
	Command:              "gemini",
	Prompt:               "gemini --yolo -p \"<prompt>\"",
	LinuxInstall:         "npm install -g @google/gemini-cli",
	WindowsInstallString: "npm install -g @google/gemini-cli; $env:Path = [System.Environment]::GetEnvironmentVariable('Path','Machine') + ';' + [System.Environment]::GetEnvironmentVariable('Path','User')",
}

var CursorCLI = CLI{
	Name:                 "Cursor CLI",
	KeyURL:               "https://cursor.com/dashboard",
	KeyENV:               "CURSOR_API_KEY",
	Codename:             "CURSOR",
	Command:              "agent",
	Prompt:               "agent -p --force \"<prompt>\"",
	LinuxInstall:         "curl https://cursor.com/install -fsS | bash",
	WindowsInstallString: "irm 'https://cursor.com/install?win32=true' | iex; $env:Path = [System.Environment]::GetEnvironmentVariable('Path','Machine') + ';' + [System.Environment]::GetEnvironmentVariable('Path','User'); $localBin = Join-Path (Join-Path $env:USERPROFILE '.local') 'bin'; if (Test-Path $localBin) { $env:Path = $localBin + ';' + $env:Path }",
}

var ClaudeCLI = CLI{
	Name:                 "Claude CLI",
	KeyURL:               "https://console.anthropic.com/settings/keys",
	KeyENV:               "ANTHROPIC_API_KEY",
	Codename:             "CLAUDE",
	Command:              "claude",
	Prompt:               "claude -p \"<prompt>\" --dangerously-skip-permissions",
	LinuxInstall:         "npm install -g @anthropic-ai/claude-code",
	WindowsInstallString: "npm install -g @anthropic-ai/claude-code; $env:Path = [System.Environment]::GetEnvironmentVariable('Path','Machine') + ';' + [System.Environment]::GetEnvironmentVariable('Path','User')",
}

var CopilotCLI = CLI{
	Name:                 "GitHub Copilot CLI",
	KeyURL:               "https://github.com/settings/personal-access-tokens/new",
	KeyENV:               "GH_TOKEN",
	Codename:             "COPILOT",
	Command:              "copilot",
	Prompt:               "copilot -p \"<prompt>\" --allow-all-tools",
	LinuxInstall:         "npm install -g @github/copilot",
	WindowsInstallString: "npm install -g @github/copilot; $env:Path = [System.Environment]::GetEnvironmentVariable('Path','Machine') + ';' + [System.Environment]::GetEnvironmentVariable('Path','User')",
}

var AiderCLI = CLI{
	Name:     "Aider",
	KeyURL:   "https://aider.chat/docs/llms.html",
	KeyENV:   "OPENAI_API_KEY, ANTHROPIC_API_KEY",
	Codename: "AIDER",
	Command:  "aider",
	Prompt:   "aider --yes -m \"<prompt>\"",
	LinuxInstall: "python -m pip install -U \"tree-sitter-yaml @ git+https://github.com/tree-sitter-grammars/tree-sitter-yaml.git@v0.7.1\" && python -m pip install -U aider-chat",
	WindowsInstallString: "python -m pip install -U \"tree-sitter-yaml @ git+https://github.com/tree-sitter-grammars/tree-sitter-yaml.git@v0.7.1\"; python -m pip install -U aider-chat; $env:Path = [System.Environment]::GetEnvironmentVariable('Path','Machine') + ';' + [System.Environment]::GetEnvironmentVariable('Path','User')",
}

var QodoCLI = CLI{
	Name:                 "Qodo Gen CLI",
	KeyURL:               "https://app.qodo.ai/",
	KeyENV:               "QODO_API_KEY",
	Codename:             "QODO",
	Command:              "qodo",
	Prompt:               "qodo -y --ci \"<prompt>\"",
	LinuxInstall:         "npm install -g @qodo/gen",
	WindowsInstallString: "npm install -g @qodo/gen; $env:Path = [System.Environment]::GetEnvironmentVariable('Path','Machine') + ';' + [System.Environment]::GetEnvironmentVariable('Path','User')",
}

var AllCLIs = []CLI{
	GeminiCLI,
	CursorCLI,
	ClaudeCLI,
	CopilotCLI,
	AiderCLI,
	QodoCLI,
}

package entry

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
)

// Launch executes the application
func (e *Entry) Launch() error {
	if err := e.Validate(); err != nil {
		return err
	}

	cmdString := parseExecString(e.Exec)
	envVars, cmdParts := parseExecWithEnv(cmdString)

	if len(cmdParts) == 0 {
		return ErrEmptyExec
	}

	// Create the command
	cmd := createCommand(cmdParts)
	cmd.Env = buildEnvironment(envVars)

	// If it's a terminal application, launch it in a terminal emulator
	if e.Terminal {
		termCmd, err := createTerminalCommand(envVars, cmdParts)
		if err != nil {
			return err
		}
		cmd = termCmd
	}

	// Start the process without waiting for it
	if err := cmd.Start(); err != nil {
		return fmt.Errorf("%w: %s: %v", ErrLaunchFailed, e.Name, err)
	}

	// Don't wait for the process to finish
	go cmd.Wait()

	return nil
}

// createCommand creates an exec.Cmd from command parts
func createCommand(cmdParts []string) *exec.Cmd {
	if len(cmdParts) == 1 {
		return exec.Command(cmdParts[0])
	}
	return exec.Command(cmdParts[0], cmdParts[1:]...)
}

// buildEnvironment builds the environment variables for the command
func buildEnvironment(envVars map[string]string) []string {
	env := os.Environ()
	for key, value := range envVars {
		env = append(env, fmt.Sprintf("%s=%s", key, value))
	}
	return env
}

// createTerminalCommand creates a command to run in a terminal emulator
func createTerminalCommand(envVars map[string]string, cmdParts []string) (*exec.Cmd, error) {
	terminal := findTerminal()
	if terminal == "" {
		return nil, ErrNoTerminal
	}

	fullCmd := reconstructCommand(envVars, cmdParts)

	// Different terminals have different syntax for executing commands
	var cmd *exec.Cmd
	switch terminal {
	case "gnome-terminal":
		cmd = exec.Command(terminal, "--", "sh", "-c", fullCmd)
	case "xterm":
		cmd = exec.Command(terminal, "-e", "sh", "-c", fullCmd)
	default:
		cmd = exec.Command(terminal, "-e", "sh", "-c", fullCmd)
	}

	cmd.Env = os.Environ()
	return cmd, nil
}

// findTerminal looks for an available terminal emulator
func findTerminal() string {
	terminals := []string{
		"alacritty",
		"kitty",
		"wezterm",
		"foot",
		"gnome-terminal",
		"konsole",
		"xfce4-terminal",
		"xterm",
	}

	for _, term := range terminals {
		if _, err := exec.LookPath(term); err == nil {
			return term
		}
	}

	return ""
}

// parseExecString handles desktop entry field codes
// See: https://specifications.freedesktop.org/desktop-entry-spec/latest/ar01s07.html
func parseExecString(exec string) string {
	// Remove field codes that we don't handle
	replacements := map[string]string{
		"%f": "", "%F": "",
		"%u": "", "%U": "",
		"%i": "", "%c": "", "%k": "",
		"%d": "", "%D": "", // deprecated
		"%n": "", "%N": "", // deprecated
		"%v": "", "%m": "", // deprecated
	}

	result := exec
	for old, new := range replacements {
		result = strings.ReplaceAll(result, old, new)
	}

	// Clean up extra spaces
	result = strings.TrimSpace(result)
	fields := strings.Fields(result)
	return strings.Join(fields, " ")
}

// parseExecWithEnv parses a command string into environment variables and command parts
func parseExecWithEnv(exec string) (map[string]string, []string) {
	envVars := make(map[string]string)
	var cmdParts []string

	parts := splitExecString(exec)

	for _, part := range parts {
		// Check if this part is an environment variable (KEY=VALUE)
		if idx := strings.Index(part, "="); idx > 0 && !strings.Contains(part[:idx], " ") {
			key := part[:idx]
			value := part[idx+1:]
			// Only treat as env var if key looks like a valid env var name
			if isValidEnvVarName(key) {
				envVars[key] = value
				continue
			}
		}
		// Not an env var, it's part of the command
		cmdParts = append(cmdParts, part)
	}

	return envVars, cmdParts
}

// isValidEnvVarName checks if a string is a valid environment variable name
func isValidEnvVarName(name string) bool {
	if len(name) == 0 {
		return false
	}

	// Env var names typically contain only uppercase letters, digits, and underscores
	// and don't start with a digit
	for i, ch := range name {
		if i == 0 && (ch >= '0' && ch <= '9') {
			return false
		}
		if !((ch >= 'A' && ch <= 'Z') || (ch >= 'a' && ch <= 'z') ||
			(ch >= '0' && ch <= '9') || ch == '_') {
			return false
		}
	}

	return true
}

// reconstructCommand rebuilds a command string from env vars and parts
func reconstructCommand(envVars map[string]string, cmdParts []string) string {
	var parts []string

	// Add env vars
	for key, value := range envVars {
		parts = append(parts, fmt.Sprintf("%s=%s", key, value))
	}

	// Add command parts
	parts = append(parts, cmdParts...)

	return strings.Join(parts, " ")
}

// splitExecString splits the exec string into command and arguments
// This handles quoted arguments properly
func splitExecString(exec string) []string {
	var parts []string
	var current strings.Builder
	inQuote := false
	quoteChar := rune(0)

	runes := []rune(exec)
	for i := 0; i < len(runes); i++ {
		char := runes[i]

		switch {
		case (char == '"' || char == '\'') && !inQuote:
			// Start of quote
			inQuote = true
			quoteChar = char

		case char == quoteChar && inQuote:
			// End of quote
			inQuote = false
			quoteChar = 0

		case char == ' ' && !inQuote:
			// Space outside quotes - separator
			if current.Len() > 0 {
				parts = append(parts, current.String())
				current.Reset()
			}

		case char == '\\' && i+1 < len(runes):
			// Escape sequence
			i++
			current.WriteRune(runes[i])

		default:
			current.WriteRune(char)
		}
	}

	// Add the last part
	if current.Len() > 0 {
		parts = append(parts, current.String())
	}

	return parts
}

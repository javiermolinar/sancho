package input

import "strings"

// PromptCommand describes a command suggestion entry.
type PromptCommand struct {
	Name        string
	Description string
}

// PromptMatchingCommands returns commands that match the current input prefix.
func PromptMatchingCommands(input string, commands []PromptCommand) []PromptCommand {
	if !strings.HasPrefix(strings.TrimSpace(input), "/") {
		return nil
	}
	if strings.Contains(input, " ") {
		return nil
	}

	prefix := strings.ToLower(strings.TrimSpace(input))
	matches := make([]PromptCommand, 0, len(commands))
	for _, cmd := range commands {
		if strings.HasPrefix(strings.ToLower(cmd.Name), prefix) {
			matches = append(matches, cmd)
		}
	}
	return matches
}

// PromptAutocomplete returns the first matching command and whether it exists.
func PromptAutocomplete(input string, commands []PromptCommand) (string, bool) {
	matches := PromptMatchingCommands(input, commands)
	if len(matches) == 0 {
		return "", false
	}
	return matches[0].Name + " ", true
}

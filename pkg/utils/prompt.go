package utils

import (
	"bufio"
	"fmt"
	"os"
	"strings"
	"syscall"

	"golang.org/x/term"
)

// Prompt handles user input prompts
type Prompt struct {
	reader *bufio.Reader
}

// NewPrompt creates a new prompt
func NewPrompt() *Prompt {
	return &Prompt{
		reader: bufio.NewReader(os.Stdin),
	}
}

// String prompts for a string input
func (p *Prompt) String(prompt string) (string, error) {
	fmt.Print(prompt)
	text, err := p.reader.ReadString('\n')
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(text), nil
}

// Password prompts for a password input (hidden)
func (p *Prompt) Password(prompt string) (string, error) {
	fmt.Print(prompt)
	bytePassword, err := term.ReadPassword(int(syscall.Stdin))
	if err != nil {
		return "", err
	}
	fmt.Println()
	return strings.TrimSpace(string(bytePassword)), nil
}

// Confirm prompts for a yes/no confirmation
func (p *Prompt) Confirm(prompt string) (bool, error) {
	response, err := p.String(prompt + " (y/n): ")
	if err != nil {
		return false, err
	}
	return strings.ToLower(response) == "y", nil
}

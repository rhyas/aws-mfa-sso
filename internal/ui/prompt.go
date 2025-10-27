package ui

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"

	"github.com/fatih/color"
)

// PromptUsername prompts the user for their username
func PromptUsername() string {
	color.Set(color.FgYellow)
	fmt.Print("Enter Username: ")
	color.Unset()
	scanner := bufio.NewScanner(os.Stdin)
	scanner.Scan()
	return scanner.Text()
}

// PromptPassword prompts the user for their password (hidden input)
func PromptPassword() string {
	color.Set(color.FgYellow)
	fmt.Print("Enter Password: ")
	color.Unset()
	return readPassword()
}

// PromptMFA prompts the user for their MFA code
func PromptMFA() string {
	color.Set(color.FgYellow)
	fmt.Print("Enter MFA: ")
	color.Unset()
	scanner := bufio.NewScanner(os.Stdin)
	scanner.Scan()
	return scanner.Text()
}

// readPassword reads a password from stdin without echoing it to the terminal
func readPassword() string {
	scanner := bufio.NewScanner(os.Stdin)
	// Disable echo
	exec.Command("stty", "-F", "/dev/tty", "-echo").Run()
	scanner.Scan()
	password := scanner.Text()
	// Re-enable echo
	exec.Command("stty", "-F", "/dev/tty", "echo").Run()
	fmt.Println() // Print newline after password input
	return password
}

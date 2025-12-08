package docker

import (
	"bufio"
	"fmt"
	"os"

	"syscall"

	"golang.org/x/term"
)

func PromptForCredentials() (string, string, error) {
	reader := bufio.NewReader(os.Stdin)

	fmt.Print("Enter Docker Hub username: ")
	username, err := reader.ReadString('\n')
	if err != nil {
		return "", "", err
	}
	username = username[:len(username)-1] // trim newline

	fmt.Print("Enter Docker Hub password: ")
	bytePassword, err := term.ReadPassword(int(syscall.Stdin))
	if err != nil {
		return "", "", err
	}
	fmt.Println() // new line after hidden input

	password := string(bytePassword)

	return username, password, nil
}

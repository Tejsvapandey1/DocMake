package docker

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
)

func Login(username, password string) error {
	if username == "" || password == "" {
		fmt.Println("⚠️  Docker Hub login skipped (credentials missing).")
		return nil
	}

	fmt.Println("🔐 Logging into Docker Hub...")

	cmd := exec.Command("docker", "login", "-u", username, "--password-stdin")
	cmd.Stdin = strings.NewReader(password)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	return cmd.Run()
}

func PushImage(imageName string) error {
	fmt.Println("📤 Pushing image:", imageName)

	cmd := exec.Command("docker", "push", imageName)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	return cmd.Run()
}

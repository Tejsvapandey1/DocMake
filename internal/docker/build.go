package docker

import (
	"fmt"
	"os"
	"os/exec"
)

func BuildImage(folderPath, imageName string) error {
	fmt.Println("🐳 Building Docker image:", imageName)

	cmd := exec.Command("docker", "build", "-t", imageName, "-f", "Dockerfile", ".")
	cmd.Dir = folderPath
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	return cmd.Run()
}

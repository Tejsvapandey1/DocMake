package git

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
)

func CloneRepo(repoURL string) (string, error) {

	// Extract folder name from URL
	parts := strings.Split(repoURL, "/")
	repoName := parts[len(parts)-1]
	repoName = strings.TrimSuffix(repoName, ".git")

	// Check if folder already exists
	_, err := os.Stat(repoName)
	if err == nil { // folder exists
		fmt.Println("⚠️  Repo already exists:", repoName)
		return repoName, nil
	}

	// Clone the repository
	fmt.Println("📥 Cloning repo:", repoName)

	cmd := exec.Command("git", "clone", repoURL)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	err = cmd.Run()
	if err != nil {
		return "", err
	}

	return repoName, nil
}

func RepoPath(repoURL string) string {
	parts := strings.Split(repoURL, "/")
	repoName := parts[len(parts)-1]
	return strings.TrimSuffix(repoName, ".git")
}

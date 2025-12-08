package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/spf13/cobra"
	"github.com/tejsvapandey1/docmake/internal/detect"
	"github.com/tejsvapandey1/docmake/internal/docker"
	"github.com/tejsvapandey1/docmake/internal/generator"
	"github.com/tejsvapandey1/docmake/internal/git"
)

var dockerHubUser string
var dockerHubPass string

var cloneCmd = &cobra.Command{
	Use:   "clone <repo-url>",
	Short: "Clone a repository",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		repoURL := args[0]

		// 1. Clone repo
		folderPath, err := git.CloneRepo(repoURL)
		if err != nil {
			fmt.Println("Error cloning:", err)
			return
		}
		fmt.Println("Repository cloned at:", folderPath)

		// 2. Detect tech stack
		stack, err := detect.DetectStack(folderPath)
		if err != nil {
			fmt.Println("Error detecting stack:", err)
			return
		}

		fmt.Println("Tech stack detected:")
		fmt.Println("Language:", stack.Primary)
		if stack.Framework != "" {
			fmt.Println("Framework:", stack.Framework)
		}

		// 3. Detect project meta
		var meta detect.ProjectMeta
		switch stack.Primary {
		case "node":
			meta = detect.DetectNodeDetails(folderPath)
		case "python":
			meta = detect.DetectPythonDetails(folderPath)
		case "go":
			meta.EntryFile = "main.go"
			meta.Port = "8080"
		}

		fmt.Println("Entry File:", meta.EntryFile)
		fmt.Println("Port:", meta.Port)
		if meta.Framework != "" {
			fmt.Println("Framework:", meta.Framework)
		}

		// 4. Detect .env
		envMap, envPath := detect.DetectEnv(folderPath)
		meta.Env = envMap
		meta.EnvFilePath = envPath

		if len(meta.Env) > 0 {
			fmt.Println("Detected .env keys:")
			for k := range meta.Env {
				fmt.Println(" -", k)
			}
		}

		// 5. Detect Database
		meta.Database = detect.DetectDatabase(folderPath)
		if meta.Database.Type != "" {
			fmt.Println("Detected Database:", meta.Database.Type)
		}

		// 6. Get Docker Hub credentials (from flag or env)
		if dockerHubUser == "" {
			dockerHubUser = os.Getenv("DOCKERHUB_USERNAME")
		}
		if dockerHubPass == "" {
			dockerHubPass = os.Getenv("DOCKERHUB_PASSWORD")
		}

		// If missing, ask interactively
		if dockerHubUser == "" || dockerHubPass == "" {
			fmt.Println("🔐 Docker Hub credentials required.")
			dockerHubUser, dockerHubPass, err = docker.PromptForCredentials()
			if err != nil {
				fmt.Println("Error reading credentials:", err)
				return
			}
		}

		// 7. Create final image name
		imageName := fmt.Sprintf("%s/%s:latest", dockerHubUser, filepath.Base(folderPath))

		// 8. Generate Dockerfile
		err = generator.GenerateDockerfile(folderPath, stack, meta)
		if err != nil {
			fmt.Println("Error creating Dockerfile:", err)
			return
		}
		fmt.Println("📦 Dockerfile generated successfully!")

		// 9. Generate docker-compose.yml (using final image name)
		err = generator.GenerateComposeFile(folderPath, stack, meta, imageName)
		if err != nil {
			fmt.Println("Error generating docker-compose file:", err)
			return
		}
		fmt.Println("📦 docker-compose.yml generated successfully!")

		// 10. Build image
		err = docker.BuildImage(folderPath, imageName)
		if err != nil {
			fmt.Println("❌ Error building Docker image:", err)
			return
		}
		fmt.Println("🐳 Docker image built:", imageName)

		// 11. Docker Hub login
		err = docker.Login(dockerHubUser, dockerHubPass)
		if err != nil {
			fmt.Println("❌ Docker Hub login failed:", err)
			return
		}

		// 12. Push image
		err = docker.PushImage(imageName)
		if err != nil {
			fmt.Println("❌ Docker image push failed:", err)
			return
		}

		fmt.Println("📤 Docker image pushed successfully:", imageName)

		// 13. Auto-start the project locally
		fmt.Println("🚀 Starting project locally using docker compose...")
		composeCmd := exec.Command("docker", "compose", "up", "-d")
		composeCmd.Dir = folderPath
		composeCmd.Stdout = os.Stdout
		composeCmd.Stderr = os.Stderr

		err = composeCmd.Run()
		if err != nil {
			fmt.Println("❌ Failed to start project:", err)
			return
		}

		fmt.Println("🎉 Project is now running locally!")
	},
}

func init() {
	rootCmd.AddCommand(cloneCmd)

	cloneCmd.Flags().StringVar(&dockerHubUser, "hub-user", "", "Docker Hub username")
	cloneCmd.Flags().StringVar(&dockerHubPass, "hub-pass", "", "Docker Hub password or token")
}

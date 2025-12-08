package cmd

import (
	"os"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "docmake",
	Short: "docmake: Auto-dockerize and auto-deploy your repo",
	Long:  "docmake automatically clones, detects stack, builds Dockerfiles, pushes images, and generates docker-compose files.",
}

func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

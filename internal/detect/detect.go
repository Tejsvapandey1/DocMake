package detect

import (
	"encoding/json"
	"os"
	"path/filepath"
)

type TechStack struct {
	Primary   string
	Framework string
}

// DetectStack scans a repo folder and returns the detected tech stack
func DetectStack(repoPath string) (*TechStack, error) {
	tech := &TechStack{}

	// Check files in the root directory
	entries, err := os.ReadDir(repoPath)
	if err != nil {
		return nil, err
	}

	for _, entry := range entries {
		name := entry.Name()

		// ==== GO ====
		if name == "go.mod" {
			tech.Primary = "go"
			return tech, nil
		}

		// ==== PYTHON ====
		if name == "requirements.txt" {
			tech.Primary = "python"
			return tech, nil
		}

		// ==== JAVASCRIPT / NODE ====
		if name == "package.json" {
			jsType := detectNode(repoPath)
			return jsType, nil
		}

		// ==== JAVA ====
		if name == "pom.xml" {
			tech.Primary = "java-maven"
			return tech, nil
		}
		if name == "build.gradle" {
			tech.Primary = "java-gradle"
			return tech, nil
		}
	}

	// ==== FALLBACK BY EXTENSIONS ====
	var pyCount, goCount, jsCount int

	filepath.Walk(repoPath, func(path string, info os.FileInfo, err error) error {
		if info.IsDir() {
			return nil
		}

		switch filepath.Ext(path) {
		case ".py":
			pyCount++
		case ".go":
			goCount++
		case ".js", ".jsx":
			jsCount++
		}

		return nil
	})

	// Pick most common
	if pyCount > goCount && pyCount > jsCount {
		tech.Primary = "python"
	} else if goCount > pyCount && goCount > jsCount {
		tech.Primary = "go"
	} else if jsCount > pyCount && jsCount > goCount {
		tech.Primary = "javascript"
	} else {
		tech.Primary = "unknown"
	}

	return tech, nil
}

func detectNode(repoPath string) *TechStack {
	data, err := os.ReadFile(filepath.Join(repoPath, "package.json"))
	if err != nil {
		return &TechStack{Primary: "node"}
	}

	var pkg map[string]interface{}
	json.Unmarshal(data, &pkg)

	deps := map[string]interface{}{}
	if d, ok := pkg["dependencies"].(map[string]interface{}); ok {
		deps = d
	}

	if _, ok := deps["next"]; ok {
		return &TechStack{Primary: "node", Framework: "nextjs"}
	}
	if _, ok := deps["react"]; ok {
		return &TechStack{Primary: "node", Framework: "react"}
	}

	return &TechStack{Primary: "node"}
}

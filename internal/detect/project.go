package detect

import (
	"bufio"
	"encoding/json"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

type ProjectMeta struct {
	EntryFile string
	Port      string
	Framework string
	Database  DatabaseInfo

	Env map[string]string

	EnvFilePath string
}

type DatabaseInfo struct {
	Type       string
	Port       string
	EnvVar     string
	DefaultURI string
}

func DetectNodeDetails(path string) ProjectMeta {
	meta := ProjectMeta{}

	// Read package.json
	pkgPath := filepath.Join(path, "package.json")
	data, err := os.ReadFile(pkgPath)
	if err == nil {
		var pkg map[string]interface{}
		json.Unmarshal(data, &pkg)

		// Detect entry file from "main"
		if mainFile, ok := pkg["main"].(string); ok {
			meta.EntryFile = mainFile
		}

		// Detect framework from dependencies
		if deps, ok := pkg["dependencies"].(map[string]interface{}); ok {
			if _, ok := deps["express"]; ok {
				meta.Framework = "express"
			}
			if _, ok := deps["next"]; ok {
				meta.Framework = "nextjs"
			}
			if _, ok := deps["react"]; ok {
				meta.Framework = "react"
			}
			if _, ok := deps["nest"]; ok {
				meta.Framework = "nestjs"
			}
		}
	}

	// Fallback entry file guesses
	candidates := []string{"server.js", "app.js", "index.js"}
	for _, c := range candidates {
		if fileExists(filepath.Join(path, c)) {
			meta.EntryFile = c
			break
		}
	}

	// Detect port usage by scanning .js files
	filepath.Walk(path, func(p string, info os.FileInfo, err error) error {
		if strings.HasSuffix(p, ".js") {
			content, _ := os.ReadFile(p)

			re := regexp.MustCompile(`listen\((\d+)\)`)
			m := re.FindStringSubmatch(string(content))
			if len(m) > 1 {
				meta.Port = m[1]
			}
		}
		return nil
	})

	if meta.Port == "" {
		meta.Port = "3000" // default
	}

	return meta
}

func DetectPythonDetails(path string) ProjectMeta {
	meta := ProjectMeta{}

	// Detect common entry files
	candidates := []string{"app.py", "main.py", "run.py"}
	for _, c := range candidates {
		if fileExists(filepath.Join(path, c)) {
			meta.EntryFile = c
			break
		}
	}

	// Detect Django
	if fileExists(filepath.Join(path, "manage.py")) {
		meta.Framework = "django"
		meta.EntryFile = "manage.py"
		meta.Port = "8000"
		return meta
	}

	// Scan python files for Flask / FastAPI
	filepath.Walk(path, func(p string, info os.FileInfo, err error) error {
		if strings.HasSuffix(p, ".py") {
			content, _ := os.ReadFile(p)

			if strings.Contains(string(content), "from flask") {
				meta.Framework = "flask"
			}
			if strings.Contains(string(content), "fastapi import") {
				meta.Framework = "fastapi"
			}

			// detect flask port
			re := regexp.MustCompile(`run\(.*port\s*=\s*(\d+)`)
			m := re.FindStringSubmatch(string(content))
			if len(m) > 1 {
				meta.Port = m[1]
			}
		}
		return nil
	})

	if meta.Port == "" {
		if meta.Framework == "flask" {
			meta.Port = "5000"
		} else {
			meta.Port = "8000"
		}
	}

	return meta
}

func DetectDatabase(path string) DatabaseInfo {
	db := DatabaseInfo{}

	// Check package.json / requirements.txt
	checkFileForDB(filepath.Join(path, "package.json"), &db)
	checkFileForDB(filepath.Join(path, "requirements.txt"), &db)

	// Scan all code files
	filepath.Walk(path, func(p string, info os.FileInfo, err error) error {
		if strings.HasSuffix(p, ".js") || strings.HasSuffix(p, ".py") || strings.HasSuffix(p, ".go") {
			content, _ := os.ReadFile(p)
			detectDBFromContent(string(content), &db)
		}
		return nil
	})

	// Detect from .env
	envPath := filepath.Join(path, ".env")
	if fileExists(envPath) {
		content, _ := os.ReadFile(envPath)
		detectDBFromContent(string(content), &db)
	}

	return db
}

func detectDBFromContent(content string, db *DatabaseInfo) {
	text := strings.ToLower(content)

	switch {
	case strings.Contains(text, "mongoose") || strings.Contains(text, "pymongo"):
		db.Type = "mongo"
		db.Port = "27017"
		db.EnvVar = "MONGO_URI"
		db.DefaultURI = "mongodb://mongo:27017"

	case strings.Contains(text, "postgres") || strings.Contains(text, "pgx") ||
		strings.Contains(text, "gorm.io/driver/postgres") ||
		strings.Contains(text, "psycopg2"):
		db.Type = "postgres"
		db.Port = "5432"
		db.EnvVar = "DATABASE_URL"
		db.DefaultURI = "postgres://postgres:password@postgres:5432/postgres?sslmode=disable"

	case strings.Contains(text, "mysql") || strings.Contains(text, "mysql2"):
		db.Type = "mysql"
		db.Port = "3306"
		db.EnvVar = "MYSQL_URL"
		db.DefaultURI = "mysql://root:password@mysql:3306/db"

	case strings.Contains(text, "redis"):
		db.Type = "redis"
		db.Port = "6379"
		db.EnvVar = "REDIS_URL"
		db.DefaultURI = "redis://redis:6379"
	}
}

func checkFileForDB(path string, db *DatabaseInfo) {
	data, err := os.ReadFile(path)
	if err != nil {
		return
	}

	detectDBFromContent(string(data), db)
}

func DetectEnv(path string) (map[string]string, string) {
	envMap := make(map[string]string)
	envPath := filepath.Join(path, ".env")

	// If no .env file, return empty map and empty path
	f, err := os.Open(envPath)
	if err != nil {
		return envMap, ""
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		// skip comments and empty lines
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		// key=value (allow = in value)
		parts := strings.SplitN(line, "=", 2)
		if len(parts) != 2 {
			continue
		}
		key := strings.TrimSpace(parts[0])
		val := strings.TrimSpace(parts[1])

		// remove surrounding quotes if present
		if len(val) >= 2 {
			if (strings.HasPrefix(val, `"`) && strings.HasSuffix(val, `"`)) ||
				(strings.HasPrefix(val, `'`) && strings.HasSuffix(val, `'`)) {
				val = val[1 : len(val)-1]
			}
		}
		envMap[key] = val
	}

	// if scanner error, just return whatever parsed
	return envMap, ".env"
}

func fileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

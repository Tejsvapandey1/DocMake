package generator

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/tejsvapandey1/docmake/internal/detect"
)

// ---------------------------------------------------
// SINGLE SERVICE COMPOSE GENERATOR
// ---------------------------------------------------

func GenerateComposeFile(path string, stack *detect.TechStack, meta detect.ProjectMeta, imageName string) error {
	filePath := filepath.Join(path, "docker-compose.yml")

	var appService string

	switch stack.Primary {
	case "node":
		appService = composeNode(meta, imageName)
	case "python":
		if meta.Framework == "django" {
			appService = composeDjango(meta, imageName)
		} else {
			appService = composePython(meta, imageName)
		}
	case "go":
		appService = composeGo(meta, imageName)
	default:
		return fmt.Errorf("unsupported tech stack: %s", stack.Primary)
	}

	// Add DB service
	completeCompose := composeWithDB(appService, meta.Database)

	return os.WriteFile(filePath, []byte(completeCompose), 0644)
}

// ---------------------------------------------------
// MULTI-SERVICE COMPOSE GENERATOR
// ---------------------------------------------------

func GenerateMultiCompose(path string, frontend detect.ProjectMeta, backend detect.ProjectMeta, fImage, bImage string) error {
	filePath := filepath.Join(path, "docker-compose.yml")

	compose := composeMultiService(frontend, backend, fImage, bImage)

	return os.WriteFile(filePath, []byte(compose), 0644)
}

// ---------------------------------------------------
// NODE.JS SERVICE
// ---------------------------------------------------

func composeNode(meta detect.ProjectMeta, imageName string) string {
	envBlock := buildEnvBlock(meta)

	envFileLine := ""
	if meta.EnvFilePath != "" {
		envFileLine = fmt.Sprintf("    env_file:\n      - %s\n", meta.EnvFilePath)
	}

	envSection := envFileLine
	if envFileLine == "" && envBlock != "" {
		envSection = fmt.Sprintf("    environment:\n%s", envBlock)
	}

	return fmt.Sprintf(`
version: '3.9'

services:
  app:
    image: %s
    container_name: node_app
    ports:
      - "%s:%s"
%s    command: ["node", "%s"]
`, imageName, meta.Port, meta.Port, envSection, meta.EntryFile)
}

// ---------------------------------------------------
// PYTHON SERVICE
// ---------------------------------------------------

func composePython(meta detect.ProjectMeta, imageName string) string {
	envBlock := buildEnvBlock(meta)

	envFileLine := ""
	if meta.EnvFilePath != "" {
		envFileLine = fmt.Sprintf("    env_file:\n      - %s\n", meta.EnvFilePath)
	}

	envSection := envFileLine
	if envFileLine == "" && envBlock != "" {
		envSection = fmt.Sprintf("    environment:\n%s", envBlock)
	}

	return fmt.Sprintf(`
version: '3.9'

services:
  app:
    image: %s
    container_name: python_app
    ports:
      - "%s:%s"
%s    command: ["python", "%s"]
`, imageName, meta.Port, meta.Port, envSection, meta.EntryFile)
}

// ---------------------------------------------------
// DJANGO SERVICE
// ---------------------------------------------------

func composeDjango(meta detect.ProjectMeta, imageName string) string {
	envBlock := buildEnvBlock(meta)

	envFileLine := ""
	if meta.EnvFilePath != "" {
		envFileLine = fmt.Sprintf("    env_file:\n      - %s\n", meta.EnvFilePath)
	}

	envSection := envFileLine
	if envFileLine == "" && envBlock != "" {
		envSection = fmt.Sprintf("    environment:\n%s", envBlock)
	}

	return fmt.Sprintf(`
version: '3.9'

services:
  app:
    image: %s
    container_name: django_app
    ports:
      - "8000:8000"
%s    command: >
      sh -c "python manage.py migrate &&
             gunicorn project.wsgi:application --bind 0.0.0.0:8000"
`, imageName, envSection)
}

// ---------------------------------------------------
// GO SERVICE
// ---------------------------------------------------

func composeGo(meta detect.ProjectMeta, imageName string) string {
	return fmt.Sprintf(`
version: '3.9'

services:
  app:
    image: %s
    container_name: go_app
    ports:
      - "8080:8080"
`, imageName)
}

// ---------------------------------------------------
// ENV BLOCK + DATABASE SUPPORT
// ---------------------------------------------------

func buildEnvBlock(meta detect.ProjectMeta) string {
	if len(meta.Env) == 0 {
		if meta.Database.Type != "" && meta.Database.DefaultURI != "" {
			envName := meta.Database.EnvVar
			if envName == "" {
				envName = "DATABASE_URL"
			}
			return fmt.Sprintf("      - %s=%s\n", envName, meta.Database.DefaultURI)
		}
		return ""
	}

	var sb strings.Builder
	for k, v := range meta.Env {
		escaped := strings.ReplaceAll(v, `"`, `'`)
		sb.WriteString(fmt.Sprintf("      - %s=%s\n", k, escaped))
	}

	if meta.Database.Type != "" && meta.Database.DefaultURI != "" {
		envName := meta.Database.EnvVar
		if envName == "" {
			envName = "DATABASE_URL"
		}
		if _, exists := meta.Env[envName]; !exists {
			sb.WriteString(fmt.Sprintf("      - %s=%s\n", envName, meta.Database.DefaultURI))
		}
	}

	return sb.String()
}

// ---------------------------------------------------
// DATABASE SERVICES
// ---------------------------------------------------

func composeWithDB(appService string, db detect.DatabaseInfo) string {
	if db.Type == "" {
		return appService
	}

	dbTemplate := ""

	switch db.Type {
	case "mongo":
		dbTemplate = `
  mongo:
    image: mongo
    ports:
      - "27017:27017"
    volumes:
      - mongo_data:/data/db
`
	case "postgres":
		dbTemplate = `
  postgres:
    image: postgres
    environment:
      POSTGRES_PASSWORD: password
    ports:
      - "5432:5432"
    volumes:
      - pg_data:/var/lib/postgresql/data
`
	case "mysql":
		dbTemplate = `
  mysql:
    image: mysql:8
    environment:
      MYSQL_ROOT_PASSWORD: password
    ports:
      - "3306:3306"
    volumes:
      - mysql_data:/var/lib/mysql
`
	case "redis":
		dbTemplate = `
  redis:
    image: redis
    ports:
      - "6379:6379"
`
	}

	volumeBlock := `
volumes:
  mongo_data:
  pg_data:
  mysql_data:
`

	return appService + dbTemplate + volumeBlock
}

// ---------------------------------------------------
// MULTI-SERVICE TEMPLATE
// ---------------------------------------------------

func composeMultiService(frontend, backend detect.ProjectMeta, fImage, bImage string) string {
	return fmt.Sprintf(`
version: '3.9'

services:

  backend:
    image: %s
    ports:
      - "%s:%s"

  frontend:
    image: %s
    ports:
      - "3000:3000"
    depends_on:
      - backend

  mongo:
    image: mongo
    ports:
      - "27017:27017"

volumes:
  db_data:
`, bImage, backend.Port, backend.Port, fImage)
}

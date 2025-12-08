package generator

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/tejsvapandey1/docmake/internal/detect"
)

func GenerateDockerfile(path string, stack *detect.TechStack, meta detect.ProjectMeta) error {
	filePath := filepath.Join(path, "Dockerfile")

	var content string

	switch stack.Primary {

	case "go":
		content = dockerfileGo(meta)

	case "python":
		content = dockerfilePython(meta)

	case "node":
		switch meta.Framework {
		case "nextjs":
			content = dockerfileNext()
		case "react":
			content = dockerfileReact()
		case "express", "nestjs", "":
			// Default node backend
			content = dockerfileNode(meta)
		}

	default:
		return fmt.Errorf("unsupported tech stack: %s", stack.Primary)
	}

	return os.WriteFile(filePath, []byte(content), 0644)
}

func dockerfileGo(meta detect.ProjectMeta) string {
	return `
FROM golang:1.22 AS builder
WORKDIR /app

COPY . .
RUN go build -o app .

FROM debian:bookworm-slim
WORKDIR /app

COPY --from=builder /app/app .

EXPOSE 8080
CMD ["./app"]
`
}

func dockerfilePython(meta detect.ProjectMeta) string {
	return fmt.Sprintf(`
FROM python:3.11-slim

WORKDIR /app

COPY requirements.txt .

RUN pip install --no-cache-dir -r requirements.txt

COPY . .

EXPOSE %s

CMD ["python", "%s"]
`, meta.Port, meta.EntryFile)
}

func dockerfileNode(meta detect.ProjectMeta) string {
	return fmt.Sprintf(`
FROM node:20

WORKDIR /app

COPY package*.json ./
RUN npm install

COPY . .

EXPOSE %s

CMD ["node", "%s"]
`, meta.Port, meta.EntryFile)
}

func dockerfileReact() string {
	return `
FROM node:20 AS builder
WORKDIR /app

COPY package*.json ./
RUN npm install

COPY . .

RUN npm run build

FROM nginx:alpine
COPY --from=builder /app/build /usr/share/nginx/html
`
}

func dockerfileNext() string {
	return `
FROM node:20 AS deps
WORKDIR /app

COPY package*.json ./
RUN npm install

COPY . .

RUN npm run build

EXPOSE 3000
CMD ["npm", "start"]
`
}

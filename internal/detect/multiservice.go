package detect

import (
	"os"
	"path/filepath"
)

type MultiService struct {
	FrontendPath string
	BackendPath  string
}

func DetectMultiService(basePath string) MultiService {
	ms := MultiService{}

	possibleFrontends := []string{"frontend", "client", "web", "ui"}
	possibleBackends := []string{"backend", "server", "api"}

	for _, f := range possibleFrontends {
		p := filepath.Join(basePath, f)
		if exists(p) {
			ms.FrontendPath = p
			break
		}
	}

	for _, b := range possibleBackends {
		p := filepath.Join(basePath, b)
		if exists(p) {
			ms.BackendPath = p
			break
		}
	}

	return ms
}

func exists(path string) bool {
	_, err := os.Stat(path)
	return !os.IsNotExist(err)
}

// Package apptesting prvides global utils for tests
package apptesting

import (
	"encoding/json"
	"fmt"
	"iredparser/common"
	"os"
	"path/filepath"
)

const testCredsFile = ".test.creds.json"

type Cred struct {
	Server   string `json:"server"`
	Login    string `json:"login"`
	Password string `json:"password"`
}

func GetAuthConfigs() ([]common.ServerConfig, error) {
	var configs []common.ServerConfig

	rootDir, err := findProjectRoot()
	if err != nil {
		return nil, fmt.Errorf("failed to find project root: %w", err)
	}

	credsPath := filepath.Join(rootDir, testCredsFile)

	credsFile, err := os.Open(credsPath)
	if err != nil {
		return nil, err
	}
	defer credsFile.Close()

	err = json.NewDecoder(credsFile).Decode(&configs)
	if err != nil {
		return nil, err
	}

	return configs, nil
}

func findProjectRoot() (string, error) {
	dir, err := os.Getwd()
	if err != nil {
		return "", err
	}

	for {
		if _, err := os.Stat(filepath.Join(dir, "pyproject.toml")); err == nil {
			return dir, nil
		}

		parent := filepath.Dir(dir)
		if parent == dir {
			break // Дошли до корня диска FS
		}
		dir = parent
	}

	return "", fmt.Errorf("project root not found")
}

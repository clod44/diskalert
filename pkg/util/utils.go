package util

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
)
func ResolvePath(path string) (string, error) {
    if filepath.IsAbs(path) {
        return path, nil
    }

    cwd, err := os.Getwd()
    if err != nil {
        return "", fmt.Errorf("failed to get current working directory: %w", err)
    }

    return filepath.Join(cwd, path), nil
}

func GetAppDir() string {
	executablePath, err := os.Executable()
	if err != nil {
		log.Fatalf("Could not determine executable path: %v", err)
	}
	return filepath.Dir(executablePath)
}
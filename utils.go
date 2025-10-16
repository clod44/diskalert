package main
import (
	"fmt"
	"os"
	"path/filepath"
)
func resolvePath(path string) (string, error) {
    if filepath.IsAbs(path) {
        return path, nil
    }

    cwd, err := os.Getwd()
    if err != nil {
        return "", fmt.Errorf("failed to get current working directory: %w", err)
    }

    return filepath.Join(cwd, path), nil
}
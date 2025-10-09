package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
)

func getDefaultConfig() Config {
	return Config{
		LogToFile:            true,
		LogFilePath:          "./",
		DiskPath:             "/",
		Threshold:            80,
		CheckIntervalSeconds: 5,
		Port:                 6969,
	}
}

func getAppDir() string {
	executablePath, err := os.Executable()
	if err != nil {
		log.Fatalf("Could not determine executable path: %v", err)
	}
	return filepath.Dir(executablePath)
}

func loadConfig() Config {
	appDir := getAppDir()
	configFilePath := filepath.Join(appDir, configFileName)

	data, err := os.ReadFile(configFilePath)
	if err == nil {
		var cfg Config
		if json.Unmarshal(data, &cfg) == nil {
			fmt.Printf("Configuration loaded from %s.\n", configFilePath)
			return cfg
		}
	}

	cfg := getDefaultConfig()
	data, _ = json.MarshalIndent(cfg, "", "  ")

	if os.WriteFile(configFilePath, data, 0644) == nil {
		fmt.Printf("Configuration file not found or invalid. Created default config at %s.\n", configFilePath)
	} else {
		fmt.Println("Failed to write default config file. Using in-memory defaults.")
	}
	return cfg
}

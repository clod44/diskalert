package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
)

type Config struct {
	LogToFile            bool   `json:"log_to_file"`
	LogFilePath          string `json:"log_file_path"`
	DiskPath             string `json:"disk_path"`
	Threshold            int    `json:"threshold"`
	CheckIntervalSeconds int    `json:"check_interval_seconds"`
	Port                 int    `json:"port"`
	CertDir              string `json:"cert_dir"`
	CertFileName         string `json:"cert_file_name"`
	KeyFileName          string `json:"key_file_name"`
	IP                   string `json:"ip"`
	VapidDir			 string `json:"vapid_dir"`
	VapidSecretKey		 string `json:"vapid_secret_key"`
	VapidPublicKey		 string `json:"vapid_public_key"`
}

func getDefaultConfig() Config {
	return Config{
		LogToFile:            true,
		LogFilePath:          "./",
		DiskPath:             "/",
		Threshold:            80,
		CheckIntervalSeconds: 5,
		Port:                 6969,
		CertDir:              "./ssl",
		CertFileName:         "diskalert.crt",
		KeyFileName:          "diskalert.key",
		IP:                   "192.168.66.153",
		VapidDir:			  "./vapid",
		VapidSecretKey: 	  "vapid_secret_key",
		VapidPublicKey:		  "vapid_public_key",

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
	var configFileName string = "diskalert.config.json"
	appDir := getAppDir()
	configFilePath := filepath.Join(appDir, configFileName)

	defaultCfg := getDefaultConfig()
	config := defaultCfg

	data, err := os.ReadFile(configFilePath)
	if err != nil {
		log.Printf("Configuration file not found or invalid. Created default config at %s.", configFilePath)

		out, _ := json.MarshalIndent(defaultCfg, "", "    ")
		os.WriteFile(configFilePath, out, 0644)

		return defaultCfg
	}

	if err := json.Unmarshal(data, &config); err != nil {
		log.Fatalf("Error parsing configuration file: %v", err)
	}

	log.Printf("Configuration loaded from %s.", configFilePath)

	fmt.Printf("DEBUG: Final Loaded Config: %+v\n", config)

	return config
}

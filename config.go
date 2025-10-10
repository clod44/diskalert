package main

import (
	"encoding/json"
	"log"
	"os"
	"path/filepath"
	"reflect"
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
	const configFileName = "diskalert.config.json"
	configFilePath := filepath.Join(getAppDir(), configFileName)

	finalConfig := getDefaultConfig()
	
	data, err := os.ReadFile(configFilePath)

	if os.IsNotExist(err) {
		log.Printf("Configuration file not found. Creating default config at %s.", configFilePath)
		
		out, _ := json.MarshalIndent(finalConfig, "", "    ")
		if writeErr := os.WriteFile(configFilePath, out, 0644); writeErr != nil {
			log.Printf("Warning: Failed to write default config file: %v", writeErr)
		}
		
		log.Println("--- Config (New Default) ---")
		logJson(finalConfig) 
		return finalConfig
	}
	if err != nil {
		log.Fatalf("Error reading configuration file. you may want to fix the issue or delete the config itself for automatic recreation. %s: %v", configFilePath, err)
	}


	var rawConfigMap map[string]interface{}
	if err := json.Unmarshal(data, &rawConfigMap); err != nil {
		log.Fatalf("Error parsing configuration file. you may want to fix the issue or delete the config itself for automatic recreation. %s: %v", configFilePath, err)
	}
	
	if err := json.Unmarshal(data, &finalConfig); err != nil {
		log.Fatalf("Error merging configuration data: %v", err)
	}
	
	t := reflect.TypeOf(finalConfig)
	
	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		jsonKey := field.Tag.Get("json")
		if jsonKey == "" {
			continue
		}
		if _, exists := rawConfigMap[jsonKey]; !exists {
			defaultValue := reflect.ValueOf(finalConfig).Field(i).Interface()
			log.Printf("WARN: Field '%s' was missing. Adding default value: %v", jsonKey, defaultValue)
		}
	}

	out, _ := json.MarshalIndent(finalConfig, "", "    ")
	if writeErr := os.WriteFile(configFilePath, out, 0644); writeErr != nil {
		log.Println("merged config file:")
		logJson(out)
		log.Printf("Warning: Failed to write merged config file: %v", writeErr)
	}
	
	log.Printf("Configuration loaded and merged from %s. File structure updated.", configFilePath)
	log.Println("--- Config (Loaded and Merged) ---")
	logJson(finalConfig) 
	
	return finalConfig
}
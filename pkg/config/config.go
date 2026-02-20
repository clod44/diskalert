package config

import (
	"diskalert/pkg/util"
	"encoding/json"

	"log"
	"os"
	"path/filepath"
)

const configFileName = "diskalert.conf"

var Cfg Config = LoadConfig();
type Config struct {
	LogToFile            bool   `json:"log_to_file"`
	LogFile          	 string `json:"log_file"`
	ExcludePaths         []string `json:"exclude_paths"`
	Threshold            int    `json:"threshold"`
	CheckIntervalSeconds int    `json:"check_interval_seconds"`
	Port                 int    `json:"port"`
	CertFile             string `json:"cert_file"`
	KeyFile              string `json:"key_file"`
	VapidPublic          string `json:"vapid_public"`
	VapidPrivate         string `json:"vapid_private"`
	IP                   string `json:"ip"`
}

func getDefaultConfig() Config {
	return Config{
		LogToFile:            true,
		LogFile:              "./logs/diskalert.log",
		ExcludePaths:         []string{"/sys", "/proc", "/dev", "/run"}, 
		Threshold:            80,
		CheckIntervalSeconds: 5,
		Port:                 6969,
		CertFile:             "./ssl/diskalert.crt",
		KeyFile:              "./ssl/diskalert.key",
		VapidPublic:          "./vapid/vapid_public",
		VapidPrivate:         "./vapid/vapid_private",
		IP:                   "127.0.0.1",
	}
}


func LoadConfig() Config {
	appDir := util.GetAppDir()
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
    
	return config
}
package main

import (
	"log"
	"sync"
	"time"
)

const configFileName = "diskalert.config.json"
const logFileName = "diskalert.log"

type Config struct {
	LogToFile            bool   `json:"log_to_file"`
	LogFilePath          string `json:"log_file_path"`
	DiskPath             string `json:"disk_path"`
	Threshold            int    `json:"threshold"`
	CheckIntervalSeconds int    `json:"check_interval_seconds"`
	Port                 int    `json:"port"`
}

// DiskStatus holds the metrics and uses a Mutex for thread-safe access.
type DiskStatus struct {
	mu          sync.RWMutex
	DiskPath    string  `json:"disk_path"`
	Threshold   int     `json:"threshold"`
	TotalGB     float64 `json:"total_gb"`
	FreeGB      float64 `json:"free_gb"`
	UsedPercent float64 `json:"used_percent"`
	IsAlert     bool    `json:"is_alert"`
	LastCheck   string  `json:"last_check"`
}

// Global variable to hold the latest metrics, accessible by monitor.go and web.go
var CurrentStatus = &DiskStatus{}

func main() {
	config := loadConfig()

	logFileHandle := setupLogger(config)
	if logFileHandle != nil {
		defer logFileHandle.Close()
	}

	CurrentStatus.DiskPath = config.DiskPath
	CurrentStatus.Threshold = config.Threshold

	go startWebServer(config) //"go" makes it a background process type shi without blocking the flow

	interval := time.Duration(config.CheckIntervalSeconds) * time.Second
	log.Printf("Monitoring %s every %d seconds. Threshold is %d%%.", config.DiskPath, config.CheckIntervalSeconds, config.Threshold)
	log.Println("--------------------------------------------------------------------------------")

	for {
		log.Println("Checking disk stats...")

		checkDiskUsage(config)

		nextCheckTime := time.Now().Add(interval)
		log.Printf("Next check will be %d seconds later, at %s.", config.CheckIntervalSeconds, nextCheckTime.Format("15:04:05"))

		time.Sleep(interval)
	}
}

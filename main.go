package main

import (
	"log"
	"time"
)

// Global variable to hold the latest metrics, accessible by monitor.go and web.go
var CurrentStatus = &DiskStatus{}

func main() {
	var config = loadConfig()

	logFileHandle := setupLogger(config)
	if logFileHandle != nil {
		defer logFileHandle.Close()
	}

	CurrentStatus.DiskPath = config.DiskPath
	CurrentStatus.Threshold = config.Threshold

	setupTLSFiles(config)     // this gotta be a blocking process so web server doesnt start before this
	InitSubscriptionDB(config)
	setupVAPIDKeys(config)
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

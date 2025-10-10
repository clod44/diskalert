package main

import (
	"log"
	"time"
)


var CurrentStatus = &DiskStatus{}
var cfg Config 

func main() {
	cfg = loadConfig()

	logFileHandle := setupLogger()
	if logFileHandle != nil {
		defer logFileHandle.Close()
	}

	CurrentStatus.DiskPath = cfg.DiskPath
	CurrentStatus.Threshold = cfg.Threshold

	setupTLSFiles()     // this gotta be a blocking process so web server doesn't start before this
	InitSubscriptionDB()
	setupVAPIDKeys()
	go startWebServer() //"go" makes it a background process type shi without blocking the flow

	interval := time.Duration(cfg.CheckIntervalSeconds) * time.Second
	log.Printf("Monitoring %s every %d seconds. Threshold is %d%%.", cfg.DiskPath, cfg.CheckIntervalSeconds, cfg.Threshold)
	log.Println("--------------------------------------------------------------------------------")

	for {
		log.Println("Checking disk stats...")
		checkDiskUsage() 
		
		if(CurrentStatus.IsAlert){
			err := SendAlertsToAllSubscribers("crazy", "message") 
			if err != nil { 
				log.Printf("Error sending alerts: %v", err)
			}
		}

		nextCheckTime := time.Now().Add(interval)
		log.Printf("Next check will be %d seconds later, at %s.", cfg.CheckIntervalSeconds, nextCheckTime.Format("15:04:05"))
		time.Sleep(interval)
	}
}

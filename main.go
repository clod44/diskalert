package main

import (
	"fmt"
	"log"
	"time"
)

type App struct {
	diskStatus DiskStatus
	vapidPrivateContent string
	vapidPublic64 string 
	cfg Config
}
var APP = &App{} //don't use direct instancing to prevent passing values to the scopes. use pointers to pass the original object



func main() {
	APP.cfg = loadConfig()

	logFileHandle := setupLogger()
	if logFileHandle != nil {
		defer logFileHandle.Close()
	}


	setupTLSFiles()     // this gotta be a blocking process so web server doesn't start before this
	InitSubscriptionDB()
	setupVAPIDKeys()
	go startWebServer() //"go" makes it a background process type shi without blocking the flow
	ManageServiceFile()
	
	interval := time.Duration(APP.cfg.CheckIntervalSeconds) * time.Second
	log.Printf("Monitoring %s every %d seconds. Threshold is %d%%.", APP.cfg.DiskPath, APP.cfg.CheckIntervalSeconds, APP.cfg.Threshold)
	log.Println("--------------------------------------------------------------------------------")

	for {
		log.Println("Checking disk stats...")
		UpdateDiskStatus() 
		
		if(APP.diskStatus.IsAlert){
			 alertMessage := fmt.Sprintf(
				"Disk usage is over threshold! %.1f%%. Contact integration team",
				APP.diskStatus.UsedPercent,
			)
			err := SendNotificationToAll("DISK USAGE ALERT", alertMessage) 
			if err != nil { 
				log.Printf("Error sending alerts: %v", err)
			}
		}

		nextCheckTime := time.Now().Add(interval)
		log.Printf("Next check will be %d seconds later, at %s.", APP.cfg.CheckIntervalSeconds, nextCheckTime.Format("15:04:05"))
		time.Sleep(interval)
	}
}

package main

import (
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


	SetupTLSFiles()     // this gotta be a blocking process so web server doesn't start before this
	InitSubscriptionDB()
	InitDiskHistoryDB()
	SetupVAPIDKeys()
	go StartWebServer() //"go" makes it a background process type shi without blocking the flow
	ManageServiceFile()
	
	interval := time.Duration(APP.cfg.CheckIntervalSeconds) * time.Second
	log.Printf("Monitoring %s every %d seconds. Threshold for all disks is %d%%.", APP.cfg.CheckIntervalSeconds, APP.cfg.Threshold)
	log.Println("--------------------------------------------------------------------------------")

	for {
		log.Println("Checking disk stats...")
		UpdateDiskStatus() 
		
		logMonitoredDisks(APP.diskStatus.Records)

		nextCheckTime := time.Now().Add(interval)
		log.Printf("Next check will be %d seconds later, at %s.", APP.cfg.CheckIntervalSeconds, nextCheckTime.Format("15:04:05"))
		time.Sleep(interval)
	}
}

func logMonitoredDisk(disk MonitoredDisk) {
	log.Printf("  Disk: [%.1f%%] %s (%d bytes used / %d bytes total) (%s)", disk.UsedPercent, disk.DiskPath, disk.UsedSize, disk.TotalSize, disk.UUID)
}

func logMonitoredDisks(disks []MonitoredDisk) (bool) {
	APP.diskStatus.mu.RLock()
	defer APP.diskStatus.mu.RUnlock()

	overallAlert := false
	
	log.Printf("Disk Usage Alert Report (Threshold: %d%%)", APP.cfg.Threshold)
	log.Println("--------------------------------------------------")

	for _, disk := range disks{
		if disk.IsAlert {
			overallAlert = true
			logMonitoredDisk(disk)
			log.Println("")
		}
	}

	if overallAlert {
		log.Println("--------------------------------------------------")
		log.Println("CRITICAL ALERT: One or more disks are over the usage threshold.")
	} else {
		log.Println("INFO: No disks currently in alert state.")
	}

	return overallAlert
}
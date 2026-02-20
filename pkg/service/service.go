package service

import (
	"diskalert/pkg/util"
	"fmt"
	"log"
	"os"
	"os/user"
	"path/filepath"
)

const serviceFileName = "diskalert.service"
const serviceFileContent = `[Unit]
Description=Disk Alert Monitor
After=network.target

[Service]
User=%s
Group=%s
WorkingDirectory=%s
ExecStart=%s/diskalert
Restart=always
RestartSec=5

[Install]
WantedBy=default.target`

func GetServicePath() (string, error) {
	var serviceDir = "/etc/systemd/system/";
	return filepath.Join(serviceDir, serviceFileName), nil
}

func ManageServiceFile() error {
	serviceFilePath, err := GetServicePath()
	if err != nil {
		return err
	}

	if _, err := os.Stat(serviceFilePath); err == nil {
		log.Printf("Service file already exists at: %s. Delete it manually for a fresh installation.", serviceFilePath)
		return nil
	} else if !os.IsNotExist(err) {
		return fmt.Errorf("error checking service file status: %w", err)
	}

	serviceDir := filepath.Dir(serviceFilePath)
	if err := os.MkdirAll(serviceDir, 0755); err != nil {
		return fmt.Errorf("failed to create systemd user directory %s: %w", serviceDir, err)
	}

	currentUser, err := user.Current()
	if err != nil {
		return fmt.Errorf("failed to get current user details: %w", err)
	}
	appDir := util.GetAppDir() 

	content := fmt.Sprintf(serviceFileContent, 
		currentUser.Username,
		currentUser.Gid,
		appDir,
		appDir,
	)

	if err := os.WriteFile(serviceFilePath, []byte(content), 0644); err != nil {
		return fmt.Errorf("failed to write service file to %s: %w", serviceFilePath, err)
	}

	log.Printf("Successfully created service file: %s", serviceFilePath)
    
    log.Println("----------------------------------------------------------------")
	log.Println("🎉 Service file created successfully!")
	log.Println("You can stop this instance of the diskalert and run it as a service.")
	log.Println("NEXT STEPS (required for user services):")
	log.Println("1. Reload the systemd manager (MUST BE DONE FIRST):")
	log.Println("   systemctl daemon-reload")
	log.Println("2. Enable the service to start on boot:")
	log.Println("   systemctl enable diskalert.service")
	log.Println("3. Start the service now:")
	log.Println("   systemctl start diskalert.service    --or--    service diskalert start")
    log.Println("4. Check the service status:")
	log.Println("   systemctl status diskalert.service    --or--    service diskalert status")
    log.Println("----------------------------------------------------------------")
    
	return nil
}
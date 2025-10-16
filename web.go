package main

import (
	"encoding/json"
	"fmt"
	"io/fs"
	"log"
	"net/http"
	"os"
	"path/filepath"
)

// resolveConfigPath resolves a configuration path.
// If the path is absolute, it returns the path as is.
// If the path is relative, it resolves it relative to the application's directory (getAppDir()).
func resolveConfigPath(configPath string) string {
	if filepath.IsAbs(configPath) {
		return configPath
	}
	return filepath.Join(getAppDir(), configPath)
}

func statsHandler(w http.ResponseWriter, r *http.Request) {
	APP.diskStatus.mu.RLock()
	defer APP.diskStatus.mu.RUnlock()
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(APP.diskStatus); err != nil {
		http.Error(w, "Failed to encode metrics", http.StatusInternalServerError)
		log.Printf("ERROR: Failed to encode metrics to JSON: %v", err)
	}
}

func vapidKeyHandler(w http.ResponseWriter, r *http.Request) {
	if APP.vapidPublic64 == "" {
		http.Error(w, "VAPID public key not initialized", http.StatusInternalServerError)
		log.Println("ERROR: VAPID public key is empty.")
		return
	}
	w.Header().Set("Content-Type", "text/plain")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(APP.vapidPublic64))
}

func subscribeHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	var sub PushSubscription
	if err := json.NewDecoder(r.Body).Decode(&sub); err != nil {
		log.Printf("ERROR: Failed to decode subscription JSON: %v", err)
		http.Error(w, "Invalid subscription format", http.StatusBadRequest)
		return
	}
	if err := AddSubscription(sub); err != nil {
		log.Printf("ERROR: Failed to add subscription to DB: %v", err)
		http.Error(w, "Failed to save subscription", http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusCreated) 
	if _, err := w.Write([]byte(`{"message": "Subscription successful"}`)); err != nil {
		log.Printf("WARN: Failed to write success response: %v", err)
	}
}

func unsubscribeHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	var sub PushSubscription
	if err := json.NewDecoder(r.Body).Decode(&sub); err != nil {
		log.Printf("ERROR: Failed to decode unsubscription JSON: %v", err)
		http.Error(w, "Invalid request format", http.StatusBadRequest)
		return
	}
	if sub.Endpoint == "" {
		http.Error(w, "Endpoint required for unsubscription", http.StatusBadRequest)
		return
	}
	if err := RemoveSubscription(sub.Endpoint); err != nil {
		log.Printf("ERROR: Failed to remove subscription: %v", err)
		http.Error(w, "Failed to remove subscription from database", http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"message": "Unsubscription successful"}`))
}

func subscriptionsHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	subscriptions, err := GetAllSubscriptions()
	if err != nil {
		log.Printf("ERROR: Failed to retrieve subscriptions from DB: %v", err)
		http.Error(w, "Failed to retrieve subscriptions", http.StatusInternalServerError)
		return
	}
	if err := json.NewEncoder(w).Encode(subscriptions); err != nil {
		log.Printf("ERROR: Failed to encode subscriptions JSON: %v", err)
		http.Error(w, "Failed to serialize subscriptions", http.StatusInternalServerError)
		return
	}
}

func testNotificationHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	var req struct {
		TargetEndpoint *string `json:"target_endpoint"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		log.Printf("ERROR: Failed to decode notification JSON: %v", err)
		http.Error(w, "Invalid request format", http.StatusBadRequest)
		return
	}
	if req.TargetEndpoint == nil {
		http.Error(w, "Target endpoint required for notification", http.StatusBadRequest)
		return
	}
	/*
	if err := SendNotification(req.TargetEndpoint, NotificationPayload{Title: "Test Notification", Message: "This is a test notification. you can ignore this"}); err != nil {
		log.Printf("ERROR: Failed to send notification: %v", err)
		http.Error(w, "Failed to send notification", http.StatusInternalServerError)
		return
	}
	*/
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"message": "Notification sent successfully"}`))
}

func handleCertDownload(w http.ResponseWriter, r *http.Request) {
	// Use the new resolver function
	certFilePath := resolveConfigPath(APP.cfg.CertFile)
	
	// Use filepath.Base to get just the filename for the Content-Disposition header
	certFileName := filepath.Base(APP.cfg.CertFile) 

	if _, err := os.Stat(certFilePath); os.IsNotExist(err) {
		log.Printf("Certificate file not found at: %s", certFilePath)
		http.Error(w, "Certificate file not found.", http.StatusNotFound)
		return
	}
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=%s", certFileName))
	w.Header().Set("Content-Type", "application/x-pem-file")
	http.ServeFile(w, r, certFilePath)
	log.Printf("Served certificate file: %s to client %s", certFileName, r.RemoteAddr)
}

func startWebServer() {
	// Use the new resolver function for both certificate and key files
	var certFilePath = resolveConfigPath(APP.cfg.CertFile)
	var keyFilePath = resolveConfigPath(APP.cfg.KeyFile)

	http.HandleFunc("/api/stats", statsHandler)
	http.HandleFunc("/api/vapid-key", vapidKeyHandler)
	http.HandleFunc("/api/subscribe", subscribeHandler)
	http.HandleFunc("/api/unsubscribe", unsubscribeHandler) 
	http.HandleFunc("/api/subscriptions", subscriptionsHandler)
	http.HandleFunc("/api/test-notification", testNotificationHandler)
	http.HandleFunc("/download/cert", handleCertDownload)

	publicFS, err := fs.Sub(Assets, "public")
	if err != nil {
		log.Fatalf("Failed to create sub-filesystem for 'public' (check if public folder is correctly embedded): %v", err)
	}
	fileServer := http.FileServer(http.FS(publicFS))
	http.Handle("/", fileServer)

	bindAddress := fmt.Sprintf(":%d", APP.cfg.Port)

	log.Println("========================================================================================")
	log.Printf("DiskAlert Web Dashboard is now LIVE!")
	log.Printf("Access Dashboard at: https://<YOUR_IP>:%d/", APP.cfg.Port)
	log.Printf("JSON Metrics API:    https://<YOUR_IP>:%d/api/metrics", APP.cfg.Port)
	log.Println("========================================================================================")

	listenErr := http.ListenAndServeTLS(bindAddress, certFilePath, keyFilePath, nil)
	if listenErr != nil {
		log.Fatalf("HTTPS Web server failed to start on port %d: %v", APP.cfg.Port, listenErr)
	}
}
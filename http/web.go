package http

import (
	"diskalert/diskmon"
	"diskalert/pkg/config"
	"diskalert/pkg/util"
	"diskalert/storage/diskdb"
	"encoding/json"
	"fmt"
	"io/fs"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

func forecastHandler(w http.ResponseWriter, r *http.Request) {
	defaultLimit := 50
	w.Header().Set("Content-Type", "application/json")
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	query := r.URL.Query()	
	uuid := query.Get("uuid")
	if uuid == "" {
		http.Error(w, "UUID parameter is required", http.StatusBadRequest)
		return
	}
	limit := defaultLimit
	limitStr := query.Get("limit")
	if limitStr != "" {
		parsedLimit, err := strconv.Atoi(limitStr)
		if err == nil && parsedLimit > 0 {
			limit = parsedLimit
		} else {
			log.Printf("WARN: Invalid 'limit' parameter provided, using default %d. Error: %v", defaultLimit, err)
		}
	}
	diskRecords, err := diskmon.GetDiskForecasts(uuid, limit)
	if err != nil {
		log.Printf("ERROR: Database query failed for UUID %s: %v", uuid, err)
		http.Error(w, "Internal server error while fetching history", http.StatusInternalServerError)
		return
	}
	jsonData, err := json.Marshal(diskRecords)
	if err != nil {
		log.Printf("ERROR: Failed to marshal disk records to JSON: %v", err)
		http.Error(w, "Internal error processing data", http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
	w.Write(jsonData)
}
func historyHandler(w http.ResponseWriter, r *http.Request) {
	defaultLimit := 30
	w.Header().Set("Content-Type", "application/json")
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	query := r.URL.Query()	
	uuid := query.Get("uuid")
	if uuid == "" {
		http.Error(w, "UUID parameter is required", http.StatusBadRequest)
		return
	}
	limit := defaultLimit
	limitStr := query.Get("limit")
	if limitStr != "" {
		parsedLimit, err := strconv.Atoi(limitStr)
		if err == nil && parsedLimit > 0 {
			limit = parsedLimit
		}else {
			log.Printf("WARN: Invalid 'limit' parameter provided, using default %d. Error: %v", defaultLimit, err)
		}
	}
	diskRecords, err := diskdb.GetDiskRecords(uuid, limit)
	if err != nil {
		log.Printf("ERROR: Database query failed for UUID %s: %v", uuid, err)
		http.Error(w, "Internal server error while fetching history", http.StatusInternalServerError)
		return
	}
	jsonData, err := json.Marshal(diskRecords)
	if err != nil {
		log.Printf("ERROR: Failed to marshal disk records to JSON: %v", err)
		http.Error(w, "Internal error processing data", http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
	w.Write(jsonData)
}
func statsHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	diskStatus.mu.RLock()
	defer diskStatus.mu.RUnlock()
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(diskStatus); err != nil {
		http.Error(w, "Failed to encode metrics", http.StatusInternalServerError)
		log.Printf("ERROR: Failed to encode metrics to JSON: %v", err)
	}
}

func vapidKeyHandler(w http.ResponseWriter, r *http.Request) {
	if vapidPublic64 == "" {
		http.Error(w, "VAPID public key not initialized", http.StatusInternalServerError)
		log.Println("ERROR: VAPID public key is empty.")
		return
	}
	w.Header().Set("Content-Type", "text/plain")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(vapidPublic64))
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
		SendNotificationToAll("Test Notification", "This is a test notification. You can ignore this.")
		return;
	}
	subscription, err := GetSubscription(*req.TargetEndpoint)
	if err != nil {
		log.Printf("ERROR: Failed to retrieve subscription for endpoint %s: %v", *req.TargetEndpoint, err)
		http.Error(w, "Failed to retrieve subscription", http.StatusInternalServerError)
		return
	}
	SendNotification(*subscription, NotificationPayload{Title: "Test Notification", Message: "This is a test notification. you can ignore this"})
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"message": "Notification sent successfully"}`))
}

func handleCertDownload(w http.ResponseWriter, r *http.Request) {
	certFilePath, err := util.ResolvePath(Cfg.CertFile)
	if err != nil {
		log.Printf("ERROR: Failed to resolve path for certificate file: %v", err)
		http.Error(w, "Failed to resolve path for certificate file.", http.StatusInternalServerError)
		return
	}
	certFileName := filepath.Base(config.Cfg.CertFile) 
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

func homeHandler(w http.ResponseWriter, r *http.Request) {
    data := HomeData{
        DiskStatus: diskmon.DiskStatus,
    } 
	err := tpl.Execute(w, data)

    if err != nil {
        http.Error(w, "Could not render template page", http.StatusInternalServerError)
        fmt.Printf("Template execution error for /test: %v\n", err)
        return
    }
}

func serviceWorkerHeaderHandler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// The request path is relative to the Handler path, so for /public/sw.js
		// the r.URL.Path here will contain the full path.
		if strings.HasSuffix(r.URL.Path, "sw.js") {
			// This header is what tells the browser the script is allowed to control the root scope (/).
			w.Header().Set("Service-Worker-Allowed", "/")
		}
		// Serve the actual file content using the next handler (your FileServer)
		next.ServeHTTP(w, r)
	})
}


func StartWebServer() {
	certFilePath, err := util.ResolvePath(config.Cfg.CertFile)
	if err != nil {
		log.Printf("ERROR: Failed to resolve path for certificate file: %v", err)
		return
	}
	keyFilePath, err := util.ResolvePath(config.Cfg.KeyFile)
	if err != nil {
		log.Printf("ERROR: Failed to resolve path for key file: %v", err)
		return
	}
	
	http.HandleFunc("/", homeHandler)
	http.HandleFunc("/api/stats", statsHandler)
	http.HandleFunc("/api/history", historyHandler)
	http.HandleFunc("/api/forecast", forecastHandler)
	http.HandleFunc("/api/vapid-key", vapidKeyHandler)
	http.HandleFunc("/api/subscribe", subscribeHandler)
	http.HandleFunc("/api/unsubscribe", unsubscribeHandler) 
	http.HandleFunc("/api/subscriptions", subscriptionsHandler)
	http.HandleFunc("/api/test-notification", testNotificationHandler)
	http.HandleFunc("/download/cert", handleCertDownload)
	

	publicFS, err := fs.Sub(Assets, "public")
	if err != nil {
		log.Fatalf("Failed to create sub-filesystem for 'public': %v", err)
	}
	fileServer := http.FileServer(http.FS(publicFS))
	wrappedFileServer := serviceWorkerHeaderHandler(fileServer)
	http.Handle("/public/", http.StripPrefix("/public/", wrappedFileServer))
	
	bindAddress := fmt.Sprintf(":%d", config.Cfg.Port)

	log.Println("========================================================================================")
	log.Printf("DiskAlert Web Dashboard is now LIVE!")
	log.Printf("Access Dashboard at: https://<YOUR_IP>:%d/", config.Cfg.Port)
	log.Printf("JSON Metrics API:    https://<YOUR_IP>:%d/api/metrics", config.Cfg.Port)
	log.Println("========================================================================================")

	listenErr := http.ListenAndServeTLS(bindAddress, certFilePath, keyFilePath, nil)
	if listenErr != nil {
		log.Fatalf("HTTPS Web server failed to start on port %d: %v", config.Cfg.Port, listenErr)
	}
}
package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"path/filepath"

	webpush "github.com/SherClockHolmes/webpush-go"
	_ "modernc.org/sqlite"
)

type Keys struct {
	P256DH string `json:"p256dh"`
	Auth 	 string `json:"auth"`
}

type PushSubscription struct {
	Endpoint string `json:"endpoint"`
	Keys 	 Keys 	`json:"keys"`
}

type NotificationPayload struct {
	Title 	string 	`json:"title"` 	
	Message string 	`json:"message"` 
}

// DB is the global database connection pool.
var DB *sql.DB

func InitSubscriptionDB() {
	dbFileName := "subscriptions.db"
	dbFilePath := filepath.Join(getAppDir(), dbFileName)
	
	db, err := sql.Open("sqlite", dbFilePath)
	if err != nil {
		log.Fatalf("FATAL: Failed to open SQLite database: %v", err)
	}
	
	DB = db
	if _, err := DB.Exec("PRAGMA journal_mode = WAL;"); err != nil {
        log.Fatalf("FATAL: Failed to enable WAL mode: %v", err)
    }
	query := `
	CREATE TABLE IF NOT EXISTS subscriptions (
		endpoint TEXT PRIMARY KEY,
		p256dh TEXT,
		auth TEXT
	);`

	if _, err := DB.Exec(query); err != nil {
		DB.Close()
		log.Fatalf("FATAL: Failed to create subscriptions table: %v", err)
	}
	
	log.Printf("SQLite database initialized at %s.", dbFilePath)

	var count int
	row := DB.QueryRow("SELECT COUNT(*) FROM subscriptions")
	if err := row.Scan(&count); err == nil {
		log.Printf("Loaded %d existing push subscriptions from database.", count)
	}
}

func AddSubscription(sub PushSubscription) error {
	query := `
	INSERT OR REPLACE INTO subscriptions (endpoint, p256dh, auth) 
	VALUES (?, ?, ?)`
	
	_, err := DB.Exec(query, sub.Endpoint, sub.Keys.P256DH, sub.Keys.Auth)
	if err != nil {
		return err
	}
	
	log.Printf("Subscription added/reconciled for endpoint: %s", sub.Endpoint)
	return nil
}

func RemoveSubscription(endpoint string) error {
	query := `DELETE FROM subscriptions WHERE endpoint = ?`
	
	result, err := DB.Exec(query, endpoint)
	if err != nil {
		return err
	}
	
	rowsAffected, _ := result.RowsAffected()
	log.Printf("Subscription removed for endpoint %s. Rows affected: %d", endpoint, rowsAffected)
	return nil
}

func GetAllSubscriptions() ([]PushSubscription, error) {
	query := `SELECT endpoint, p256dh, auth FROM subscriptions`
	
	rows, err := DB.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var subscriptions []PushSubscription
	for rows.Next() {
		var sub PushSubscription
		var p256dh, auth string
		
		if err := rows.Scan(&sub.Endpoint, &p256dh, &auth); err != nil {
			return nil, err
		}
		
		sub.Keys = Keys{
			P256DH: p256dh,
			Auth: 	auth,
		}
		subscriptions = append(subscriptions, sub)
	}
	
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return subscriptions, nil
}

func GetSubscription(endpoint string) (*PushSubscription, error) {
	query := `SELECT * FROM subscriptions WHERE endpoint = ?`	
	row := DB.QueryRow(query, endpoint)
	var sub PushSubscription
	var p256dh, auth string
	if err := row.Scan(&sub.Endpoint, &p256dh, &auth); err != nil {
		if err == sql.ErrNoRows {
			return nil, nil // Not found
		}
		return nil, err
	}
	sub.Keys = Keys{
		P256DH: p256dh,
		Auth: 	auth,
	}
	return &sub, nil
}

func SendNotification(sub PushSubscription, payload NotificationPayload) {
	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		log.Printf("failed to marshal notification payload: %v", err)
	}

	wpSub := &webpush.Subscription{
			Endpoint: sub.Endpoint,
			Keys: webpush.Keys{
				P256dh: sub.Keys.P256DH, 
				Auth: sub.Keys.Auth,
			},
		}

	resp, err := webpush.SendNotification(payloadBytes, wpSub, &webpush.Options{
		Subscriber: 	 "mailto:admin@your-disk-monitor.com", 
		VAPIDPublicKey: 	APP.vapidPublic64,
		VAPIDPrivateKey:    APP.vapidPrivateContent,
		TTL: 	 			60 * 60 * 24, // 24 hours
	})

	if err != nil {
		if resp != nil && (resp.StatusCode == 404 || resp.StatusCode == 410) {
			log.Printf("Subscription expired/invalid for %s. Deleting from DB...", sub.Endpoint)
			if removeErr := RemoveSubscription(sub.Endpoint); removeErr != nil {
				log.Printf("Error cleaning up expired subscription %s: %v", sub.Endpoint, removeErr)
			}
		} else {
			log.Printf("Push notification failed for %s: %v", sub.Endpoint, err)
		}
	} else {
		log.Printf("Successfully sent push notification to %s", sub.Endpoint)
	}
	
	if resp != nil {
		resp.Body.Close()
	}
}

func SendNotificationToAll(title string, message string) error {
	subs, err := GetAllSubscriptions()
	if err != nil {
		return fmt.Errorf("could not retrieve subscriptions: %w", err)
	}

	if len(subs) == 0 {
		log.Println("No active subscriptions found. Skipping push notification.")
		return nil
	}

	payload := NotificationPayload{
		Title: title,
		Message: message,
	}
	
	log.Printf("Attempting to send alert to %d subscribers...", len(subs))

	for _, sub := range subs {
		SendNotification(sub, payload)
	}
	return nil
}
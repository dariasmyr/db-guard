package internal

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
)

func HandleBackupFailure(err error, backupFilePath, database string, webhookURL string) {
	err = os.Remove(backupFilePath)
	if err != nil {
		log.Fatalf("❗️Error removing backup file \"%s\": %v", backupFilePath, err)
	}
	log.Printf("❗️Error backing up database \"%s\"", database)
	if webhookURL != "" {
		sendWebhookNotification(map[string]string{
			"status":  "failure",
			"message": fmt.Sprintf("❗️Error backing up database \"%s\"", database),
		}, webhookURL)
	}
}

func HandleBackupSuccess(backupFilePath string, database string, webhookURL string) {
	log.Printf("Database \"%s\" backed up successfully. File name: \"%s\"\n", database, backupFilePath)
	if webhookURL != "" {
		sendWebhookNotification(map[string]string{
			"status":  "success",
			"message": fmt.Sprintf("🎉Database \"%s\" backed up successfully.\nFile name: \"%s\"", database, backupFilePath),
		}, webhookURL)
	}
}

func sendWebhookNotification(payload map[string]string, webhookURL string) {
	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		log.Printf("❗️Failed to marshal payload: %v", err)
		return
	}
	resp, err := http.Post(webhookURL, "application/json", bytes.NewBuffer(payloadBytes))
	if err != nil {
		log.Printf("❗️Failed to send webhook: %v", err)
		return
	}
	defer func() {
		if err := resp.Body.Close(); err != nil {
			log.Printf("❗️Failed to close response body: %v", err)
		}
	}()

	if resp.StatusCode != http.StatusOK {
		log.Printf("❗️Webhook returned non-OK status: %d", resp.StatusCode)
	}
}

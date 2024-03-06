package internal

import (
	"db_dump/internal"
	"fmt"
	"log"
	"os"
	"strconv"
)

func HandleBackupFailure(err error, backupFilePath, database string, telegramNotify bool) {
	os.Remove(backupFilePath)
	if telegramNotify {
		sendTelegramNotification(fmt.Sprintf("❗️Error backing up database %s", database))
	}
}

func HandleBackupSuccess(telegramNotify bool, backupFilePath, database, backupFileName string) {
	log.Printf("Database %s backed up successfully. File name: %s\n", database, backupFileName)
	if telegramNotify {
		sendTelegramNotification(fmt.Sprintf("🎉Database %s backed up successfully.\nFile name: %s", database, backupFileName))
	}
}

func SendTelegramNotification(message string) {
	if os.Getenv("TELEGRAM_BOT_TOKEN") != "" && os.Getenv("CHANNEL_ID") != "" {
		channelID, _ := strconv.ParseInt(os.Getenv("CHANNEL_ID"), 10, 64)
		_ = internal.SendMessage(channelID, message)
	}
}
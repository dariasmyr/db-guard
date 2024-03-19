package internal

import (
	"fmt"
	"log"
	"os"
	"strconv"
)

func HandleBackupFailure(err error, backupFilePath, database string, telegramNotify bool) {
	os.Remove(backupFilePath)
	log.Printf("❗️Error backing up database \"%s\"", database)
	if telegramNotify {
		sendTelegramNotification(fmt.Sprintf("❗️Error backing up database \"%s\"", database))
	}
}

func HandleBackupSuccess(telegramNotify bool, backupFilePath, database, backupFileName string) {
	log.Printf("Database \"%s\" backed up successfully. File name: \"%s\"\n", database, backupFileName)
	if telegramNotify {
		sendTelegramNotification(fmt.Sprintf("🎉Database \"%s\" backed up successfully.\nFile name: \"%s\"", database, backupFileName))
	}
}

func sendTelegramNotification(message string) {
	if os.Getenv("TELEGRAM_BOT_TOKEN") != "" && os.Getenv("CHANNEL_ID") != "" {
		channelID, _ := strconv.ParseInt(os.Getenv("CHANNEL_ID"), 10, 64)
		_ = SendMessage(channelID, message)
	}
}

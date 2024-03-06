package internal

import (
	"os"
)

func CreateBackupFile(backupFilePath string) (*os.File, error) {
	return os.Create(backupFilePath)
}
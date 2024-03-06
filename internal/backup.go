package internal

import (
	"compress/gzip"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

func BackupDatabase(host string, port int, user string, password string, database string, backupDir string, compress bool, compressionLevel int, telegramNotify bool) error {
	if password == "" {
		return fmt.Errorf("database password is required")
	}

	err := os.Setenv("PGPASSWORD", password)
	if err != nil {
		return fmt.Errorf("failed to set PGPASSWORD: %v", err)
	}

	backupFileName := generateBackupFileName(database, compress)

	backupFilePath := filepath.Join(backupDir, backupFileName)

	backupFile, err := CreateBackupFile(backupFilePath)
	if err != nil {
		return fmt.Errorf("failed to create backup file: %v", err)
	}
	defer backupFile.Close()

	compressionLevel = adjustCompressionLevel(compressionLevel)

	gzipWriter, err := CreateGzipWriter(backupFile, compressionLevel)
	if err != nil {
		return fmt.Errorf("failed to create gzip writer: %v", err)
	}
	defer gzipWriter.Close()

	cmdArgs := ConstructPgDumpCommandArgs(host, port, user, database)

	err = ExecuteBackupCommand(cmdArgs, compress, gzipWriter, backupFile)
	if err != nil {
		HandleBackupFailure(err, backupFilePath, database, telegramNotify)
		return fmt.Errorf("backup failed: %v", err)
	}

	HandleBackupSuccess(telegramNotify, backupFilePath, database, backupFileName)

	return nil
}

func adjustCompressionLevel(compressionLevel int) int {
	if compressionLevel < -2 || compressionLevel > 9 {
		return gzip.BestCompression
	}
	if compressionLevel == -1 {
		return gzip.DefaultCompression
	} else if compressionLevel == -2 {
		return gzip.HuffmanOnly
	}
	return compressionLevel
}

func generateBackupFileName(database string, compress bool) string {
	fileExtension := "sql"
	if compress {
		fileExtension = "sql.gz"
	}
	return fmt.Sprintf("%s-%s.%s", database, time.Now().Format("2006-01-02T15-04-05"), fileExtension)
}

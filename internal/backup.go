package internal

import (
	"db_dump/internal"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
)


func BackupDatabase(host string, port int, user string, password string, database string, backupDir string, compress bool, compressionLevel int, telegramNotify bool) error {
	if password == "" {
		return fmt.Errorf("database password is required")
	}

	err := os.Setenv("PGPASSWORD", password)
	if err != nil {
		return fmt.Errorf("failed to set PGPASSWORD: %v", err)
	}

	backupFileName := internal.GenerateBackupFileName(database, compress)

	backupFilePath := filepath.Join(backupDir, backupFileName)

	backupFile, err := internal.CreateBackupFile(backupFilePath)
	if err != nil {
		return fmt.Errorf("failed to create backup file: %v", err)
	}
	defer backupFile.Close()

	compressionLevel = internal.AdjustCompressionLevel(compressionLevel)

	gzipWriter, err := internal.CreateGzipWriter(backupFile, compressionLevel)
	if err != nil {
		return fmt.Errorf("failed to create gzip writer: %v", err)
	}
	defer gzipWriter.Close()

	cmdArgs := internal.ConstructPgDumpCommandArgs(host, port, user, database)

	err = internal.ExecuteBackupCommand(cmdArgs, compress, gzipWriter, backupFile)
	if err != nil {
		internal.HandleBackupFailure(err, backupFilePath, database, telegramNotify)
		return fmt.Errorf("backup failed: %v", err)
	}

	internal.HandleBackupSuccess(telegramNotify, backupFilePath, database, backupFileName)

	return nil
}



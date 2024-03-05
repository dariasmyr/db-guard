package main

import (
	"compress/gzip"
	"db_dump/internal"
	"flag"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strconv"
	"sync"
	"time"
)

type DatabaseBackupStatus struct {
	running bool
}

var (
	statusMap      map[string]*DatabaseBackupStatus
	statusMapMutex sync.Mutex
)

func main() {
	internal.InitParams()
	if os.Getenv("TELEGRAM_BOT_TOKEN") != "" && os.Getenv("CHANNEL_ID") != "" {
		go internal.StartBot()
	}

	log.Printf("Start parsing flags")

	flag.Parse()

	log.Printf("Finished parsing flags")

	// Check if pg_dump exists in the system
	log.Printf("Check if pg_dump exists in the system")
	if _, err := exec.LookPath("pg_dump"); err != nil {
		log.Fatal("pg_dump not found in system PATH")
	}

	// Ensure required flags are provided
	log.Printf("Ensure required flags are provided")
	if internal.User == "" || internal.Password == "" || internal.Database == "" {
		flag.PrintDefaults()
		os.Exit(1)
	}

	// Ensure backup directory exists
	log.Println("Ensure backup directory exists")
	backupDataDir := filepath.Join(internal.BackupDir)
	if _, err := os.Stat(backupDataDir); os.IsNotExist(err) {
		os.Mkdir(backupDataDir, os.ModePerm)
	}

	statusMap = make(map[string]*DatabaseBackupStatus)

	backupDone := make(chan struct{})

	// Start main routine
	go func() {
		defer close(backupDone)

		ticker := time.NewTicker(time.Duration(internal.IntervalSec) * time.Second)
		defer ticker.Stop()

		for {
			log.Printf("Starting new backup cycle")
			select {
			case <-ticker.C:
				// Check if backup for this database is already running
				if isBackupRunning(internal.Database) {
					log.Printf("Backup for database %s is already running, skipping this cycle", internal.Database)
					continue
				}

				// Mark backup for this database as running
				setBackupRunning(internal.Database, true)

				go func() {
					defer func() {
						// Mark backup for this database as not running after completion
						setBackupRunning(internal.Database, false)
					}()

					log.Printf("Initial backup")
					if err := backupDatabase(internal.Host, internal.Port, internal.User, internal.Password, internal.Database, internal.BackupDir, internal.Compress, internal.CompressionLevel, internal.TelegramNotify); err != nil {
						log.Println("Error backing up database:", err)
					}
				}()
				cleanupBackups(internal.BackupDir, internal.MaxBackupCount)
			}
		}
	}()

	select {
	case <-backupDone:
		log.Println("Backup completed")
	}
}

func isBackupRunning(database string) bool {
	statusMapMutex.Lock()
	defer statusMapMutex.Unlock()
	status, exists := statusMap[database]
	return exists && status.running
}

func setBackupRunning(database string, running bool) {
	statusMapMutex.Lock()
	defer statusMapMutex.Unlock()
	if status, exists := statusMap[database]; exists {
		status.running = running
	} else {
		statusMap[database] = &DatabaseBackupStatus{running: running}
	}
}

func backupDatabase(host string, port int, user string, password string, database string, backupDir string, compress bool, compressionLevel int, telegramNotify bool) error {
	var fileExtension string = "sql"
	if compress {
		fileExtension = "sql.gz"
	}
	// Format current time for backup file name
	backupFileName := fmt.Sprintf("%s-%s.%s", database, time.Now().Format("2006-01-02T15-04-05"), fileExtension)

	// Check if password is provided
	if password == "" {
		flag.PrintDefaults()
		log.Fatal("Database password is required")
	}

	err := os.Setenv("PGPASSWORD", password)
	if err != nil {
		return fmt.Errorf("failed to set PGPASSWORD: %v", err)
	}

	// Construct backup file path
	backupFilePath := filepath.Join(backupDir, backupFileName)

	// Create the backup file
	backupFile, err := os.Create(backupFilePath)
	if err != nil {
		return fmt.Errorf("failed to create backup file: %v", err)
	}
	defer backupFile.Close()

	if compressionLevel < -2 || compressionLevel > 9 {
		return fmt.Errorf("invalid compression level: %d", compressionLevel)
	}
	if compressionLevel == -1 {
		compressionLevel = gzip.DefaultCompression
	} else if compressionLevel == -2 {
		compressionLevel = gzip.HuffmanOnly
	} else if compressionLevel < 0 {
		compressionLevel = gzip.NoCompression
	} else if compressionLevel > 9 {
		compressionLevel = gzip.BestCompression
	} else {
		compressionLevel = gzip.BestSpeed
	}

	gzipWriter, err := gzip.NewWriterLevel(backupFile, compressionLevel)
	if err != nil {
		return fmt.Errorf("failed to create gzip writer: %v", err)
	}
	defer gzipWriter.Close()

	// Construct backup command
	cmdArgs := []string{
		"-h", host,
		"-p", strconv.Itoa(port),
		"-U", user,
		"-d", database,
	}

	// Combine commands with shell
	cmd := exec.Command("pg_dump", cmdArgs...)
	log.Printf("CMD: %v", cmd)

	// Redirect command output to backup file
	if compress {
		cmd.Stdout = gzipWriter
	} else {
		cmd.Stdout = backupFile
	}

	// Execute backup command
	err = cmd.Run()
	if err != nil {
		// In case of backup failure, remove the partially created backup file
		os.Remove(backupFilePath)
		// Send Telegram notification about the backup error
		if telegramNotify {
			if os.Getenv("TELEGRAM_BOT_TOKEN") != "" && os.Getenv("CHANNEL_ID") != "" {
				var channelID int64
				channelID, err = strconv.ParseInt(os.Getenv("CHANNEL_ID"), 10, 64)
				_ = internal.SendMessage(channelID, fmt.Sprintf("Error backing up database %s❗️", database))
			}
		}
		return fmt.Errorf("backup failed: %v", err)
	}

	log.Printf("Database %s backed up successfully. File name: %s\n", database, backupFileName)

	if telegramNotify && os.Getenv("TELEGRAM_BOT_TOKEN") != "" && os.Getenv("CHANNEL_ID") != "" {
		var channelID int64
		channelID, err = strconv.ParseInt(os.Getenv("CHANNEL_ID"), 10, 64)
		_ = internal.SendMessage(channelID, fmt.Sprintf("🎉Database %s backed up successfully.\nFile name: %s", database, backupFileName))
	}
	return nil
}

func cleanupBackups(backupDir string, maxBackupCount int) {
	// List backup files
	log.Printf("List backup files")
	files, err := filepath.Glob(filepath.Join(backupDir, "*.sql*"))
	if err != nil {
		log.Println("Error listing backup files:", err)
		return
	}

	// Sort files by modification time (oldest first)
	sortBackupFiles(files)

	// Delete excess backup files
	log.Printf("Delete excess backup files, maxBackupCount: %d", maxBackupCount)
	log.Printf("Length of files %d", len(files))
	if len(files) > maxBackupCount {
		filesToDelete := files[:len(files)-maxBackupCount]
		for _, file := range filesToDelete {
			if err := os.Remove(file); err != nil {
				log.Println("Error deleting backup file:", err)
			} else {
				log.Println("Deleted old backup file:", file)
			}
		}
	}
}

func sortBackupFiles(files []string) {
	sort.Slice(files, func(i, j int) bool {
		file1Info, _ := os.Stat(files[i])
		file2Info, _ := os.Stat(files[j])
		return file1Info.ModTime().Before(file2Info.ModTime())
	})
}

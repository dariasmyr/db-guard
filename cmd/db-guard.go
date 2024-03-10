package main

import (
	"db_dump/internal"
	"flag"
	"log"
	"os"
	"os/exec"
	"path/filepath"
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

	// Perform initial backup
	go func() {
		log.Printf("Initial backup")
		if err := internal.BackupDatabase(internal.Host, internal.Port, internal.User, internal.Password, internal.Database, internal.BackupDir, internal.Compress, internal.CompressionLevel, internal.TelegramNotify); err != nil {
			log.Println("Error backing up database:", err)
		}
	}()

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

					log.Printf("Backup")
					if err := internal.BackupDatabase(internal.Host, internal.Port, internal.User, internal.Password, internal.Database, internal.BackupDir, internal.Compress, internal.CompressionLevel, internal.TelegramNotify); err != nil {
						log.Println("Error backing up database:", err)
					}
				}()
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

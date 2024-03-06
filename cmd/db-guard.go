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
				if internal.IsBackupRunning(internal.Database) {
					log.Printf("Backup for database %s is already running, skipping this cycle", internal.Database)
					continue
				}

				// Mark backup for this database as running
				internal.SetBackupRunning(internal.Database, true)

				go func() {
					defer func() {
						// Mark backup for this database as not running after completion
						internal.SetBackupRunning(internal.Database, false)
					}()

					log.Printf("Initial backup")
					if err := internal.BackupDatabase(internal.Host, internal.Port, internal.User, internal.Password, internal.Database, internal.BackupDir, internal.Compress, internal.CompressionLevel, internal.TelegramNotify); err != nil {
						log.Println("Error backing up database:", err)
					}
				}()
				internal.CleanupBackups(internal.BackupDir, internal.MaxBackupCount)
			}
		}
	}()

	select {
	case <-backupDone:
		log.Println("Backup completed")
	}
}

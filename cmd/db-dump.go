package main

import (
	"compress/gzip"
	"db_dump/internal"
	"flag"
	"fmt"
	"io"
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
	go internal.StartBot()
	var (
		host           string
		port           int
		user           string
		password       string
		database       string
		maxBackupCount int
		intervalSec    int
		compress       bool
		backupDir      string
		telegramNotify bool
	)

	flag.StringVar(&host, "host", "localhost", "Database host")
	flag.IntVar(&port, "port", 5432, "Database port")
	flag.StringVar(&user, "user", "", "Database user")
	flag.StringVar(&password, "password", "", "Database password")
	flag.StringVar(&database, "database", "", "Database name")
	flag.IntVar(&maxBackupCount, "max-backup-count", 10, "Maximum number of backups to keep")
	flag.IntVar(&intervalSec, "interval-seconds", 60, "Interval in seconds between backups")
	flag.BoolVar(&compress, "compress", true, "Compress backups")
	flag.StringVar(&backupDir, "dir", "backups", "Backup directory")
	flag.BoolVar(&telegramNotify, "telegram-notifications", false, "Telegram notifications")

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
	if user == "" || password == "" || database == "" {
		flag.PrintDefaults()
		os.Exit(1)
	}

	// Ensure backup directory exists
	log.Println("Ensure backup directory exists")
	if _, err := os.Stat(backupDir); os.IsNotExist(err) {
		os.Mkdir(backupDir, os.ModePerm)
	}

	statusMap = make(map[string]*DatabaseBackupStatus)

	backupDone := make(chan struct{})

	// Start main routine
	go func() {
		defer close(backupDone)

		ticker := time.NewTicker(time.Duration(intervalSec) * time.Second)
		defer ticker.Stop()

		for {
			log.Printf("Starting new backup cycle")
			select {
			case <-ticker.C:
				// Check if backup for this database is already running
				if isBackupRunning(database) {
					log.Printf("Backup for database %s is already running, skipping this cycle", database)
					continue
				}

				// Mark backup for this database as running
				setBackupRunning(database, true)

				go func() {
					defer func() {
						// Mark backup for this database as not running after completion
						setBackupRunning(database, false)
					}()

					log.Printf("Initial backup")
					if err := backupDatabase(host, port, user, password, database, backupDir, compress, telegramNotify); err != nil {
						log.Println("Error backing up database:", err)
					}
				}()
				cleanupBackups(backupDir, maxBackupCount)
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

func backupDatabase(host string, port int, user string, password string, database string, backupDir string, compress bool, telegramNotify bool) error {
	// Format current time for backup file name
	backupFileName := fmt.Sprintf("%s-%s.sql", database, time.Now().Format("20060102_150405"))

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
	cmd.Stdout = backupFile

	// Execute backup command
	err = cmd.Run()
	if err != nil {
		return fmt.Errorf("backup failed: %v", err)
	}

	// If compression is enabled, compress the backup file
	if compress {
		log.Printf("Compressing backup file")
		backupFile.Close() // Close the file before compressing

		backupFile, err := os.Open(backupFilePath)
		if err != nil {
			return fmt.Errorf("failed to open backup file: %v", err)
		}
		defer backupFile.Close()

		backupFileGz, err := os.Create(backupFilePath + ".gz")
		if err != nil {
			return fmt.Errorf("failed to create compressed backup file: %v", err)
		}
		defer backupFileGz.Close()

		// Compress backup file using gzip
		backupGz := gzip.NewWriter(backupFileGz)
		defer backupGz.Close()

		_, err = io.Copy(backupGz, backupFile)
		if err != nil {
			return fmt.Errorf("failed to compress backup file: %v", err)
		}

		// Remove the uncompressed backup file
		err = os.Remove(backupFilePath)
		if err != nil {
			return fmt.Errorf("failed to remove uncompressed backup file: %v", err)
		}
	}

	log.Printf("Database %s backed up successfully to %s\n", database, backupFileName)

	if telegramNotify {
		var channelID int64
		channelID, err = strconv.ParseInt(os.Getenv("CHANNEL_ID"), 10, 64)
		_ = internal.SendMessage(channelID, fmt.Sprintf("Database %s backed up successfully to %s\n", database, backupFileName))
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

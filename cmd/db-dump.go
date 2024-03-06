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
	if password == "" {
		return fmt.Errorf("database password is required")
	}

	err := os.Setenv("PGPASSWORD", password)
	if err != nil {
		return fmt.Errorf("failed to set PGPASSWORD: %v", err)
	}

	backupFileName := generateBackupFileName(database, compress)

	backupFilePath := filepath.Join(backupDir, backupFileName)

	backupFile, err := createBackupFile(backupFilePath)
	if err != nil {
		return fmt.Errorf("failed to create backup file: %v", err)
	}
	defer backupFile.Close()

	compressionLevel = adjustCompressionLevel(compressionLevel)

	gzipWriter, err := createGzipWriter(backupFile, compressionLevel)
	if err != nil {
		return fmt.Errorf("failed to create gzip writer: %v", err)
	}
	defer gzipWriter.Close()

	cmdArgs := constructPgDumpCommandArgs(host, port, user, database)

	err = executeBackupCommand(cmdArgs, compress, gzipWriter, backupFile)
	if err != nil {
		handleBackupFailure(err, backupFilePath, database, telegramNotify)
		return fmt.Errorf("backup failed: %v", err)
	}

	handleBackupSuccess(telegramNotify, backupFilePath, database, backupFileName)

	return nil
}

func generateBackupFileName(database string, compress bool) string {
	fileExtension := "sql"
	if compress {
		fileExtension = "sql.gz"
	}
	return fmt.Sprintf("%s-%s.%s", database, time.Now().Format("2006-01-02T15-04-05"), fileExtension)
}

func createBackupFile(backupFilePath string) (*os.File, error) {
	return os.Create(backupFilePath)
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

func createGzipWriter(backupFile *os.File, compressionLevel int) (*gzip.Writer, error) {
	return gzip.NewWriterLevel(backupFile, compressionLevel)
}

func constructPgDumpCommandArgs(host string, port int, user string, database string) []string {
	return []string{"-h", host, "-p", strconv.Itoa(port), "-U", user, "-d", database}
}

func executeBackupCommand(cmdArgs []string, compress bool, gzipWriter *gzip.Writer, backupFile *os.File) error {
	cmd := exec.Command("pg_dump", cmdArgs...)
	if compress {
		cmd.Stdout = gzipWriter
	} else {
		cmd.Stdout = backupFile
	}
	log.Printf("CMD: %v", cmd)
	return cmd.Run()
}

func handleBackupFailure(err error, backupFilePath, database string, telegramNotify bool) {
	os.Remove(backupFilePath)
	if telegramNotify {
		sendTelegramNotification(fmt.Sprintf("❗️Error backing up database %s", database))
	}
}

func handleBackupSuccess(telegramNotify bool, backupFilePath, database, backupFileName string) {
	log.Printf("Database %s backed up successfully. File name: %s\n", database, backupFileName)
	if telegramNotify {
		sendTelegramNotification(fmt.Sprintf("🎉Database %s backed up successfully.\nFile name: %s", database, backupFileName))
	}
}

func sendTelegramNotification(message string) {
	if os.Getenv("TELEGRAM_BOT_TOKEN") != "" && os.Getenv("CHANNEL_ID") != "" {
		channelID, _ := strconv.ParseInt(os.Getenv("CHANNEL_ID"), 10, 64)
		_ = internal.SendMessage(channelID, message)
	}
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

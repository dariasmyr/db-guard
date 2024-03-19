package internal

import (
	"log"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

func CleanupBackups(backupDir string, maxBackupCount int) {
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
	log.Printf("Delete excess backup files (allowed: %d, now: %d files)", maxBackupCount, len(files))
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
		date1 := extractDateFromFileName(files[i])
		date2 := extractDateFromFileName(files[j])
		time1, _ := time.Parse("2006-01-02T15-04-05", date1)
		time2, _ := time.Parse("2006-01-02T15-04-05", date2)
		return time1.Before(time2)
	})
}

func extractDateFromFileName(fileName string) string {
	dateStartIndex := strings.Index(fileName, "-") + 1
	dateEndIndex := strings.Index(fileName, ".sql")
	if dateEndIndex == -1 {
		dateEndIndex = len(fileName)
	}
	return fileName[dateStartIndex:dateEndIndex]
}

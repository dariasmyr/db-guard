package internal

import (
	"db_dump/internal"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
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
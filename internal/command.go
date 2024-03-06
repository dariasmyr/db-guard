package internal

import (
	"log"
	"os"
	"os/exec"
	"strconv"
)

func ConstructPgDumpCommandArgs(host string, port int, user string, database string) []string {
	return []string{"-h", host, "-p", strconv.Itoa(port), "-U", user, "-d", database}
}

func ExecuteBackupCommand(cmdArgs []string, compress bool, gzipWriter *gzip.Writer, backupFile *os.File) error {
	cmd := exec.Command("pg_dump", cmdArgs...)
	if compress {
		cmd.Stdout = gzipWriter
	} else {
		cmd.Stdout = backupFile
	}
	log.Printf("CMD: %v", cmd)
	return cmd.Run()
}
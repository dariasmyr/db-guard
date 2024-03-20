package internal

import (
	"bytes"
	"compress/gzip"
	"fmt"
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
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	log.Printf("Executing command: %v\n", cmd.Args)

	var err error
	if compress {
		cmd.Stdout = gzipWriter
		err = cmd.Run()
	} else {
		cmd.Stdout = backupFile
		err = cmd.Run()
	}

	if err != nil {
		log.Printf("Error executing command: %v\n", err)
		log.Printf("Stdout: %s\n", stdout.String())
		log.Printf("Stderr: %s\n", stderr.String())
		return fmt.Errorf("error executing pg_dump command: %v", err)
	}

	log.Println("Command execution successful.")
	return nil
}

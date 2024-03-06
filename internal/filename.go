package internal

import (
	"fmt"
	"time"
)

func GenerateBackupFileName(database string, compress bool) string {
	fileExtension := "sql"
	if compress {
		fileExtension = "sql.gz"
	}
	return fmt.Sprintf("%s-%s.%s", database, time.Now().Format("2006-01-02T15-04-05"), fileExtension)
}
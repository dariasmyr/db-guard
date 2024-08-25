package internal

import (
	"flag"
	"fmt"
	"os"
	"strconv"
)

var (
	Host             string
	Port             int
	User             string
	Password         string
	Database         string
	MaxBackupCount   int
	IntervalSec      int
	Compress         bool
	CompressionLevel int
	BackupDir        string
	WebhookUrl       string
)

func InitParams() error {
	// Read variable from environment and if not set, read from command line
	flag.StringVar(&Host, "host", getEnv("HOST", ""), "Database host")
	flag.IntVar(&Port, "port", getEnvAsInt("PORT", 5432), "Database port")
	flag.StringVar(&User, "user", getEnv("USER", ""), "Database user")
	flag.StringVar(&Password, "password", getEnv("PASSWORD", ""), "Database password")
	flag.StringVar(&Database, "database", getEnv("DATABASE", ""), "Database name")
	flag.IntVar(&MaxBackupCount, "max-backup-count", getEnvAsInt("MAX_BACKUP_COUNT", 10), "Maximum number of backups to keep")
	flag.IntVar(&IntervalSec, "interval-seconds", getEnvAsInt("INTERVAL_SECONDS", 60), "Interval in seconds between backups")
	flag.BoolVar(&Compress, "compress", getEnvAsBool("COMPRESS", true), "Compress backups")
	flag.IntVar(&CompressionLevel, "compression-level", getEnvAsInt("COMPRESSION_LEVEL", -1), "Compression level")
	flag.StringVar(&BackupDir, "dir", getEnv("BACKUP_DIR", "backups"), "Backup directory")
	flag.StringVar(&WebhookUrl, "webhook-url", getEnv("WEBHOOK_URL", ""), "Webhook URL")

	flag.Parse()

	// Check parameter validity
	if MaxBackupCount <= 0 || MaxBackupCount > 100 {
		return fmt.Errorf("invalid value for MaxBackupCount: %d (must be greater than 0 and less than 100)", MaxBackupCount)
	}
	if IntervalSec <= 0 {
		return fmt.Errorf("invalid value for IntervalSec: %d (must be greater than 0)", IntervalSec)
	}
	if CompressionLevel < -1 || CompressionLevel > 9 {
		return fmt.Errorf("invalid value for CompressionLevel: %d (must be between -1 and 9)", CompressionLevel)
	}

	return nil
}

func getEnv(key string, defaultValue string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return defaultValue
}

func getEnvAsInt(key string, defaultValue int) int {
	if value, exists := os.LookupEnv(key); exists {
		intValue, err := strconv.Atoi(value)
		if err == nil {
			return intValue
		}
	}
	return defaultValue
}

func getEnvAsBool(key string, defaultValue bool) bool {
	if value, exists := os.LookupEnv(key); exists {
		boolValue, err := strconv.ParseBool(value)
		if err == nil {
			return boolValue
		}
	}
	return defaultValue
}

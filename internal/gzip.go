package internal

func CreateGzipWriter(backupFile *os.File, compressionLevel int) (*gzip.Writer, error) {
	return gzip.NewWriterLevel(backupFile, compressionLevel)
}

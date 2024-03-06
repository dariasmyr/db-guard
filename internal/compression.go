package internal

import (
	"compress/gzip"
)

func AdjustCompressionLevel(compressionLevel int) int {
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

package utils

import (
	"crypto/rand"
	"encoding/hex"
	"hash/fnv"
	"sync/atomic"
	"time"
)

var requestCounter uint64

func Hash(s string) uint32 {
	h := fnv.New32a()
	h.Write([]byte(s))
	return h.Sum32()
}

func GenerateRequestID() string {
	timestamp := time.Now().UnixNano() / int64(time.Millisecond)

	// Get random part (8 bytes)
	randomBytes := make([]byte, 4)
	_, err := rand.Read(randomBytes)
	if err != nil {
		// If random generation fails, use timestamp as fallback
		randomBytes = []byte{
			byte(timestamp >> 24),
			byte(timestamp >> 16),
			byte(timestamp >> 8),
			byte(timestamp),
		}
	}

	// Get counter (atomic increment)
	counter := atomic.AddUint64(&requestCounter, 1)

	// Format request ID
	return hex.EncodeToString([]byte{
		byte(timestamp >> 24),
		byte(timestamp >> 16),
		byte(timestamp >> 8),
		byte(timestamp),
	}) + "-" + hex.EncodeToString(randomBytes) + "-" +
		hex.EncodeToString([]byte{
			byte(counter >> 24),
			byte(counter >> 16),
			byte(counter >> 8),
			byte(counter),
		})
}

// IsPathSafe checks if the given path is safe (doesn't contain traversal attempts)
func IsPathSafe(path string) bool {
	return !(path == ".." ||
		path == "." ||
		path == "/" ||
		path == "" ||
		path == "\\" ||
		path == "~" ||
		path == "*" ||
		path == "|" ||
		path == ">" ||
		path == "<")
}

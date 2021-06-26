package client

import (
	"crypto/sha256"
	"fmt"
	"os"
	"regexp"
	"strings"
)

// CreateDirIfNotExists will try create a directry if it does not exist
func CreateDirIfNotExists(path string) error {
	if isDir(path) {
		return nil
	}

	return os.MkdirAll(path, 0750)
}

// Check path is a directory
func isDir(path string) bool {
	info, err := os.Stat(path)
	if err != nil || os.IsNotExist(err) {
		return false
	}

	return info.IsDir()
}

// Check path is a file
func isFile(path string) bool {
	info, err := os.Stat(path)
	if err != nil || os.IsNotExist(err) {
		return false
	}

	return !info.IsDir()
}

// hashString will return a sha256 hash of a string
func hashString(s string) string {
	return fmt.Sprintf("%x", sha256.Sum256([]byte(s)))[0:8]
}

// inArray is a Golang clone of in_array
func inArray(x string, a []string) bool {
	for _, n := range a {
		if x == n {
			return true
		}
	}
	return false
}

// inRegexpArray is a Golang clone of in_array
func inRegexpArray(x string, a []*regexp.Regexp) bool {
	for _, n := range a {
		if n.MatchString(x) {
			return true
		}
	}
	return false
}

// IsCompressed returns whether a file is compressed simply based on filename
func isCompressed(file string) bool {
	return strings.HasSuffix(file, ".zst")
}

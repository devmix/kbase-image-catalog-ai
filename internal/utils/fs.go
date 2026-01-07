package utils

import (
	"errors"
	"os"
)

func IsDirectory(path string) bool {
	fileInfo, err := os.Stat(path)
	if err != nil {
		return false
	}
	return fileInfo.IsDir()
}

func IsFileExists(filename string) bool {
	fileInfo, err := os.Stat(filename)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return false
		}
		// For all other errors (like permission errors), we still return false
		// because the file cannot be determined to exist
		return false
	}

	// If we got here, the file exists. Check if it's a directory.
	// If it's a directory, return false since we only want to identify files
	return !fileInfo.IsDir()
}

package utils

import (
	"os"
)

// IsDirExist checks if a path is an existed dir
func IsDirExist(path string) bool {
	fi, err := os.Stat(path)

	if err != nil {
		return os.IsExist(err)
	}

	return fi.IsDir()
}

// IsFileExist checks if a file url is an exist file
func IsFileExist(filename string) bool {
	_, err := os.Stat(filename)
	return err == nil || os.IsExist(err)
}

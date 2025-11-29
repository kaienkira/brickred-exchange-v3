package main

import (
	"os"
)

func utilCheckFileExists(filePath string) bool {
	info, err := os.Stat(filePath)
	if err != nil {
		return false
	}
	if info.IsDir() {
		return false
	}

	return true
}

func utilCheckDirExists(dirPath string) bool {
	info, err := os.Stat(dirPath)
	if err != nil {
		return false
	}
	if info.IsDir() == false {
		return false
	}

	return true
}

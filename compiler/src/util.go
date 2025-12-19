package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

func utilAtoi(str string) int {
	i, err := strconv.Atoi(str)
	if err != nil {
		return 0
	} else {
		return i
	}
}

func utilGetFileNameWithoutExtension(filePath string) string {
	base := filepath.Base(filePath)
	ext := filepath.Ext(base)
	if ext == "" {
		return base
	} else {
		return strings.TrimSuffix(base, ext)
	}
}

func utilGetFullPath(filePath string) string {
	fullPath, err := filepath.Abs(filePath)
	if err != nil {
		return ""
	} else {
		return fullPath
	}
}

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

func utilWriteAllText(filePath string, fileContent string) bool {
	err := os.WriteFile(filePath, []byte(fileContent), 0644)
	if err != nil {
		fmt.Fprintf(os.Stderr,
			"error: write file %s failed: %s",
			filePath, err.Error())
		return false
	}

	return true
}

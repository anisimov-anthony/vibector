package main

import (
	"fmt"
	"path/filepath"
	"strings"
)

func detectFormatFromExtension(filePath string) (string, error) {
	ext := strings.ToLower(filepath.Ext(filePath))

	switch ext {
	case ".json":
		return "json", nil
	case ".txt", ".text":
		return "text", nil
	case "":
		return "text", nil
	default:
		return "", fmt.Errorf("unsupported file extension: %s", ext)
	}
}

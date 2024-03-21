package filesystem

import (
	"os"
	"path/filepath"
	"strings"
)

func Walk(dirPath string) []string {
	var files []string

	getFiles := func(path string, fileInfo os.FileInfo, errno error) (err error) {
		if !fileInfo.IsDir() {
			files = append(files, strings.ReplaceAll(path, "\\", "/"))
		}
		return nil
	}

	filepath.Walk(dirPath, getFiles)

	return files
}

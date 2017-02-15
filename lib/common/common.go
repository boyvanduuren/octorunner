package common

import (
	"os"
	"path/filepath"
)

func CheckDirNotExists(dir string) bool {
	s, err := os.Stat(dir)
	return os.IsNotExist(err) == true || !s.IsDir()
}

/*
Find all the files in a given directory
 */
func FindAllFiles(path string) []string {
	collectedFiles := []string{}

	if CheckDirNotExists(path) {
		return collectedFiles
	}

	collect := func(path string, f os.FileInfo, err error) error {
		if !f.IsDir() {
			collectedFiles = append(collectedFiles, path)
		}
		return nil
	}
	filepath.Walk(path, collect)
	return collectedFiles
}

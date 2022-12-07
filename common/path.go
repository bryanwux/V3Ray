package common

import (
	"log"
	"os"
	"path/filepath"
)

func GetPath(fileName string) string {
	paths := make([]string, 0)
	if execFile, err := os.Executable(); err == nil {
		paths = append(paths, filepath.Join(filepath.Dir(execFile), fileName))
	}

	for _, p := range paths {
		if _, err := os.Stat(p); err == nil {
			log.Printf("using %v\n", p)
			return p
		}
	}
	return ""
}

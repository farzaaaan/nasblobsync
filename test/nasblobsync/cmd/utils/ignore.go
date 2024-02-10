package utils

import "path/filepath"

var ignorePatterns = []string{"*.thump.png"}

func ShouldIgnoreFile(path string) bool {
	for _, pattern := range ignorePatterns {
		matched, err := filepath.Match(pattern, filepath.Base(path))
		if err == nil && matched {
			return true
		}
	}
	return false
}

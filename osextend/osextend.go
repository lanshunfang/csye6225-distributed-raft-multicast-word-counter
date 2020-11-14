package osextend

import "os"

// FileExists ...
// Check if a file path exist and must be a file, not dir
func FileExists(filepath string) bool {
	info, err := os.Stat(filepath)
	if os.IsNotExist(err) {
		return false
	}
	return !info.IsDir()
}

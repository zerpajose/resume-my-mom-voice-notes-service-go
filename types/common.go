package types

import "path/filepath"

func GetFileExtension(filename string) string {
	return filepath.Ext(filename)[1:] // removes the dot
}

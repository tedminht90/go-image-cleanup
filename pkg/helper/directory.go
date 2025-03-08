package helper

import (
	"os"
	"path/filepath"
)

// EnsureDirectoryExists kiểm tra nếu thư mục tồn tại, nếu không thì tạo mới
// Trả về lỗi nếu có
func EnsureDirectoryExists(dirPath string) error {
	// Kiểm tra nếu thư mục đã tồn tại
	if _, err := os.Stat(dirPath); os.IsNotExist(err) {
		// Tạo thư mục với quyền 0755 (rwxr-xr-x)
		if err := os.MkdirAll(dirPath, 0755); err != nil {
			return err
		}
	}
	return nil
}

// GetParentDirectory trả về thư mục cha của một đường dẫn file
func GetParentDirectory(filePath string) string {
	return filepath.Dir(filePath)
}

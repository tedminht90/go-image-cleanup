// pkg/helper/mask.go
package helper

// MaskValue masks sensitive information
func MaskValue(value string) string {
	if len(value) <= 8 {
		return "********"
	}
	return value[:4] + "..." + value[len(value)-4:]
}

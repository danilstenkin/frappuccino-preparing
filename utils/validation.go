package utils

import "strings"

// Проверяет, что размер валиден (он входит в допустимый список).
func IsValidSize(validSizes []string, size string) bool {
	for _, s := range validSizes {
		if strings.ToLower(s) == strings.ToLower(size) {
			return true
		}
	}
	return false
}

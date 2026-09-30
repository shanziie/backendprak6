package helper

import (
	"strings"
	"unicode"
)

func passwordStrength(pwd string) string {
	if len(pwd) < 8 {
		return "minimal 8 karakter"
	}
	if len(pwd) > 72 {
		return "maksimal 72 karakter"
	}

	hasLetter := false
	hasDigit := false

	for _, r := range pwd {
		if unicode.IsLetter(r) {
			hasLetter = true
		}
		if unicode.IsDigit(r) {
			hasDigit = true
		}
	}

	if !hasLetter || !hasDigit {
		return "harus memuat huruf dan angka"
	}

	commonPasswords := []string{"password123", "12345678", "qwertyuiop", "admin123"}
	for _, cp := range commonPasswords {
		if strings.EqualFold(pwd, cp) {
			return "password terlalu umum"
		}
	}

	return ""
}
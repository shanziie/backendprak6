package service

import "testing"

func TestCheckPasswordStrength(t *testing.T) {
	tests := []struct {
		name     string
		password string
		want     string
	}{
		{"Terlalu Pendek", "abc12", "minimal 8 karakter"},
		{"Hanya Huruf", "abcdefgh", "harus memuat huruf dan angka"},
		{"Hanya Angka", "12345678", "harus memuat huruf dan angka"},
		{"Valid", "rahasia123", ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := CheckPasswordStrength(tt.password); got != tt.want {
				t.Errorf("CheckPasswordStrength() = %v, want %v", got, tt.want)
			}
		})
	}
}
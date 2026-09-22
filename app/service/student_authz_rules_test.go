package service

import (
	"testing"

	"api-students/app/model"
	"api-students/helper"
)

func TestCanAccessStudent(t *testing.T) {
	perms := helper.NewPermissionSet(map[string][]string{
		"admin": {"student:read:any", "student:update:any"},
		"staff": {"student:read:any"},
		"user":  {},
	})

	tests := []struct {
		name       string
		user       model.AuthUser
		ownerID    int
		permission string
		want       bool
	}{
		{
			name:       "owner access own data",
			user:       model.AuthUser{UserID: 10, Role: "user"},
			ownerID:    10,
			permission: "student:read:any",
			want:       true,
		},
		{
			name:       "user access other data",
			user:       model.AuthUser{UserID: 10, Role: "user"},
			ownerID:    20,
			permission: "student:read:any",
			want:       false,
		},
		{
			name:       "staff read other data",
			user:       model.AuthUser{UserID: 88, Role: "staff"},
			ownerID:    20,
			permission: "student:read:any",
			want:       true,
		},
		{
			name:       "staff update other data forbidden",
			user:       model.AuthUser{UserID: 88, Role: "staff"},
			ownerID:    20,
			permission: "student:update:any",
			want:       false,
		},
		{
			name:       "admin update other data allowed",
			user:       model.AuthUser{UserID: 99, Role: "admin"},
			ownerID:    20,
			permission: "student:update:any",
			want:       true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := CanAccessStudent(tt.user, tt.ownerID, perms, tt.permission)
			if got != tt.want {
				t.Errorf("got %v, want %v", got, tt.want)
			}
		})
	}
}
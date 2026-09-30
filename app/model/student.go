package model

import (
	"time"
)

type Student struct {
	ID        int       `json:"id"`
	Nim       string    `json:"nim"`
	Name      string    `json:"name"`
	Grade     float64   `json:"grade"`
	OwnerID   int       `json:"owner_id"`
	CreatedAt time.Time `json:"created_at"`
}

type CreateStudentRequest struct {
	Nim   string  `json:"nim"   validate:"required,min=8,max=15,alphanum"`
	Name  string  `json:"name"  validate:"required,min=3,max=100"`
	Grade float64 `json:"grade" validate:"required,min=0,max=100"`
}

type PatchStudentRequest struct {
	Name  *string  `json:"name,omitempty"  validate:"omitnil,min=3,max=100"`
	Grade *float64 `json:"grade,omitempty" validate:"omitnil,min=0,max=100"`
}
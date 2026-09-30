package model

import "time"

type WebResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
	Data    any    `json:"data,omitempty"`
}

type CursorQuery struct {
	Limit    int
	Search   string
	IsActive *bool
	After    *Cursor
}

type Cursor struct {
	CreatedAt time.Time
	ID        int
}
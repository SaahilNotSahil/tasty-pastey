package models

import (
	"time"

	"github.com/google/uuid"
)

type User struct {
	ID        uuid.UUID `json:"id" gorm:"primaryKey"`
	Email     string    `json:"email" binding:"required" gorm:"unique"`
	Password  string    `json:"password" binding:"required"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type Paste struct {
	ID         string    `json:"id" gorm:"primaryKey"`
	Title      string    `json:"title"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
	Visibility int       `json:"visibility"`
	UserID     uuid.UUID `json:"user_id"`
}

type PasteAccess struct {
	ID        uuid.UUID `json:"id" gorm:"primaryKey"`
	PasteID   string    `json:"paste_id"`
	UserID    uuid.UUID `json:"user_id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

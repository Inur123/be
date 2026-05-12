package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type User struct {
	ID        uuid.UUID  `gorm:"type:uuid;primaryKey" json:"id"`
	Name      string     `gorm:"not null" json:"name" binding:"required"`
	Email     string     `gorm:"unique;not null" json:"email" binding:"required,email"`
	Password  string     `gorm:"not null" json:"-"` // Password tidak dikembalikan di response JSON
	RoleID    *uuid.UUID `gorm:"type:uuid;index" json:"role_id"` // Kunci tamu opsional ke Role
	Role      Role       `gorm:"foreignKey:RoleID" json:"role"`
	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt time.Time  `json:"updated_at"`
}

// Hook GORM untuk mengisi UUID secara otomatis sebelum baris disimpan
func (u *User) BeforeCreate(tx *gorm.DB) (err error) {
	if u.ID == uuid.Nil {
		u.ID = uuid.New()
	}
	return
}

type RegisterInput struct {
	Name     string `json:"name" binding:"required"`
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=6"`
}

type LoginInput struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

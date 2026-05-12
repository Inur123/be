package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Role struct {
	ID          uuid.UUID    `gorm:"type:uuid;primaryKey" json:"id"`
	Name        string       `gorm:"unique;not null" json:"name" binding:"required"` // Contoh: "Superadmin", "Admin Cabang"
	Description string       `json:"description"`
	IsDefault   bool         `gorm:"default:false" json:"is_default"` // Penanda peran default untuk pendaftar baru
	Permissions []Permission `gorm:"many2many:role_permissions;" json:"permissions"` // Relasi Many-to-Many
	CreatedAt   time.Time    `json:"created_at"`
	UpdatedAt   time.Time    `json:"updated_at"`
}

// Hook GORM untuk mengisi UUID secara otomatis sebelum baris disimpan
func (r *Role) BeforeCreate(tx *gorm.DB) (err error) {
	if r.ID == uuid.Nil {
		r.ID = uuid.New()
	}
	return
}

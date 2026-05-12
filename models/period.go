package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// Period merepresentasikan masa khidmat / kepengurusan organisasi milik pengguna
type Period struct {
	ID        uuid.UUID `gorm:"type:uuid;primaryKey" json:"id"`
	UserID    uuid.UUID `gorm:"type:uuid;uniqueIndex:idx_user_period_name;not null" json:"user_id"` // Pemilik periode
	Name      string    `gorm:"uniqueIndex:idx_user_period_name;not null" json:"name" binding:"required"` // Contoh: "2025-2026"
	IsActive  bool      `gorm:"default:false;index" json:"is_active"` // Penanda periode yang sedang berjalan
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// Hook GORM untuk mengisi UUID secara otomatis sebelum entitas disimpan
func (p *Period) BeforeCreate(tx *gorm.DB) (err error) {
	if p.ID == uuid.Nil {
		p.ID = uuid.New()
	}
	return
}

// Skema validasi masukan untuk pembuatan masa kepengurusan baru
type CreatePeriodInput struct {
	Name     string `json:"name" binding:"required"`
	IsActive bool   `json:"is_active"`
}

// Skema validasi masukan untuk pemutakhiran masa kepengurusan
type UpdatePeriodInput struct {
	Name string `json:"name" binding:"required"`
}

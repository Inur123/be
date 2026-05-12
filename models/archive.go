package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// Archive merepresentasikan pencatatan/arsip surat organisasi yang terikat pada satu masa kepengurusan
type Archive struct {
	ID           uuid.UUID `gorm:"type:uuid;primaryKey" json:"id"`
	UserID       uuid.UUID `gorm:"type:uuid;index;not null" json:"user_id"`   // Pemilik arsip
	PeriodID     uuid.UUID `gorm:"type:uuid;index;not null" json:"period_id"` // Terikat otomatis pada periode aktif saat itu
	LetterNumber string    `gorm:"not null" json:"letter_number" binding:"required"` // Nomer Surat
	Organization string    `gorm:"not null" json:"organization" binding:"required"`  // IPNU, IPPNU, BERSAMA, CBP-KPP
	ContactName  string    `gorm:"not null" json:"contact_name" binding:"required"`  // Nama Pengirim / Penerima
	Description  *string   `json:"description"`                                      // Deskripsi opsional (Nullable)
	LetterType   string    `gorm:"not null;index" json:"letter_type" binding:"required"` // Jenis Surat: "Masuk" atau "Keluar"
	LetterDate   time.Time `gorm:"not null" json:"letter_date"`                      // Tanggal Surat
	Subject      string    `gorm:"not null" json:"subject" binding:"required"`       // Perihal utama
	FileUrl      *string   `json:"file_url"`                                         // Tautan dokumen digital (Nullable)
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

// Hook GORM untuk mengisi UUID secara otomatis
func (a *Archive) BeforeCreate(tx *gorm.DB) (err error) {
	if a.ID == uuid.Nil {
		a.ID = uuid.New()
	}
	return
}

// Skema masukan untuk penambahan arsip surat baru
type CreateArchiveInput struct {
	LetterNumber string  `json:"letter_number" binding:"required"`
	Organization string  `json:"organization" binding:"required"`
	ContactName  string  `json:"contact_name" binding:"required"`
	Description  *string `json:"description"`
	LetterType   string  `json:"letter_type" binding:"required"`
	LetterDate   *string `json:"letter_date"` // Format string tanggal, jika kosong akan diisi tanggal sekarang di controller
	Subject      string  `json:"subject" binding:"required"`
	FileUrl      *string `json:"file_url"`
}

// Skema masukan untuk pemutakhiran arsip surat
type UpdateArchiveInput struct {
	LetterNumber string  `json:"letter_number" binding:"required"`
	Organization string  `json:"organization" binding:"required"`
	ContactName  string  `json:"contact_name" binding:"required"`
	Description  *string `json:"description"`
	LetterType   string  `json:"letter_type" binding:"required"`
	LetterDate   *string `json:"letter_date"`
	Subject      string  `json:"subject" binding:"required"`
	FileUrl      *string `json:"file_url"`
}

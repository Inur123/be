package controllers

import (
	"net/http"
	"time"

	"laci-cabang-be/config"
	"laci-cabang-be/models"

	"github.com/gin-gonic/gin"
)

// CreateArchive menyimpan berkas surat baru dengan auto-tagging periode aktif
func CreateArchive(c *gin.Context) {
	userID, ok := getUserIDFromContext(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Tidak terautentikasi"})
		return
	}

	var input models.CreateArchiveInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// LAPIS PENGAMAN GATEKEEPER: Pastikan entitas memiliki masa kepengurusan yang aktif berjalan
	var activePeriod models.Period
	if err := config.DB.Where("user_id = ? AND is_active = ?", userID, true).First(&activePeriod).Error; err != nil {
		c.JSON(http.StatusForbidden, gin.H{
			"error": "Akses Ditolak: Kepengurusan Anda belum memiliki Masa Periode Aktif. Silakan aktifkan terlebih dahulu di menu Periodisasi.",
		})
		return
	}

	// Tentukan Tanggal Surat: Jika tidak diisi, otomatis gunakan waktu saat ini
	letterDate := time.Now()
	if input.LetterDate != nil && *input.LetterDate != "" {
		// Coba parsing format tanggal umum (RFC3339 atau YYYY-MM-DD)
		if parsed, err := time.Parse(time.RFC3339, *input.LetterDate); err == nil {
			letterDate = parsed
		} else if parsed, err := time.Parse("2006-01-02", *input.LetterDate); err == nil {
			letterDate = parsed
		}
	}

	// AUTO-INJECT: Suntikkan UserID dan PeriodID aktif secara otomatis
	newArchive := models.Archive{
		UserID:       userID,
		PeriodID:     activePeriod.ID,
		LetterNumber: input.LetterNumber,
		Organization: input.Organization,
		ContactName:  input.ContactName,
		Description:  input.Description,
		LetterType:   input.LetterType,
		LetterDate:   letterDate,
		Subject:      input.Subject,
		FileUrl:      input.FileUrl,
	}

	if err := config.DB.Create(&newArchive).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal menyimpan arsip surat"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "Arsip surat berhasil ditambahkan ke periode aktif",
		"archive": newArchive,
	})
}

// GetArchives mengambil daftar dokumen arsip milik pribadi (mendukung filter opsional)
func GetArchives(c *gin.Context) {
	userID, ok := getUserIDFromContext(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Tidak terautentikasi"})
		return
	}

	periodID := c.Query("period_id")
	letterType := c.Query("letter_type")
	organization := c.Query("organization")

	var archives []models.Archive
	query := config.DB.Where("user_id = ?", userID).Order("letter_date DESC, created_at DESC")

	if periodID != "" {
		query = query.Where("period_id = ?", periodID)
	}
	if letterType != "" {
		query = query.Where("letter_type = ?", letterType)
	}
	if organization != "" {
		query = query.Where("organization = ?", organization)
	}

	if err := query.Find(&archives).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal mengambil daftar arsip surat"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"archives": archives,
	})
}

// GetArchiveById mengambil rincian satu surat secara spesifik
func GetArchiveById(c *gin.Context) {
	userID, ok := getUserIDFromContext(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Tidak terautentikasi"})
		return
	}

	archiveID := c.Param("id")
	var archive models.Archive

	if err := config.DB.Where("id = ? AND user_id = ?", archiveID, userID).First(&archive).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Arsip surat tidak ditemukan atau bukan milik Anda"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"archive": archive,
	})
}

// UpdateArchive memutakhirkan informasi arsip surat pribadi
func UpdateArchive(c *gin.Context) {
	userID, ok := getUserIDFromContext(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Tidak terautentikasi"})
		return
	}

	archiveID := c.Param("id")
	var archive models.Archive

	if err := config.DB.Where("id = ? AND user_id = ?", archiveID, userID).First(&archive).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Arsip surat tidak ditemukan"})
		return
	}

	var input models.UpdateArchiveInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Memperbarui atribut inti
	archive.LetterNumber = input.LetterNumber
	archive.Organization = input.Organization
	archive.ContactName = input.ContactName
	archive.Description = input.Description
	archive.LetterType = input.LetterType
	archive.Subject = input.Subject
	archive.FileUrl = input.FileUrl

	if input.LetterDate != nil && *input.LetterDate != "" {
		if parsed, err := time.Parse(time.RFC3339, *input.LetterDate); err == nil {
			archive.LetterDate = parsed
		} else if parsed, err := time.Parse("2006-01-02", *input.LetterDate); err == nil {
			archive.LetterDate = parsed
		}
	}

	if err := config.DB.Save(&archive).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal menyimpan pemutakhiran arsip"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Arsip surat berhasil diperbarui",
		"archive": archive,
	})
}

// DeleteArchive menghapus arsip surat pribadi secara permanen
func DeleteArchive(c *gin.Context) {
	userID, ok := getUserIDFromContext(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Tidak terautentikasi"})
		return
	}

	archiveID := c.Param("id")
	var archive models.Archive

	if err := config.DB.Where("id = ? AND user_id = ?", archiveID, userID).First(&archive).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Arsip surat tidak ditemukan"})
		return
	}

	if err := config.DB.Delete(&archive).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal menghapus arsip surat"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Arsip surat berhasil dihapus",
	})
}

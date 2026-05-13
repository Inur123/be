package controllers

import (
	"net/http"

	"laci-cabang-be/config"
	"laci-cabang-be/models"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// helper untuk mengekstrak identitas pemilik dari token
func getUserIDFromContext(c *gin.Context) (uuid.UUID, bool) {
	userIDStr, exists := c.Get("user_id")
	if !exists {
		return uuid.Nil, false
	}
	parsed, err := uuid.Parse(userIDStr.(string))
	if err != nil {
		return uuid.Nil, false
	}
	return parsed, true
}

// CreatePeriod membuat masa khidmat/kepengurusan baru eksklusif milik pengguna yang login
func CreatePeriod(c *gin.Context) {
	userID, ok := getUserIDFromContext(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Akses ditolak: Klaim identitas pengguna tidak ditemukan"})
		return
	}

	var input models.CreatePeriodInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Periksa apakah nama periode sudah ada untuk pengguna ini secara spesifik
	var existing models.Period
	if err := config.DB.Where("name = ? AND user_id = ?", input.Name, userID).First(&existing).Error; err == nil {
		c.JSON(http.StatusConflict, gin.H{"error": "Nama masa kepengurusan tersebut telah Anda buat sebelumnya"})
		return
	}

	// LOGIKA CERDAS 1: Jika belum ada periode sama sekali milik pengguna ini, otomatis jadikan aktif!
	var count int64
	config.DB.Model(&models.Period{}).Where("user_id = ?", userID).Count(&count)
	if count == 0 {
		input.IsActive = true
	}

	// LOGIKA CERDAS 2: Jika diaktifkan, nonaktifkan seluruh periode lain milik pengguna ini (Auto-Demisioner Pribadi)
	if input.IsActive {
		config.DB.Model(&models.Period{}).Where("user_id = ?", userID).Update("is_active", false)
	}

	newPeriod := models.Period{
		UserID:   userID,
		Name:     input.Name,
		Scope:    input.Scope,
		IsActive: input.IsActive,
	}

	if input.Scope == "" {
		newPeriod.Scope = "Cabang"
	}

	if err := config.DB.Create(&newPeriod).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal menyimpan masa kepengurusan baru"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "Masa kepengurusan berhasil dibuat",
		"period":  newPeriod,
	})
}

// GetPeriods mengambil daftar masa kepengurusan murni milik pribadi yang login
func GetPeriods(c *gin.Context) {
	userID, ok := getUserIDFromContext(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Tidak terautentikasi"})
		return
	}

	scope := c.Query("scope") // Contoh: ?scope=Cabang
	var periods []models.Period

	query := config.DB.Where("user_id = ?", userID).Order("created_at DESC")
	if scope != "" {
		query = query.Where("scope = ?", scope)
	}

	if err := query.Find(&periods).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal mengambil daftar masa kepengurusan"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"periods": periods,
	})
}

// GetActivePeriod mengambil data periode aktif murni milik pribadi
func GetActivePeriod(c *gin.Context) {
	userID, ok := getUserIDFromContext(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Tidak terautentikasi"})
		return
	}

	var period models.Period
	if err := config.DB.Where("user_id = ? AND is_active = ?", userID, true).First(&period).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Anda belum memiliki masa kepengurusan aktif"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"period": period,
	})
}

// SetActivePeriod mengalihkan status aktif secara eksklusif dalam ruang lingkup pribadi
func SetActivePeriod(c *gin.Context) {
	userID, ok := getUserIDFromContext(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Tidak terautentikasi"})
		return
	}

	periodID := c.Param("id")

	var targetPeriod models.Period
	if err := config.DB.Where("id = ? AND user_id = ?", periodID, userID).First(&targetPeriod).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Masa kepengurusan tidak ditemukan atau bukan milik Anda"})
		return
	}

	// Demisionerkan seluruh periode lain milik pengguna ini
	config.DB.Model(&models.Period{}).Where("user_id = ?", userID).Update("is_active", false)

	// Aktifkan periode target
	if err := config.DB.Model(&targetPeriod).Update("is_active", true).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal mengaktifkan masa kepengurusan"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Masa kepengurusan aktif berhasil dialihkan",
		"period":  targetPeriod,
	})
}

// UpdatePeriod memutakhirkan nama/informasi dari masa kepengurusan milik pribadi
func UpdatePeriod(c *gin.Context) {
	userID, ok := getUserIDFromContext(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Tidak terautentikasi"})
		return
	}

	periodID := c.Param("id")

	var period models.Period
	if err := config.DB.Where("id = ? AND user_id = ?", periodID, userID).First(&period).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Masa kepengurusan tidak ditemukan atau bukan milik Anda"})
		return
	}

	var input models.UpdatePeriodInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Jika nama diubah, pastikan tidak duplikat dengan milik pribadi yang lain
	if input.Name != period.Name {
		var duplicate models.Period
		if err := config.DB.Where("name = ? AND user_id = ? AND id != ?", input.Name, userID, period.ID).First(&duplicate).Error; err == nil {
			c.JSON(http.StatusConflict, gin.H{"error": "Nama masa kepengurusan tersebut telah Anda gunakan"})
			return
		}
	}

	period.Name = input.Name
	if err := config.DB.Save(&period).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal menyimpan pembaruan masa kepengurusan"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Data masa kepengurusan berhasil diperbarui",
		"period":  period,
	})
}

// DeletePeriod menghapus entitas masa kepengurusan milik pribadi dengan proteksi ketat
func DeletePeriod(c *gin.Context) {
	userID, ok := getUserIDFromContext(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Tidak terautentikasi"})
		return
	}

	periodID := c.Param("id")

	var period models.Period
	if err := config.DB.Where("id = ? AND user_id = ?", periodID, userID).First(&period).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Masa kepengurusan tidak ditemukan atau bukan milik Anda"})
		return
	}

	// LOGIKA CERDAS 3: Proteksi mutlak
	if period.IsActive {
		c.JSON(http.StatusForbidden, gin.H{
			"error": "Masa kepengurusan yang sedang aktif berjalan tidak dapat dihapus. Silakan alihkan status aktif ke periode lain terlebih dahulu.",
		})
		return
	}

	if err := config.DB.Delete(&period).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal menghapus masa kepengurusan"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Masa kepengurusan berhasil dihapus secara permanen",
	})
}

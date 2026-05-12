package controllers

import (
	"net/http"

	"laci-cabang-be/config"
	"laci-cabang-be/models"

	"github.com/gin-gonic/gin"
)

// CreateRoleInput mendefinisikan skema payload pembuatan peran baru beserta susunan hak aksesnya
type CreateRoleInput struct {
	Name          string   `json:"name" binding:"required"`
	Description   string   `json:"description"`
	PermissionIDs []string `json:"permission_ids"` // Array string UUID hak akses
}

// CreateRole membuat peran baru (Eksklusif Superadmin)
func CreateRole(c *gin.Context) {
	var input CreateRoleInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Pastikan nama peran belum digunakan
	var existingRole models.Role
	if err := config.DB.Where("name = ?", input.Name).First(&existingRole).Error; err == nil {
		c.JSON(http.StatusConflict, gin.H{"error": "Nama peran sudah digunakan"})
		return
	}

	newRole := models.Role{
		Name:        input.Name,
		Description: input.Description,
	}

	// Simpan peran terlebih dahulu untuk mendapatkan UUID
	if err := config.DB.Create(&newRole).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal membuat peran baru"})
		return
	}

	// Tautkan daftar hak akses jika disertakan
	if len(input.PermissionIDs) > 0 {
		var permissions []models.Permission
		if err := config.DB.Where("id IN ?", input.PermissionIDs).Find(&permissions).Error; err == nil && len(permissions) > 0 {
			// Tambahkan asosiasi Many-to-Many
			_ = config.DB.Model(&newRole).Association("Permissions").Append(permissions)
		}
	}

	// Muat ulang peran beserta hak aksesnya untuk respons
	config.DB.Preload("Permissions").First(&newRole, "id = ?", newRole.ID)

	c.JSON(http.StatusCreated, gin.H{
		"message": "Peran berhasil dibuat",
		"role":    newRole,
	})
}

// GetRoles mengambil seluruh daftar peran beserta rincian hak aksesnya
func GetRoles(c *gin.Context) {
	var roles []models.Role
	if err := config.DB.Preload("Permissions").Find(&roles).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal mengambil daftar peran"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"roles": roles,
	})
}

// CreatePermissionInput mendefinisikan payload pembuatan butir hak akses
type CreatePermissionInput struct {
	Name        string `json:"name" binding:"required"`
	Description string `json:"description"`
}

// CreatePermission membuat butiran hak akses baru
func CreatePermission(c *gin.Context) {
	var input CreatePermissionInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var existingPerm models.Permission
	if err := config.DB.Where("name = ?", input.Name).First(&existingPerm).Error; err == nil {
		c.JSON(http.StatusConflict, gin.H{"error": "Hak akses dengan nama ini sudah ada"})
		return
	}

	newPerm := models.Permission{
		Name:        input.Name,
		Description: input.Description,
	}

	if err := config.DB.Create(&newPerm).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal membuat hak akses"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message":    "Hak akses berhasil dibuat",
		"permission": newPerm,
	})
}

// GetPermissions mengambil seluruh referensi butir hak akses
func GetPermissions(c *gin.Context) {
	var permissions []models.Permission
	if err := config.DB.Find(&permissions).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal mengambil daftar hak akses"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"permissions": permissions,
	})
}

// SetDefaultRole menetapkan peran tertentu sebagai penugasan otomatis untuk pendaftar baru
func SetDefaultRole(c *gin.Context) {
	roleID := c.Param("id")

	// Pastikan peran tersebut eksis
	var role models.Role
	if err := config.DB.First(&role, "id = ?", roleID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Peran tidak ditemukan"})
		return
	}

	if role.Name == "Superadmin" {
		c.JSON(http.StatusForbidden, gin.H{"error": "Peran Superadmin absolut tidak boleh dijadikan default registrasi umum"})
		return
	}

	// Reset seluruh peran menjadi is_default = false
	config.DB.Model(&models.Role{}).Where("1 = 1").Update("is_default", false)

	// Set peran terpilih menjadi is_default = true
	if err := config.DB.Model(&role).Update("is_default", true).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal menetapkan status default pada peran"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Peran berhasil diaktifkan sebagai default bagi pendaftar baru",
		"role":    role,
	})
}

// UpdateRoleInput mendefinisikan skema pembaruan entitas peran
type UpdateRoleInput struct {
	Name          string   `json:"name" binding:"required"`
	Description   string   `json:"description"`
	PermissionIDs []string `json:"permission_ids"` // Array string UUID hak akses yang baru di-centang
}

// UpdateRole memutakhirkan data peran dan menyegarkan asosiasi hak aksesnya (Eksklusif Superadmin)
func UpdateRole(c *gin.Context) {
	roleID := c.Param("id")

	var role models.Role
	if err := config.DB.Preload("Permissions").First(&role, "id = ?", roleID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Peran tidak ditemukan"})
		return
	}

	// VALIDASI KEAMANAN MUTLAK: Peran asli bernama "Superadmin" tidak boleh dimodifikasi
	if role.Name == "Superadmin" {
		c.JSON(http.StatusForbidden, gin.H{"error": "Akses Ditolak: Entitas peran Superadmin dilindungi secara sistem dan tidak dapat dimodifikasi"})
		return
	}

	var input UpdateRoleInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Jika nama peran diubah, pastikan tidak bentrok dengan peran lain
	if input.Name != role.Name {
		var duplicate models.Role
		if err := config.DB.Where("name = ? AND id != ?", input.Name, role.ID).First(&duplicate).Error; err == nil {
			c.JSON(http.StatusConflict, gin.H{"error": "Nama peran tersebut telah digunakan oleh entitas lain"})
			return
		}
	}

	// Update data dasar
	role.Name = input.Name
	role.Description = input.Description
	if err := config.DB.Save(&role).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal menyimpan pembaruan peran"})
		return
	}

	// Perbarui asosiasi hak akses secara atomik menggunakan Replace
	var permissions []models.Permission
	if len(input.PermissionIDs) > 0 {
		config.DB.Where("id IN ?", input.PermissionIDs).Find(&permissions)
	}
	_ = config.DB.Model(&role).Association("Permissions").Replace(permissions)

	// Muat ulang untuk respons akhir
	config.DB.Preload("Permissions").First(&role, "id = ?", role.ID)

	c.JSON(http.StatusOK, gin.H{
		"message": "Entitas peran berhasil dimutakhirkan",
		"role":    role,
	})
}

// GetRoleById mengambil rincian satu peran spesifik beserta hak aksesnya berdasarkan UUID
func GetRoleById(c *gin.Context) {
	roleID := c.Param("id")

	var role models.Role
	if err := config.DB.Preload("Permissions").First(&role, "id = ?", roleID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Peran tidak ditemukan"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"role": role,
	})
}

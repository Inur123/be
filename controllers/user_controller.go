package controllers

import (
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"laci-cabang-be/config"
	"laci-cabang-be/models"

	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
)

// GetUsers mengambil daftar seluruh pengguna terdaftar beserta relasi perannya
func GetUsers(c *gin.Context) {
	// Identifikasi siapa yang sedang melakukan permintaan (requester)
	requesterID, _ := c.Get("user_id")

	var requester models.User
	config.DB.Preload("Role").First(&requester, "id = ?", requesterID)
	isSuperadmin := requester.RoleID != nil && requester.Role.Name == "Superadmin"

	var users []models.User
	query := config.DB.Preload("Role").Preload("Role.Permissions")

	// PROTEKSI DATA: Jika penanya bukan Superadmin, jangan kirim data user berkategori Superadmin
	if !isSuperadmin {
		var superRole models.Role
		config.DB.Where("name = ?", "Superadmin").First(&superRole)
		// Filter agar role_id tidak sama dengan Superadmin ID
		query = query.Where("role_id != ? OR role_id IS NULL", superRole.ID)
	}

	if err := query.Find(&users).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal mengambil daftar pengguna"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"users": users,
	})
}

// UpdateUserRole memungkinkan Superadmin mengubah wewenang/Role dari seorang user
func UpdateUserRole(c *gin.Context) {
	userID := c.Param("id")

	var user models.User
	if err := config.DB.Where("id = ?", userID).First(&user).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Pengguna tidak ditemukan"})
		return
	}

	var input models.UpdateUserRoleInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Validasi jika RoleID diberikan, pastikan Role tersebut ada
	if input.RoleID != nil {
		var role models.Role
		if err := config.DB.Where("id = ?", *input.RoleID).First(&role).Error; err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Peran (Role) yang dituju tidak valid atau tidak ditemukan"})
			return
		}
	}

	user.RoleID = input.RoleID
	if err := config.DB.Save(&user).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal memperbarui hak akses peran pengguna"})
		return
	}

	// Muat ulang relasi peran terbaru untuk dikembalikan
	config.DB.Preload("Role").Preload("Role.Permissions").First(&user, "id = ?", user.ID)

	c.JSON(http.StatusOK, gin.H{
		"message": "Peran pengguna berhasil diperbarui",
		"user":    user,
	})
}

// UpdateProfile memfasilitasi pengguna untuk memutakhirkan identitas/data dirinya sendiri
func UpdateProfile(c *gin.Context) {
	// Ambil klaim identitas dari middleware
	userID, ok := getUserIDFromContext(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Tidak terautentikasi"})
		return
	}

	var user models.User
	if err := config.DB.Where("id = ?", userID).First(&user).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Akun pengguna tidak ditemukan"})
		return
	}

	var input models.UpdateProfileInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Jika email diubah, pastikan tidak bertabrakan dengan milik orang lain
	if input.Email != "" && input.Email != user.Email {
		var duplicate models.User
		if err := config.DB.Where("email = ? AND id != ?", input.Email, user.ID).First(&duplicate).Error; err == nil {
			c.JSON(http.StatusConflict, gin.H{"error": "Alamat email tersebut telah terpakai oleh akun lain"})
			return
		}
		user.Email = input.Email
	}

	if input.Name != "" {
		user.Name = input.Name
	}

	if input.Avatar != "" && input.Avatar != user.Avatar {
		// Jika ada avatar lama, coba hapus file fisiknya dari storage
		if user.Avatar != "" {
			// Asumsi URL format: http://localhost:8080/uploads/namafile.jpg
			// Kita ambil bagian 'namafile.jpg' saja
			parts := strings.Split(user.Avatar, "/uploads/")
			if len(parts) > 1 {
				oldFileName := parts[1]
				oldFilePath := filepath.Join("./uploads", oldFileName)
				
				// Hapus file jika ada
				if _, err := os.Stat(oldFilePath); err == nil {
					_ = os.Remove(oldFilePath)
				}
			}
		}
		user.Avatar = input.Avatar
	}

	// Jika pengguna menyertakan password baru, lakukan hashing
	if input.Password != "" {
		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(input.Password), bcrypt.DefaultCost)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal memproses pengamanan kata sandi baru"})
			return
		}
		user.Password = string(hashedPassword)
	}

	if err := config.DB.Save(&user).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal menyimpan pembaruan profil"})
		return
	}

	// Muat ulang data lengkap beserta relasinya
	config.DB.Preload("Role").Preload("Role.Permissions").First(&user, "id = ?", user.ID)

	c.JSON(http.StatusOK, gin.H{
		"message": "Data profil berhasil dimutakhirkan",
		"user":    user,
	})
}

// GetUserById mengambil detail satu pengguna berdasarkan ID
func GetUserById(c *gin.Context) {
	id := c.Param("id")
	var user models.User
	if err := config.DB.Preload("Role").Preload("Role.Permissions").Where("id = ?", id).First(&user).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Pengguna tidak ditemukan"})
		return
	}

	// Cari periode aktif untuk pengguna ini
	var activePeriod models.Period
	config.DB.Where("user_id = ? AND is_active = ?", user.ID, true).First(&activePeriod)

	c.JSON(http.StatusOK, gin.H{
		"user":           user,
		"active_period": activePeriod,
	})
}

// AdminResetPassword memungkinkan pengelola sistem untuk mereset sandi pengguna lain
func AdminResetPassword(c *gin.Context) {
	id := c.Param("id")
	var input struct {
		Password string `json:"password" binding:"required,min=6"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Kata sandi baru minimal 6 karakter"})
		return
	}

	var user models.User
	if err := config.DB.Where("id = ?", id).First(&user).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Pengguna tidak ditemukan"})
		return
	}

	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte(input.Password), bcrypt.DefaultCost)
	user.Password = string(hashedPassword)

	if err := config.DB.Save(&user).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal mereset kata sandi"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Kata sandi pengguna berhasil diperbarui"})
}

// DeleteUser menghapus akun pengguna secara permanen dari sistem
func DeleteUser(c *gin.Context) {
	id := c.Param("id")
	// Pastikan tidak menghapus diri sendiri (opsional, tergantung kebijakan)
	// adminID, _ := getUserIDFromContext(c)
	// if id == adminID { ... }

	if err := config.DB.Delete(&models.User{}, "id = ?", id).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal menghapus akun pengguna"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Akun pengguna telah dihapus secara permanen"})
}

package controllers

import (
	"net/http"
	"os"
	"time"

	"laci-cabang-be/config"
	"laci-cabang-be/models"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

// Register menangani pendaftaran pengguna baru
func Register(c *gin.Context) {
	var input models.RegisterInput

	// Validasi input JSON
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Periksa apakah email sudah terdaftar
	var existingUser models.User
	if err := config.DB.Where("email = ?", input.Email).First(&existingUser).Error; err == nil {
		c.JSON(http.StatusConflict, gin.H{"error": "Email sudah terdaftar"})
		return
	}

	// Hash password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(input.Password), bcrypt.DefaultCost)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal mengenkripsi password"})
		return
	}

	// Simpan user baru ke database
	newUser := models.User{
		Name:     input.Name,
		Email:    input.Email,
		Password: string(hashedPassword),
	}

	// Cari peran default dinamis yang telah diaktifkan Superadmin
	var defaultRole models.Role
	if err := config.DB.Where("is_default = ?", true).First(&defaultRole).Error; err == nil {
		newUser.RoleID = &defaultRole.ID
	} else {
		// Fallback pintar: Gunakan peran terdaftar pertama selain Superadmin jika ada
		var fallbackRole models.Role
		if err := config.DB.Where("name != ?", "Superadmin").First(&fallbackRole).Error; err == nil {
			newUser.RoleID = &fallbackRole.ID
		}
	}

	if err := config.DB.Create(&newUser).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal membuat akun pengguna"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "Registrasi berhasil",
		"user":    newUser,
	})
}

// Login menangani autentikasi pengguna dan menghasilkan token JWT
func Login(c *gin.Context) {
	var input models.LoginInput

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Cari user berdasarkan email
	var user models.User
	if err := config.DB.Preload("Role.Permissions").Where("email = ?", input.Email).First(&user).Error; err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Email atau password salah"})
		return
	}

	// Verifikasi password
	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(input.Password)); err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Email atau password salah"})
		return
	}

	// PROTEKSI LOGIN RBAC ABSOLUT: Periksa izin otentikasi untuk non-Superadmin
	if user.RoleID == nil || user.Role.Name != "Superadmin" {
		hasLoginPermission := false
		for _, perm := range user.Role.Permissions {
			if perm.Name == "post_auth_login" {
				hasLoginPermission = true
				break
			}
		}
		if !hasLoginPermission {
			c.JSON(http.StatusForbidden, gin.H{"error": "Akses ditolak: Peran Anda tidak diizinkan untuk melakukan otentikasi masuk"})
			return
		}
	}

	// Buat token JWT
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user_id": user.ID,
		"exp":     time.Now().Add(time.Hour * 24).Unix(), // Kedaluwarsa dalam 24 jam
	})

	jwtSecret := os.Getenv("JWT_SECRET")
	tokenString, err := token.SignedString([]byte(jwtSecret))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal membuat token autentikasi"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Login berhasil",
		"token":   tokenString,
		"user":    user,
	})
}

// Logout menangani proses keluar
func Logout(c *gin.Context) {
	// Karena JWT bersifat stateless di sisi server, proses logout utamanya dilakukan di frontend
	// dengan cara menghapus token dari localStorage / cookie.
	// Di sini kita mengirimkan respons sukses kepada frontend untuk menyelesaikan proses.
	c.JSON(http.StatusOK, gin.H{
		"message": "Logout berhasil, silakan hapus token di sisi klien",
	})
}

// GetProfile mengambil data pengguna yang sedang login (membutuhkan middleware autentikasi)
func GetProfile(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Tidak terautentikasi"})
		return
	}

	var user models.User
	if err := config.DB.Preload("Role.Permissions").Where("id = ?", userID).First(&user).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Pengguna tidak ditemukan"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"user": user,
	})
}

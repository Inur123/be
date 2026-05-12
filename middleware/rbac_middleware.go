package middleware

import (
	"net/http"

	"laci-cabang-be/config"
	"laci-cabang-be/models"

	"github.com/gin-gonic/gin"
)

// RequireSuperadmin mengunci rute agar eksklusif hanya dapat diakses oleh peran Superadmin
func RequireSuperadmin() gin.HandlerFunc {
	return func(c *gin.Context) {
		userID, exists := c.Get("user_id")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Akses ditolak: Tidak terautentikasi"})
			c.Abort()
			return
		}

		var user models.User
		// Ambil pengguna beserta relasi Role-nya
		if err := config.DB.Preload("Role").First(&user, "id = ?", userID).Error; err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Akses ditolak: Sesi tidak valid"})
			c.Abort()
			return
		}

		// Periksa apakah peran pengguna adalah "Superadmin"
		if user.RoleID == nil || user.Role.Name != "Superadmin" {
			c.JSON(http.StatusForbidden, gin.H{"error": "Akses ditolak: Rute ini membutuhkan hak Superadmin"})
			c.Abort()
			return
		}

		c.Next()
	}
}

// RequirePermission memvalidasi apakah pengguna memiliki hak akses spesifik atau merupakan seorang Superadmin
func RequirePermission(requiredPermission string) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID, exists := c.Get("user_id")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Akses ditolak: Tidak terautentikasi"})
			c.Abort()
			return
		}

		var user models.User
		// Muat pengguna beserta struktur Role dan Permissions-nya
		if err := config.DB.Preload("Role.Permissions").First(&user, "id = ?", userID).Error; err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Akses ditolak: Sesi tidak valid"})
			c.Abort()
			return
		}

		// Jika pengguna adalah Superadmin, loloskan secara mutlak
		if user.RoleID != nil && user.Role.Name == "Superadmin" {
			c.Next()
			return
		}

		// Jika bukan Superadmin, periksa butiran hak aksesnya
		hasAccess := false
		if user.RoleID != nil {
			for _, perm := range user.Role.Permissions {
				if perm.Name == requiredPermission {
					hasAccess = true
					break
				}
			}
		}

		if !hasAccess {
			c.JSON(http.StatusForbidden, gin.H{"error": "Akses ditolak: Anda tidak memiliki izin untuk tindakan ini"})
			c.Abort()
			return
		}

		c.Next()
	}
}

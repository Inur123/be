package middleware

import (
	"fmt"
	"net/http"
	"os"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

func AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Header Authorization tidak ditemukan"})
			c.Abort()
			return
		}

		// Token biasanya memiliki format "Bearer <token>"
		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Format token tidak valid, gunakan 'Bearer <token>'"})
			c.Abort()
			return
		}

		tokenString := parts[1]
		jwtSecret := os.Getenv("JWT_SECRET")

		// Parse dan validasi token
		token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
			// Pastikan algoritma penandatanganan sesuai
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("algoritma penandatanganan tidak terduga: %v", token.Header["alg"])
			}
			return []byte(jwtSecret), nil
		})

		if err != nil || !token.Valid {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Token tidak valid atau sudah kedaluwarsa"})
			c.Abort()
			return
		}

		// Ekstrak klaim dari token
		if claims, ok := token.Claims.(jwt.MapClaims); ok {
			// Konversi userID ke int/uint secara aman
			if floatID, ok := claims["user_id"].(float64); ok {
				c.Set("user_id", uint(floatID))
			} else {
				c.JSON(http.StatusUnauthorized, gin.H{"error": "Klaim pengguna tidak valid di dalam token"})
				c.Abort()
				return
			}
		} else {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Gagal memproses klaim token"})
			c.Abort()
			return
		}

		// Lanjutkan ke handler berikutnya
		c.Next()
	}
}

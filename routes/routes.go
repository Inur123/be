package routes

import (
	"net/http"
	"time"

	"laci-cabang-be/controllers"
	_ "laci-cabang-be/docs" // Impor spesifikasi dokumentasi Swagger
	"laci-cabang-be/middleware"

	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

// corsMiddleware menangani izin akses lintas asal (CORS) dari frontend
func corsMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Mengizinkan semua asal selama pengembangan
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS, GET, PUT, DELETE")

		// Jika metode adalah OPTIONS (preflight request), langsung kembalikan status sukses
		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}

		c.Next()
	}
}

func SetupRoutes(r *gin.Engine) {
	// Terapkan middleware CORS secara global untuk semua rute
	r.Use(corsMiddleware())

	// Endpoint Root (Health Check / Informasi Sistem JSON)
	r.GET("/", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status":            "online",
			"service":           "Laci Cabang v3 Backend API",
			"version":           "3.0.0",
			"environment":       "development",
			"documentation_url": "http://localhost:8080/swagger/index.html",
			"timestamp":         time.Now().Format(time.RFC3339),
		})
	})

	// Rute Antarmuka Dokumentasi Interaktif Swagger UI
	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	api := r.Group("/api")
	{
		auth := api.Group("/auth")
		{
			auth.POST("/register", controllers.Register)
			auth.POST("/login", controllers.Login)
			auth.POST("/logout", controllers.Logout)

			// Endpoint terproteksi (membutuhkan token JWT)
			auth.GET("/me", middleware.AuthMiddleware(), controllers.GetProfile)
		}
	}
}

package routes

import (
	"fmt"
	"net/http"
	"strings"
	"time"

	"laci-cabang-be/config"
	"laci-cabang-be/controllers"
	_ "laci-cabang-be/docs" // Impor spesifikasi dokumentasi Swagger
	"laci-cabang-be/middleware"
	"laci-cabang-be/models"

	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

// corsMiddleware menangani izin akses lintas asal (CORS) dari frontend
func corsMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Mengambil asal pengirim secara dinamis untuk mematuhi spesifikasi ketat CORS saat kredensial aktif
		origin := c.Request.Header.Get("Origin")
		if origin != "" {
			c.Writer.Header().Set("Access-Control-Allow-Origin", origin)
		} else {
			c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		}
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

		// Rute Manajemen Peran dan Hak Akses (RBAC) secara Flat dan Lugas
		api.GET("/roles", middleware.AuthMiddleware(), controllers.GetRoles)
		api.GET("/roles/list", middleware.AuthMiddleware(), controllers.GetRoles)
		api.POST("/roles", middleware.AuthMiddleware(), middleware.RequirePermission("post_roles"), controllers.CreateRole)
		api.GET("/roles/permissions", middleware.AuthMiddleware(), controllers.GetPermissions)
		api.POST("/roles/permissions", middleware.AuthMiddleware(), middleware.RequirePermission("post_roles_permissions"), controllers.CreatePermission)
		api.PUT("/roles/:id/default", middleware.AuthMiddleware(), middleware.RequirePermission("put_roles_:id_default"), controllers.SetDefaultRole)
		api.GET("/roles/:id", middleware.AuthMiddleware(), controllers.GetRoleById)
		api.PUT("/roles/:id", middleware.AuthMiddleware(), middleware.RequirePermission("put_roles_:id"), controllers.UpdateRole)

		// Rute Manajemen Masa Kepengurusan (Periodisasi)
		api.GET("/periods", middleware.AuthMiddleware(), controllers.GetPeriods)
		api.GET("/periods/active", middleware.AuthMiddleware(), controllers.GetActivePeriod)
		api.POST("/periods", middleware.AuthMiddleware(), middleware.RequirePermission("post_periods"), controllers.CreatePeriod)
		api.PUT("/periods/:id/active", middleware.AuthMiddleware(), middleware.RequirePermission("put_periods_:id_active"), controllers.SetActivePeriod)
		api.PUT("/periods/:id", middleware.AuthMiddleware(), middleware.RequirePermission("put_periods_:id"), controllers.UpdatePeriod)
		api.DELETE("/periods/:id", middleware.AuthMiddleware(), middleware.RequirePermission("delete_periods_:id"), controllers.DeletePeriod)

		// Rute Manajemen Akun dan Identitas Pengguna (Users)
		api.GET("/users", middleware.AuthMiddleware(), controllers.GetUsers)
		api.GET("/users/:id", middleware.AuthMiddleware(), controllers.GetUserById)
		api.PUT("/users/profile", middleware.AuthMiddleware(), controllers.UpdateProfile)
		api.PUT("/users/:id/role", middleware.AuthMiddleware(), middleware.RequirePermission("put_users_:id_role"), controllers.UpdateUserRole)
		api.PUT("/users/:id/reset-password", middleware.AuthMiddleware(), middleware.RequirePermission("put_users_:id_reset-password"), controllers.AdminResetPassword)
		api.DELETE("/users/:id", middleware.AuthMiddleware(), middleware.RequirePermission("delete_users_:id"), controllers.DeleteUser)

		// Rute Upload File
		api.POST("/upload", middleware.AuthMiddleware(), controllers.UploadFile)
	}

	// Sajikan file statis dari direktori uploads agar bisa diakses browser
	r.Static("/uploads", "./uploads")

	// SINKRONISASI OTOMATIS: Memindai rute yang didaftarkan untuk dijadikan butiran Permission
	syncPermissionsFromRoutes(r)
}

// syncPermissionsFromRoutes memindai seluruh rute /api terdaftar dan otomatis menyuntikkannya ke tabel permissions
func syncPermissionsFromRoutes(r *gin.Engine) {
	fmt.Println("Seeding: Memindai rute API terdaftar untuk di-generate menjadi butiran hak akses otomatis...")
	for _, route := range r.Routes() {
		// Kita hanya mengambil rute di bawah /api
		if !strings.HasPrefix(route.Path, "/api") {
			continue
		}

		// Susun nama izin yang deskriptif dan manusiawi
		cleanPath := strings.TrimPrefix(route.Path, "/api/")
		cleanPath = strings.ReplaceAll(cleanPath, "/", "_")
		if cleanPath == "" || cleanPath == "_" {
			cleanPath = "root"
		}

		permName := strings.ToLower(route.Method) + "_" + cleanPath
		permName = strings.ReplaceAll(permName, "__", "_")
		permName = strings.TrimSuffix(permName, "_")

		var existing models.Permission
		if err := config.DB.Where("name = ?", permName).First(&existing).Error; err != nil {
			desc := fmt.Sprintf("Mengizinkan akses eksekusi ke endpoint %s %s", route.Method, route.Path)
			config.DB.Create(&models.Permission{
				Name:        permName,
				Description: desc,
			})
		}
	}
	fmt.Println("Sinkronisasi butir hak akses rute selesai!")
}

package main

import (
	"fmt"
	"log"
	"os"

	"laci-cabang-be/config"
	"laci-cabang-be/models"
	"laci-cabang-be/routes"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

func main() {
	// Muat konfigurasi dari file .env
	if err := godotenv.Load(); err != nil {
		log.Println("Peringatan: File .env tidak ditemukan, menggunakan variabel lingkungan sistem.")
	}

	// Hubungkan ke database
	config.ConnectDB()

	// Migrasi otomatis skema database
	fmt.Println("Menjalankan migrasi database...")
	if err := config.DB.AutoMigrate(&models.User{}); err != nil {
		log.Fatalf("Gagal melakukan migrasi database: %v", err)
	}
	fmt.Println("Migrasi database berhasil!")

	// Inisialisasi router Gin
	r := gin.Default()

	// Siapkan rute API
	routes.SetupRoutes(r)

	// Dapatkan port dari .env atau gunakan port default 8080
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	// Jalankan server
	addr := fmt.Sprintf(":%s", port)
	fmt.Printf("Server berjalan di http://localhost%s\n", addr)
	if err := r.Run(addr); err != nil {
		log.Fatalf("Gagal menjalankan server: %v", err)
	}
}

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
	"golang.org/x/crypto/bcrypt"
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
	if err := config.DB.AutoMigrate(&models.Permission{}, &models.Role{}, &models.User{}, &models.Period{}); err != nil {
		log.Fatalf("Gagal melakukan migrasi database: %v", err)
	}
	fmt.Println("Migrasi database berhasil!")

	// SEEDING OTOMATIS: Buat Peran Superadmin & Akun Pertama jika belum ada
	var superadminRole models.Role
	if err := config.DB.Where("name = ?", "Superadmin").First(&superadminRole).Error; err != nil {
		fmt.Println("Seeding: Menciptakan entitas peran 'Superadmin'...")
		superadminRole = models.Role{
			Name:        "Superadmin",
			Description: "Administrator Tertinggi dengan Hak Akses Tanpa Batas",
		}
		if errCreate := config.DB.Create(&superadminRole).Error; errCreate != nil {
			fmt.Printf("❌ ERROR SEEDING ROLE SUPERADMIN: %v\n", errCreate)
		} else {
			fmt.Printf("✅ Sukses seeding Role Superadmin dengan ID: %s\n", superadminRole.ID)
		}
	}

	var adminUser models.User
	if err := config.DB.Where("email = ?", "admin@gmail.com").First(&adminUser).Error; err != nil {
		fmt.Println("Seeding: Menciptakan Akun Default 'admin@gmail.com' sebagai Superadmin...")
		hashedPassword, _ := bcrypt.GenerateFromPassword([]byte("password"), bcrypt.DefaultCost)
		adminUser = models.User{
			Name:     "Super Admin",
			Email:    "admin@gmail.com",
			Password: string(hashedPassword),
			RoleID:   &superadminRole.ID,
		}
		if errCreate := config.DB.Create(&adminUser).Error; errCreate != nil {
			fmt.Printf("❌ ERROR SEEDING USER ADMIN: %v\n", errCreate)
		} else {
			fmt.Printf("✅ Sukses seeding Akun Superadmin!\n")
		}
	}

	// SEEDING OTOMATIS: Buat Peran Default 'Anggota' untuk pendaftar baru jika belum ada
	var anggotaRole models.Role
	if err := config.DB.Where("name = ?", "Anggota").First(&anggotaRole).Error; err != nil {
		fmt.Println("Seeding: Menciptakan entitas peran default 'Anggota' untuk pendaftar baru...")
		anggotaRole = models.Role{
			Name:        "Anggota",
			Description: "Peran keanggotaan standar bagi setiap pengguna yang baru melakukan registrasi",
			IsDefault:   true,
		}
		if errCreate := config.DB.Create(&anggotaRole).Error; errCreate != nil {
			fmt.Printf("❌ ERROR SEEDING ROLE ANGGOTA: %v\n", errCreate)
		} else {
			fmt.Printf("✅ Sukses seeding Role Anggota dengan ID: %s\n", anggotaRole.ID)
		}
	}

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

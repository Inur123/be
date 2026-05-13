package controllers

import (
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/gin-gonic/gin"
)

// UploadFile menangani pengunggahan file gambar (seperti avatar) ke server
func UploadFile(c *gin.Context) {
	// Ambil file dari request form-data
	file, err := c.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Tidak ada file yang diunggah"})
		return
	}

	// Pastikan direktori uploads tersedia
	uploadDir := "./uploads"
	if _, err := os.Stat(uploadDir); os.IsNotExist(err) {
		os.Mkdir(uploadDir, 0755)
	}

	// Buat nama file unik untuk menghindari tabrakan
	extension := filepath.Ext(file.Filename)
	newFileName := fmt.Sprintf("%d%s", time.Now().UnixNano(), extension)
	dst := filepath.Join(uploadDir, newFileName)

	// Simpan file ke direktori tujuan
	if err := c.SaveUploadedFile(file, dst); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal menyimpan file di server"})
		return
	}

	// Kembalikan URL lengkap file yang bisa diakses
	// Catatan: Di produksi, gunakan domain asli. Di dev, gunakan localhost:8080
	fileURL := fmt.Sprintf("http://localhost:8080/uploads/%s", newFileName)

	c.JSON(http.StatusOK, gin.H{
		"message": "File berhasil diunggah",
		"url":     fileURL,
	})
}

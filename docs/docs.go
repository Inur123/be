package docs

import "github.com/swaggo/swag"

const docTemplate = `{
  "swagger": "2.0",
  "info": {
    "description": "Dokumentasi resmi antarmuka REST API untuk sistem Laci Cabang v3.",
    "title": "Laci Cabang v3 API",
    "contact": {},
    "version": "3.0.0"
  },
  "host": "localhost:8080",
  "basePath": "/",
  "paths": {
    "/api/auth/register": {
      "post": {
        "description": "Mendaftarkan entitas pengguna baru ke dalam sistem basis data.",
        "consumes": ["application/json"],
        "produces": ["application/json"],
        "tags": ["Authentication"],
        "summary": "Daftar Akun Baru",
        "parameters": [
          {
            "description": "Struktur Data Pendaftaran",
            "name": "body",
            "in": "body",
            "required": true,
            "schema": {
              "$ref": "#/definitions/models.RegisterInput"
            }
          }
        ],
        "responses": {
          "201": { "description": "Akun berhasil dibuat" },
          "400": { "description": "Format permintaan tidak valid" },
          "409": { "description": "Email sudah terdaftar" }
        }
      }
    },
    "/api/auth/login": {
      "post": {
        "description": "Memverifikasi kredensial pengguna dan menerbitkan token otentikasi JWT yang aman.",
        "consumes": ["application/json"],
        "produces": ["application/json"],
        "tags": ["Authentication"],
        "summary": "Masuk Sesi (Login)",
        "parameters": [
          {
            "description": "Kredensial Akses",
            "name": "body",
            "in": "body",
            "required": true,
            "schema": {
              "$ref": "#/definitions/models.LoginInput"
            }
          }
        ],
        "responses": {
          "200": { "description": "Login berhasil dan token diterbitkan" },
          "400": { "description": "Format permintaan tidak valid" },
          "401": { "description": "Kredensial tidak sah" }
        }
      }
    },
    "/api/auth/logout": {
      "post": {
        "description": "Mengakhiri sesi otentikasi pengguna secara aman.",
        "produces": ["application/json"],
        "tags": ["Authentication"],
        "summary": "Keluar Sesi (Logout)",
        "responses": {
          "200": { "description": "Sesi diakhiri" }
        }
      }
    },
    "/api/auth/me": {
      "get": {
        "security": [
          {
            "BearerAuth": []
          }
        ],
        "description": "Mengambil informasi profil dan meta-data pengguna yang sedang terautentikasi.",
        "produces": ["application/json"],
        "tags": ["Authentication"],
        "summary": "Dapatkan Profil Sesi Aktif",
        "responses": {
          "200": { "description": "Profil berhasil dimuat" },
          "401": { "description": "Token JWT tidak valid atau tidak ditemukan" }
        }
      }
    }
  },
  "definitions": {
    "models.RegisterInput": {
      "type": "object",
      "required": ["name", "email", "password"],
      "properties": {
        "name": { "type": "string", "example": "Pengguna Teladan" },
        "email": { "type": "string", "example": "teladan@cabang.com" },
        "password": { "type": "string", "example": "sandi-rahasia" }
      }
    },
    "models.LoginInput": {
      "type": "object",
      "required": ["email", "password"],
      "properties": {
        "email": { "type": "string", "example": "teladan@cabang.com" },
        "password": { "type": "string", "example": "sandi-rahasia" }
      }
    }
  },
  "securityDefinitions": {
    "BearerAuth": {
      "type": "apiKey",
      "name": "Authorization",
      "in": "header",
      "description": "Masukkan token JWT dengan format: Bearer {token}"
    }
  }
}`

// SwaggerInfo menampung struktur meta-data spesifikasi Swagger
var SwaggerInfo = &swag.Spec{
	Version:          "3.0.0",
	Host:             "localhost:8080",
	BasePath:         "/",
	Schemes:          []string{},
	Title:            "Laci Cabang v3 API",
	Description:      "Dokumentasi resmi antarmuka REST API untuk sistem Laci Cabang v3.",
	InfoInstanceName: "swagger",
	SwaggerTemplate:  docTemplate,
}

func init() {
	swag.Register(SwaggerInfo.InstanceName(), SwaggerInfo)
}

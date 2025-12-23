package main

import (
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	_ "backend-penjualan/docs" // Untuk Swagger

	"github.com/joho/godotenv"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	"backend-penjualan/models"
	"backend-penjualan/routes"
)

func main() {
	// Load .env
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, using system env vars")
	}

	// DSN
	dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s sslmode=disable TimeZone=Asia/Jakarta",
		os.Getenv("DB_HOST"), os.Getenv("DB_USER"), os.Getenv("DB_PASSWORD"),
		os.Getenv("DB_NAME"), os.Getenv("DB_PORT"))
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}

	// Setup DB connection pool
	sqlDB, err := db.DB()
	if err != nil {
		log.Fatal("Failed to get sql.DB:", err)
	}
	sqlDB.SetMaxIdleConns(10)
	sqlDB.SetMaxOpenConns(100)
	sqlDB.SetConnMaxLifetime(time.Hour)

	// Drop existing tables untuk fresh start (hilangin data lama)
	log.Println("Dropping existing tables for fresh migration...")
	if err := db.Migrator().DropTable(&models.Product{}); err != nil && !strings.Contains(err.Error(), "does not exist") {
		log.Printf("Warning: Failed to drop products table: %v", err)
	}
	if err := db.Migrator().DropTable(&models.Transaction{}); err != nil && !strings.Contains(err.Error(), "does not exist") {
		log.Printf("Warning: Failed to drop transactions table: %v", err)
	}

	// Auto-migrate models (sekarang fresh, no conflict)
	if err := db.AutoMigrate(&models.Product{}, &models.Transaction{}); err != nil {
		log.Fatal("Failed to migrate database:", err)
	}
	log.Println("Database migrated successfully (fresh tables created)")

	// Setup router & run server
	router := routes.SetupRouter(db)
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	log.Printf("Server starting on port %s", port)
	if err := router.Run(":" + port); err != nil {
		log.Fatal("Failed to start server:", err)
	}
}
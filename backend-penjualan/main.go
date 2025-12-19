package main

import (
	"fmt"
	"log"
	"os"
	"time"

	_ "backend-penjualan/docs"

	"github.com/joho/godotenv"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	"backend-penjualan/controllers"
	"backend-penjualan/models"
	"backend-penjualan/routes"
)

func main() {
    // Load .env (di awal)
    if err := godotenv.Load(); err != nil {
        log.Println("No .env file found")  // Optional, kalau nggak ada .env, lanjut
    }

    // DSN dari env vars
    dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s sslmode=disable",
        os.Getenv("DB_HOST"), os.Getenv("DB_USER"), os.Getenv("DB_PASSWORD"),
        os.Getenv("DB_NAME"), os.Getenv("DB_PORT"))
    db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
    if err != nil {
        log.Fatal("Failed to connect to database:", err)
    }

    // Sisanya sama...
    sqlDB, err := db.DB()
    if err != nil {
        log.Fatal("Failed to get sql.DB:", err)
    }
    sqlDB.SetConnMaxLifetime(time.Hour)
    db = db.Session(&gorm.Session{NewDB: true})

    if err := db.AutoMigrate(&models.Transaction{}); err != nil {
        log.Fatal("Failed to migrate database:", err)
    }

    ctrl := controllers.NewTransactionController(db)
    router := routes.SetupRouter(ctrl)
    router.Run(":8080")
}
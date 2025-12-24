package routes

import (
	"net/http"
	"time"  // Buat timestamp

	"github.com/gin-gonic/gin"
	ginSwagger "github.com/swaggo/gin-swagger"
	swaggerFiles "github.com/swaggo/files"
	"github.com/gin-contrib/cors"

	"backend-penjualan/controllers"
	"gorm.io/gorm"
)

// SetupRouter inisialisasi router dengan semua routes (products, transactions, forecast)
func SetupRouter(db *gorm.DB) *gin.Engine {
	r := gin.Default()

	// CORS untuk frontend (localhost:3000)
	r.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"http://localhost:3000", "http://127.0.0.1:3000"},
		AllowMethods:     []string{"GET", "POST", "PATCH", "DELETE", "OPTIONS", "PUT"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Accept"},
		AllowCredentials: true,
		ExposeHeaders:    []string{"Content-Length"},
	}))

	// Health check sederhana (update timestamp ke current)
	r.GET("/ping", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "OK", "timestamp": time.Now().UTC().Format(time.RFC3339)})
	})

	// Swagger docs (jalankan di /swagger/index.html)
	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	// API v1 group
	v1 := r.Group("/api/v1")
	{
		// Inisialisasi controllers di sini (butuh db)
		productCtrl := controllers.NewProductController(db)
		transactionCtrl := controllers.NewTransactionController(db)
		forecastCtrl := controllers.NewForecastController()  // Tambah ini!

		// Products routes
		v1.GET("/products", productCtrl.GetAll)
		v1.POST("/products", productCtrl.Create)
		v1.GET("/products/:id", productCtrl.GetByID)
		v1.PUT("/products/:id", productCtrl.Update)
		v1.DELETE("/products/:id", productCtrl.Delete)

		// Transactions routes
		v1.GET("/transactions", transactionCtrl.GetAll)
		v1.POST("/transactions", transactionCtrl.Create)
		v1.GET("/transactions/:id", transactionCtrl.GetByID)
		v1.PATCH("/transactions/:id", transactionCtrl.Update)
		v1.DELETE("/transactions/:id", transactionCtrl.Delete)

		// Forecast routes (baru!)
		v1.POST("/forecast/upload", forecastCtrl.UploadHandler)
		// Tambahan: Health check buat ML service (test koneksi)
		v1.GET("/forecast/health", func(c *gin.Context) {
			// Simple ping ke ML URL (dari controller config)
			url := forecastCtrl.FastAPIURL
			resp, err := http.Get(url + "/health") // Asumsi ML punya /health, atau ganti ke root
			if err != nil || resp.StatusCode != http.StatusOK {
				c.JSON(http.StatusServiceUnavailable, gin.H{"status": "ML service down", "details": err.Error()})
				return
			}
			c.JSON(http.StatusOK, gin.H{"status": "ML service healthy", "url": url})
		})
	}

	return r
}
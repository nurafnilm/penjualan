package routes

import (
    "net/http"

    "github.com/gin-gonic/gin"
    ginSwagger "github.com/swaggo/gin-swagger"
    swaggerFiles "github.com/swaggo/files"
    "github.com/gin-contrib/cors"

    "backend-penjualan/controllers" // Adjust kalau path beda
    "gorm.io/gorm"
)

// SetupRouter inisialisasi router dengan semua routes (products & transactions)
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

    // Health check sederhana
    r.GET("/ping", func(c *gin.Context) {
        c.JSON(http.StatusOK, gin.H{"message": "OK", "timestamp": "2025-12-22T00:00:00Z"}) // Adjust date kalau perlu
    })

    // Swagger docs (jalankan di /swagger/index.html)
    r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

    // API v1 group
    v1 := r.Group("/api/v1")
    {
        // Inisialisasi controllers di sini (butuh db)
        productCtrl := controllers.NewProductController(db)
        transactionCtrl := controllers.NewTransactionController(db)

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
    }

    return r
}
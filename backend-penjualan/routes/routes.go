package routes

import (
    "github.com/gin-gonic/gin"
    ginSwagger "github.com/swaggo/gin-swagger"
    swaggerFiles "github.com/swaggo/files"

    "backend-penjualan/controllers"
)

// @title Sales API
// @version 1.0
// @description API for electronic sales transactions (Erajaya-like)
// @host localhost:8080
// @BasePath /api/v1
// @schemes http https

func SetupRouter(ctrl *controllers.TransactionController) *gin.Engine {
    r := gin.Default()

    // Swagger setup
    r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

    // API routes
    v1 := r.Group("/api/v1")
    {
        v1.GET("/transactions", ctrl.GetAll)
        v1.POST("/transactions", ctrl.Create)
        v1.PATCH("/transactions/:id", ctrl.Update)  // BARU: PATCH by ID
        v1.DELETE("/transactions/:id", ctrl.Delete) // BARU: DELETE by ID
    }

    return r
}
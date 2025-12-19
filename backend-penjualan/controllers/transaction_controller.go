package controllers

import (
	"net/http"
	"strconv"
	"time"

	"backend-penjualan/models"
	"errors"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type TransactionController struct {
    DB *gorm.DB
}

func NewTransactionController(db *gorm.DB) *TransactionController {
    return &TransactionController{DB: db}
}

// GET: Ambil semua transaksi (dengan optional filter)
// @Summary Get all sales transactions
// @Description Retrieve list of all sales transactions, optional filter by product_id or start_date
// @Tags transactions
// @Accept json
// @Produce json
// @Param product_id query string false "Filter by product ID (e.g., IPHONE15-001)"
// @Param start_date query string false "Filter by start date (YYYY-MM-DD, e.g., 2025-12-01)"
// @Success 200 {array} models.Transaction
// @Router /api/v1/transactions [get]
func (ctrl *TransactionController) GetAll(c *gin.Context) {
    var transactions []models.Transaction

    // Base query
    query := ctrl.DB

    // Optional filter: product_id
    if productID := c.Query("product_id"); productID != "" {
        query = query.Where("product_id = ?", productID)
    }

    // Optional filter: start_date (YYYY-MM-DD)
    if startDate := c.Query("start_date"); startDate != "" {
        parsedDate, err := time.Parse("2006-01-02", startDate)
        if err != nil {
            c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid start_date format. Use YYYY-MM-DD"})
            return
        }
        query = query.Where("created_at >= ?", parsedDate)
    }

    if err := query.Find(&transactions).Error; err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
        return
    }

    // Hitung total di response
    for i := range transactions {
        transactions[i].Total = float64(transactions[i].Quantity) * transactions[i].Price
    }
    c.JSON(http.StatusOK, transactions)
}

// POST: Buat transaksi baru
// @Summary Create a new sales transaction
// @Description Create a new transaction for electronic product sale
// @Tags transactions
// @Accept json
// @Produce json
// @Param transaction body models.Transaction true "Transaction data"
// @Success 201 {object} models.Transaction
// @Failure 400 {object} map[string]string "Validation error"
// @Router /api/v1/transactions [post]
func (ctrl *TransactionController) Create(c *gin.Context) {
    var input models.Transaction
    if err := c.ShouldBindJSON(&input); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
        return
    }

    if input.Quantity <= 0 || input.Price <= 0 {
        c.JSON(http.StatusBadRequest, gin.H{"error": "Quantity and price must be positive"})
        return
    }

    var existing models.Transaction
    if err := ctrl.DB.Where("product_id = ?", input.ProductID).First(&existing).Error; err == nil {
        c.JSON(http.StatusConflict, gin.H{"error": "Product ID already has a transaction"})
        return
    }

    if err := ctrl.DB.Create(&input).Error; err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
        return
    }

    input.Total = float64(input.Quantity) * input.Price
    c.JSON(http.StatusCreated, input)
}

// PATCH: Update partial transaksi by ID
// @Summary Update a sales transaction partially
// @Description Update specific fields of a transaction (e.g., quantity or price)
// @Tags transactions
// @Accept json
// @Produce json
// @Param id path int true "Transaction ID"
// @Param updates body map[string]interface{} true "Fields to update (e.g., quantity, price)"
// @Success 200 {object} models.Transaction
// @Failure 400 {object} map[string]string "Invalid input"
// @Failure 404 {object} map[string]string "Not found"
// @Router /api/v1/transactions/{id} [patch]
func (ctrl *TransactionController) Update(c *gin.Context) {
    id, err := strconv.Atoi(c.Param("id"))
    if err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID format"})
        return
    }

    // Ambil transaksi existing
    var transaction models.Transaction
    if err := ctrl.DB.First(&transaction, id).Error; err != nil {
        if errors.Is(err, gorm.ErrRecordNotFound) {
            c.JSON(http.StatusNotFound, gin.H{"error": "Transaction not found"})
        } else {
            c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
        }
        return
    }

    // Bind partial updates dari body (map[string]interface{} biar fleksibel)
    var updates map[string]interface{}
    if err := c.ShouldBindJSON(&updates); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
        return
    }

    // Validasi sederhana (kalau ada quantity/price di updates)
    if q, ok := updates["quantity"].(float64); ok && int(q) <= 0 {
        c.JSON(http.StatusBadRequest, gin.H{"error": "Quantity must be positive"})
        return
    }
    if p, ok := updates["price"].(float64); ok && p <= 0 {
        c.JSON(http.StatusBadRequest, gin.H{"error": "Price must be positive"})
        return
    }

    // Update partial
    if err := ctrl.DB.Model(&transaction).Updates(updates).Error; err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
        return
    }

    // Recompute total & response full data
    ctrl.DB.First(&transaction, id)
    transaction.Total = float64(transaction.Quantity) * transaction.Price
    c.JSON(http.StatusOK, transaction)
}

// DELETE: Hapus transaksi by ID (soft delete)
// @Summary Delete a sales transaction
// @Description Soft delete a transaction by ID
// @Tags transactions
// @Produce json
// @Param id path int true "Transaction ID"
// @Success 200 {object} map[string]string "Success message"
// @Failure 400 {object} map[string]string "Invalid input"
// @Failure 404 {object} map[string]string "Not found"
// @Router /api/v1/transactions/{id} [delete]
func (ctrl *TransactionController) Delete(c *gin.Context) {
    id, err := strconv.Atoi(c.Param("id"))
    if err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID format"})
        return
    }

    // Hapus soft (set deleted_at)
    if err := ctrl.DB.Delete(&models.Transaction{}, id).Error; err != nil {
        if errors.Is(err, gorm.ErrRecordNotFound) {
            c.JSON(http.StatusNotFound, gin.H{"error": "Transaction not found"})
        } else {
            c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
        }
        return
    }

    c.JSON(http.StatusOK, gin.H{"message": "Transaction deleted successfully"})
}
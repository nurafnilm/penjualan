package controllers

import (
	"fmt"
	"net/http"
	"strconv"
	"time"

	"backend-penjualan/models"
	"errors"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// TransactionController godoc
// @Description Transaction controller handles CRUD operations for transactions
type TransactionController struct {
	DB *gorm.DB
}

func NewTransactionController(db *gorm.DB) *TransactionController {
	return &TransactionController{DB: db}
}

// GetAll godoc
// @Summary Get all transactions
// @Description Retrieve list of all sales transactions with optional filters (product_id, start_date, search by buyer or product name)
// @Tags transactions
// @Accept json
// @Produce json
// @Param product_id query int false "Filter by product ID"
// @Param start_date query string false "Filter by start date (YYYY-MM-DD)"
// @Param search query string false "Search by buyer name or product name (partial)"
// @Success 200 {array} models.Transaction "List of transactions (with preloaded Product)"
// @Failure 400 {object} map[string]string "Invalid filter format"
// @Failure 500 {object} map[string]string "Query failed"
// @Router /transactions [get]
func (ctrl *TransactionController) GetAll(c *gin.Context) {
	var transactions []models.Transaction

	// Base query dengan preload Product
	query := ctrl.DB.Preload("Product")

	// Optional filter: product_id (exact)
	if productIDStr := c.Query("product_id"); productIDStr != "" {
		if productID, err := strconv.Atoi(productIDStr); err == nil {
			query = query.Where("product_id = ?", productID)
		} else {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid product_id format (must be number)"})
			return
		}
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

	// Optional filter: search by nama_pembeli or product.nama (partial, case-insensitive)
	if search := c.Query("search"); search != "" {
		// JOIN untuk filter on nama_pembeli or products.nama
		query = query.Joins("JOIN products ON products.id = transactions.product_id").
			Where("transactions.nama_pembeli ILIKE ? OR products.nama ILIKE ?", "%"+search+"%", "%"+search+"%")
	}

	if err := query.Find(&transactions).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Query failed: %v", err.Error())})
		return
	}

	c.JSON(http.StatusOK, transactions)
}

// GetByID godoc
// @Summary Get transaction by ID
// @Description Retrieve a specific transaction by ID with preloaded Product
// @Tags transactions
// @Accept json
// @Produce json
// @Param id path int true "Transaction ID"
// @Success 200 {object} models.Transaction "Transaction details (with Product)"
// @Failure 400 {object} map[string]string "Invalid ID"
// @Failure 404 {object} map[string]string "Transaction not found"
// @Failure 500 {object} map[string]string "Query failed"
// @Router /transactions/{id} [get]
func (ctrl *TransactionController) GetByID(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID format"})
		return
	}

	var transaction models.Transaction
	if err := ctrl.DB.Preload("Product").First(&transaction, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "Transaction not found"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Query failed: %v", err.Error())})
		}
		return
	}

	c.JSON(http.StatusOK, transaction)
}

// Create godoc
// @Summary Create a new transaction
// @Description Create a new sales transaction (fetch price from product, compute total)
// @Tags transactions
// @Accept json
// @Produce json
// @Param input body CreateTransactionInput true "Transaction input (nama_pembeli, product_id, quantity)"
// @Success 201 {object} models.Transaction "Created transaction (with Product)"
// @Failure 400 {object} map[string]string "Validation error"
// @Failure 404 {object} map[string]string "Product not found"
// @Failure 500 {object} map[string]string "Create failed"
// @Router /transactions [post]
type CreateTransactionInput struct {
	NamaPembeli string `json:"nama_pembeli"`
	ProductID   int64  `json:"product_id"`
	Quantity    int64  `json:"quantity"`
}

func (ctrl *TransactionController) Create(c *gin.Context) {
	var input CreateTransactionInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("Invalid input: %v", err)})
		return
	}

	// Validasi manual
	if input.NamaPembeli == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Nama pembeli required"})
		return
	}
	if input.ProductID <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Product ID required & positive"})
		return
	}
	if input.Quantity <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Quantity must be at least 1"})
		return
	}

	// Fetch product (cast int64 ke uint)
	var product models.Product
	if err := ctrl.DB.First(&product, uint(input.ProductID)).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "Product not found"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Product query failed: %v", err.Error())})
		}
		return
	}

	// Create transaction (cast ke uint)
	transaction := models.Transaction{
		NamaPembeli: input.NamaPembeli,
		ProductID:   uint(input.ProductID),
		Quantity:    uint(input.Quantity),
		Harga:       product.Harga,
		Total:       float64(input.Quantity) * product.Harga,
	}
	if err := ctrl.DB.Create(&transaction).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Create failed: %v", err.Error())})
		return
	}

	// Preload & response
	if err := ctrl.DB.Preload("Product").First(&transaction, transaction.ID).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Preload failed: %v", err.Error())})
		return
	}
	c.JSON(http.StatusCreated, transaction)
}

// Update godoc
// @Summary Update a transaction partially
// @Description Update partial fields (nama_pembeli or quantity), recompute total from product price
// @Tags transactions
// @Accept json
// @Produce json
// @Param id path int true "Transaction ID"
// @Param input body UpdateTransactionInput true "Fields to update (optional)"
// @Success 200 {object} models.Transaction "Updated transaction (with Product)"
// @Failure 400 {object} map[string]string "Validation error or no changes"
// @Failure 404 {object} map[string]string "Transaction not found"
// @Failure 500 {object} map[string]string "Update failed"
// @Router /transactions/{id} [patch]
type UpdateTransactionInput struct {
	NamaPembeli *string `json:"nama_pembeli"`
	Quantity    *uint   `json:"quantity"`
}

func (ctrl *TransactionController) Update(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID format"})
		return
	}

	var transaction models.Transaction
	if err := ctrl.DB.First(&transaction, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "Transaction not found"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Query failed: %v", err.Error())})
		}
		return
	}

	// Struct partial untuk bind: optional fields, type-safe
	var input UpdateTransactionInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("Invalid input: %v", err)})
		return
	}

	// Validasi manual (hanya kalo field diisi)
	updates := map[string]interface{}{}
	if input.NamaPembeli != nil {
		if *input.NamaPembeli == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Nama pembeli cannot be empty"})
			return
		}
		updates["nama_pembeli"] = *input.NamaPembeli
	}
	if input.Quantity != nil {
		if *input.Quantity == 0 { // uint, jadi ==0 invalid
			c.JSON(http.StatusBadRequest, gin.H{"error": "Quantity must be at least 1"})
			return
		}
		updates["quantity"] = *input.Quantity
	}

	// Kalo gak ada update apa-apa, return early
	if len(updates) == 0 {
		c.JSON(http.StatusOK, gin.H{"message": "No changes provided"})
		return
	}

	// Update fields partial
	if err := ctrl.DB.Model(&transaction).Updates(updates).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Update failed: %v", err.Error())})
		return
	}

	// Recompute harga/total: refresh transaction dulu biar Quantity up-to-date
	if err := ctrl.DB.Preload("Product").First(&transaction, id).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to refresh transaction"})
		return
	}

	// Update harga & total berdasarkan product terbaru
	newHarga := transaction.Product.Harga // Dari preload, lebih aman
	newTotal := float64(transaction.Quantity) * newHarga
	if err := ctrl.DB.Model(&transaction).Updates(map[string]interface{}{
		"harga": newHarga,
		"total": newTotal,
	}).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Recompute failed: %v", err.Error())})
		return
	}

	// Response full (udah preloaded)
	c.JSON(http.StatusOK, transaction)
}

// Delete godoc
// @Summary Delete a transaction
// @Description Soft delete a transaction by ID (or hard delete if Unscoped)
// @Tags transactions
// @Accept json
// @Produce json
// @Param id path int true "Transaction ID"
// @Success 200 {object} map[string]string "Success message"
// @Failure 400 {object} map[string]string "Invalid ID"
// @Failure 404 {object} map[string]string "Transaction not found"
// @Failure 500 {object} map[string]string "Delete failed"
// @Router /transactions/{id} [delete]
func (ctrl *TransactionController) Delete(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID format"})
		return
	}

	// Soft delete: set deleted_at; kalo mau hard, ganti ke ctrl.DB.Delete(&models.Transaction{}, id)
	if err := ctrl.DB.Unscoped().Delete(&models.Transaction{}, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "Transaction not found"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Delete failed: %v", err.Error())})
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Transaction deleted successfully"})
}
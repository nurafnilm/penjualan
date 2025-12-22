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

type TransactionController struct {
	DB *gorm.DB
}

func NewTransactionController(db *gorm.DB) *TransactionController {
	return &TransactionController{DB: db}
}

// GET: Ambil semua transaksi (dengan optional filter: product_id, start_date, search=nama_pembeli or product.nama)
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

// GET by ID untuk detail (preload Product)
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

// POST: Buat transaksi baru (dengan nama_pembeli, product_id, quantity; fetch harga dari product)
func (ctrl *TransactionController) Create(c *gin.Context) {
	var input struct {
		NamaPembeli string `json:"nama_pembeli"`
		ProductID   int64  `json:"product_id"`
		Quantity    int64  `json:"quantity"`
	}
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

// PATCH: Update partial (quantity atau nama_pembeli; recompute total)
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
	type UpdateInput struct {
		NamaPembeli *string `json:"nama_pembeli"` // Pointer biar optional
		Quantity    *uint   `json:"quantity"`     // Pointer & uint biar match model
	}
	var input UpdateInput
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

// DELETE: Soft delete by ID (asumsi model punya gorm.DeletedAt; kalo hard delete, hapus Unscoped())
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
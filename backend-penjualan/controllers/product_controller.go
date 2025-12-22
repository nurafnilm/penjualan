package controllers

import (
	"net/http"
	"strconv"

	"backend-penjualan/models"
	"errors"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type ProductController struct {
	DB *gorm.DB
}

func NewProductController(db *gorm.DB) *ProductController {
	return &ProductController{DB: db}
}

// GET all products (sama)
func (ctrl *ProductController) GetAll(c *gin.Context) {
	var products []models.Product
	query := ctrl.DB
	if search := c.Query("search"); search != "" {
		query = query.Where("nama ILIKE ?", "%"+search+"%")
	}
	if err := query.Find(&products).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, products)
}

// GET by ID (sama)
func (ctrl *ProductController) GetByID(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID"})
		return
	}
	var product models.Product
	if err := ctrl.DB.First(&product, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "Product not found"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		}
		return
	}
	c.JSON(http.StatusOK, product)
}

// POST new product (FIXED: Validasi manual)
func (ctrl *ProductController) Create(c *gin.Context) {
	var input models.Product
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	// FIXED: Validasi manual
	if input.Nama == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Nama required"})
		return
	}
	if input.Harga <= 0 || input.Harga > 1000000000000 { // Max 1 triliun
		c.JSON(http.StatusBadRequest, gin.H{"error": "Harga positif & max 1 triliun"})
		return
	}
	if err := ctrl.DB.Create(&input).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, input)
}

// PUT update product (FIXED: Validasi manual)
func (ctrl *ProductController) Update(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID"})
		return
	}
	var product models.Product
	if err := ctrl.DB.First(&product, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Product not found"})
		return
	}
	var updates map[string]interface{}
	if err := c.ShouldBindJSON(&updates); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	// FIXED: Validasi manual untuk harga
	if h, ok := updates["harga"].(float64); ok {
		if h <= 0 || h > 1000000000000 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Harga positif & max 1 triliun"})
			return
		}
	}
	if err := ctrl.DB.Model(&product).Updates(updates).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	ctrl.DB.First(&product, id)
	c.JSON(http.StatusOK, product)
}

// DELETE (sama)
func (ctrl *ProductController) Delete(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID"})
		return
	}
	if err := ctrl.DB.Delete(&models.Product{}, id).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Product deleted"})
}
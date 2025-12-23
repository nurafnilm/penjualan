package controllers

import (
	"net/http"
	"strconv"

	"backend-penjualan/models"
	"errors"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// ProductController godoc
// @Description Product controller handles CRUD operations for products
type ProductController struct {
	DB *gorm.DB
}

func NewProductController(db *gorm.DB) *ProductController {
	return &ProductController{DB: db}
}

// GetAll godoc
// @Summary Get all products
// @Description Retrieve list of all products with optional search by name
// @Tags products
// @Accept json
// @Produce json
// @Param search query string false "Search by product name (partial match)"
// @Success 200 {array} models.Product "List of products"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /products [get]
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

// GetByID godoc
// @Summary Get product by ID
// @Description Retrieve a specific product by ID
// @Tags products
// @Accept json
// @Produce json
// @Param id path int true "Product ID"
// @Success 200 {object} models.Product "Product details"
// @Failure 400 {object} map[string]string "Invalid ID"
// @Failure 404 {object} map[string]string "Product not found"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /products/{id} [get]
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

// Create godoc
// @Summary Create a new product
// @Description Create a new product with validation
// @Tags products
// @Accept json
// @Produce json
// @Param product body models.Product true "Product data (nama required, harga positive & max 1T)"
// @Success 201 {object} models.Product "Created product"
// @Failure 400 {object} map[string]string "Validation error"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /products [post]
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

// Update godoc
// @Summary Update a product
// @Description Update a product by ID with partial updates and validation
// @Tags products
// @Accept json
// @Produce json
// @Param id path int true "Product ID"
// @Param updates body object true "Fields to update (e.g., nama, harga)"
// @Success 200 {object} models.Product "Updated product"
// @Failure 400 {object} map[string]string "Invalid ID or validation error"
// @Failure 404 {object} map[string]string "Product not found"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /products/{id} [put]
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

// Delete godoc
// @Summary Delete a product
// @Description Delete a product by ID
// @Tags products
// @Accept json
// @Produce json
// @Param id path int true "Product ID"
// @Success 200 {object} map[string]string "Success message"
// @Failure 400 {object} map[string]string "Invalid ID"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /products/{id} [delete]
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
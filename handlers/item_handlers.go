package handlers

import (
	"invoice-go/models"
	"invoice-go/utils"
	"net/http"
	"strconv"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type ItemHandler struct {
	DB *gorm.DB
}

type CreateItemInput struct {
	Name        	string  `json:"name" binding:"required"`
	Description 	string  `json:"description"`
	UnitPrice       float64 `json:"unit_price" binding:"required,gte=0"`
	Type    		string  `json:"type" binding:"required"`
}

type UpdateItemInput struct {
	Name        	string  `json:"name"`
	Description 	string  `json:"description"`
	UnitPrice       float64 `json:"unit_price" binding:"omitempty,gte=0"`
	Type    		string  `json:"type"`
}

// GetProducts retrieves all products with optional filtering
func (h *ItemHandler) GetItems(c *gin.Context) {
	var products []models.Item
	db := h.DB

	// Apply filters if provided
	if itemType := c.Query("type"); itemType != "" {
		db = db.Where("type = ?", itemType)
	}

	if minPrice := c.Query("min_price"); minPrice != "" {
		if price, err := strconv.ParseFloat(minPrice, 64); err == nil {
			db = db.Where("price >= ?", price)
		}
	}

	if maxPrice := c.Query("max_price"); maxPrice != "" {
		if price, err := strconv.ParseFloat(maxPrice, 64); err == nil {
			db = db.Where("price <= ?", price)
		}
	}

	result := db.Find(&products)
	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve items"})
		return
	}

	c.JSON(http.StatusOK, products)
}

// GetProduct retrieves a single product by ID
func (h *ItemHandler) GetItem(c *gin.Context) {
	id := c.Param("item_id")
	var item models.Item

	result := h.DB.First(&item, id)
	if result.Error != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Item not found"})
		return
	}

	c.JSON(http.StatusOK, item)
}

// CreateProduct adds a new product
func (h *ItemHandler) CreateItem(c *gin.Context) {
	var input CreateItemInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	/** Validate type
	if !isValidType(input.Type) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid type"})
		return
	}
	*/

	item := models.Item{
		Name:        		input.Name,
		Description: 		input.Description,
		UnitPrice:       	input.UnitPrice,
		Type:    			input.Type,
	}

	result := h.DB.Create(&item)
	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create item"})
		return
	}

	c.JSON(http.StatusCreated, item)
}

// UpdateProduct updates an existing product
func (h *ItemHandler) UpdateItem(c *gin.Context) {
	id := c.Param("item_id")
	var item models.Item
	
	if result := h.DB.First(&item, id); result.Error != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Item not found"})
		return
	}

	var input UpdateItemInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Apply updates only for fields that were provided
	updates := make(map[string]interface{})
	
	if input.Name != "" {
		updates["name"] = input.Name
	}
	
	if input.Description != "" {
		updates["description"] = input.Description
	}
	
	if input.UnitPrice != 0 {
		updates["unit_price"] = input.UnitPrice
	}
	
	if input.Type != "" {
		/**if !isValidType(input.Type) {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid type"})
			return
		}
		*/
		updates["type"] = input.Type
	}

	// Apply updates
	if len(updates) > 0 {
		result := h.DB.Model(&item).Updates(updates)
		if result.Error != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update item"})
			return
		}
	}

	// Get updated product
	h.DB.First(&item, id)
	c.JSON(http.StatusOK, item)
}


// Utility function to validate product category
/**
func isValidCategory(category string) bool {
	validCategories := []string{"Electronics", "Apparel", "Footwear", "Furniture", "Appliances"}
	for _, c := range validCategories {
		if category == c {
			return true
		}
	}
	return false
}
*/

// DeleteProduct deletes a product and its associated images
func (h *ItemHandler) DeleteItem(c *gin.Context) {
    // Parse product ID from URL
    ItemID, err := strconv.ParseUint(c.Param("item_id"), 10, 32)
    if err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid item ID"})
        return
    }

    // Check if product exists
    var item models.Item
    if err := h.DB.First(&item, ItemID).Error; err != nil {
        c.JSON(http.StatusNotFound, gin.H{"error": "Item not found"})
        return
    }

    // Delete associated images
    if err := utils.DeleteItemImages(uint(ItemID)); err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete item images"})
        return
    }

    // Delete the product from database
    if err := h.DB.Delete(&item).Error; err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete item"})
        return
    }

    c.JSON(http.StatusOK, gin.H{"message": "Item and associated images deleted successfully"})
}

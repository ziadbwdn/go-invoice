package handlers

import (
	"invoice-go/models"
	"net/http"
	"time"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// test

type OrderHandler struct {
	DB *gorm.DB
}

type OrderItemInput struct {
	ItemID   uint    `json:"item_id" binding:"required"`
	Quantity float64 `json:"quantity" binding:"required,gt=0"`
}

type CreateOrderInput struct {
	CustomerCompanyID uint             `json:"customer_company_id" binding:"required"`
	Items             []OrderItemInput `json:"items" binding:"required,min=1"`
}

// GetOrders retrieves all orders with their items and company details
func (h *OrderHandler) GetOrders(c *gin.Context) {
	var orders []models.Order
	result := h.DB.Preload("OrderItems.Item").
		Preload("CustomerCompany").
		Find(&orders)
	
	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve orders"})
		return
	}
	c.JSON(http.StatusOK, orders)
}

// GetOrder retrieves a single order by ID
func (h *OrderHandler) GetOrder(c *gin.Context) {
	id := c.Param("id")
	var order models.Order
	result := h.DB.Preload("OrderItems.Item").
		Preload("CustomerCompany").
		First(&order, id)
	
	if result.Error != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Order not found"})
		return
	}
	c.JSON(http.StatusOK, order)
}

// CreateOrder handles order creation with transaction
func (h *OrderHandler) CreateOrder(c *gin.Context) {
	var input CreateOrderInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	tx := h.DB.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// Validate company exists
	var company models.Company
	if err := tx.First(&company, input.CustomerCompanyID).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusNotFound, gin.H{"error": "Company not found"})
		return
	}

	var totalPrice float64
	var orderItems []models.OrderItem

	// Process each order item
	for _, itemInput := range input.Items {
		var item models.Item
		if err := tx.First(&item, itemInput.ItemID).Error; err != nil {
			tx.Rollback()
			c.JSON(http.StatusNotFound, gin.H{"error": "Item not found"})
			return
		}

		if item.Stock < int(itemInput.Quantity) {
			tx.Rollback()
			c.JSON(http.StatusBadRequest, gin.H{"error": "Insufficient stock"})
			return
		}

		// Update item stock
		item.Stock -= int(itemInput.Quantity)
		if err := tx.Model(&item).Update("stock", item.Stock).Error; err != nil {
			tx.Rollback()
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update stock"})
			return
		}

		// Calculate item total
		itemTotal := item.UnitPrice * itemInput.Quantity
		totalPrice += itemTotal

		orderItems = append(orderItems, models.OrderItem{
			ItemID:    itemInput.ItemID,
			Quantity:  itemInput.Quantity,
			UnitPrice: item.UnitPrice,
			ItemTotal: itemTotal,
		})
	}

	// Create order
	order := models.Order{
		CustomerCompanyID: input.CustomerCompanyID,
		OrderDate:         time.Now(),
		TotalPrice:        &totalPrice,
		Status:            "pending",
		OrderItems:        orderItems,
	}

	if err := tx.Create(&order).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create order"})
		return
	}

	if err := tx.Commit().Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Transaction failed"})
		return
	}

	// Reload with associations
	h.DB.Preload("OrderItems.Item").Preload("CustomerCompany").First(&order, order.OrderID)
	c.JSON(http.StatusCreated, order)
}

// GetRevenueByType gets revenue statistics grouped by item type
func (h *OrderHandler) GetRevenueByType(c *gin.Context) {
	var results []map[string]interface{}
	
	err := h.DB.Table("order_items").
		Select("items.type as category, SUM(order_items.item_total) as total_revenue, COUNT(*) as order_count").
		Joins("JOIN items ON order_items.item_id = items.item_id").
		Group("items.type").
		Find(&results).Error

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch revenue data"})
		return
	}
	
	c.JSON(http.StatusOK, results)
}
package handlers

import (
    "net/http"
    "github.com/gin-gonic/gin"
    "gorm.io/gorm"
    "invoice-go/models"
	"strings"
	"time"
)

type AddressHandler struct {
	DB *gorm.DB
}

type AddressInput struct {
	CompanyID     uint    `json:"company_id" binding:"required"`
	AddressType   string  `json:"address_type" binding:"required,oneof=billing shipping office other"`
	Street        string  `json:"street" binding:"required"`
	City          string  `json:"city" binding:"required"`
	StateProvince *string `json:"state_province,omitempty"`
	PostalCode    string  `json:"postal_code" binding:"required"`
	Country       string  `json:"country" binding:"required"`
}

// GetAllAddresses returns all addresses (optionally filtered by company)
func (h *AddressHandler) GetAddresses(c *gin.Context) {
	var addresses []models.Address
	query := h.DB.Preload("Company")
	
	if companyID := c.Query("company_id"); companyID != "" {
		query = query.Where("company_id = ?", companyID)
	}

	if err := query.Find(&addresses).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve addresses"})
		return
	}
	c.JSON(http.StatusOK, addresses)
}

// GetAddressByID returns a single address by ID
func (h *AddressHandler) GetAddress(c *gin.Context) {
	id := c.Param("id")
	var address models.Address
	result := h.DB.Preload("Company").First(&address, id)
	
	if result.Error != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Address not found"})
		return
	}
	c.JSON(http.StatusOK, address)
}

// CreateAddress creates a new address
func (h *AddressHandler) CreateAddress(c *gin.Context) {
	var input AddressInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Validate company exists
	var company models.Company
	if result := h.DB.First(&company, input.CompanyID); result.Error != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Company not found"})
		return
	}

	address := models.Address{
		CompanyID:     input.CompanyID,
		AddressType:   input.AddressType,
		Street:        input.Street,
		City:          input.City,
		StateProvince: input.StateProvince,
		PostalCode:    input.PostalCode,
		Country:       input.Country,
	}

	// preventing error
	address.AddressType = strings.Title(strings.ToLower(input.AddressType))
	
	if err := h.DB.Create(&address).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create address"})
		return
	}

	c.JSON(http.StatusCreated, address)
}

// UpdateAddress updates an existing address
func (h *AddressHandler) UpdateAddress(c *gin.Context) {
	id := c.Param("id")
	var address models.Address
	// First retrieve the existing address to preserve timestamps
	if err := h.DB.Select("address_id", "created_at", "updated_at").First(&address, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Address not found"})
		return
	}
	
	// Store original timestamps
	originalCreatedAt := address.CreatedAt
	
	var input models.AddressRequest
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	
	// Start transaction
	tx := h.DB.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()
	
	// Validate company exists if changing company_id
	if input.CompanyID != 0 {
		var company models.Company
		if result := tx.First(&company, input.CompanyID); result.Error != nil {
			tx.Rollback()
			c.JSON(http.StatusBadRequest, gin.H{"error": "Company not found"})
			return
		}
	}
	
	// Prepare StateProvince pointer
	var stateProvincePtr *string
	if input.StateProvince != "" {
		stateProv := input.StateProvince
		stateProvincePtr = &stateProv
	}
	
	// Format address type properly
	addressType := strings.Title(strings.ToLower(input.AddressType))
	
	// Create update map with all fields that should be updated
	updates := map[string]interface{}{
		"company_id":     input.CompanyID,
		"address_type":   addressType,
		"street":         input.Street,
		"city":           input.City,
		"state_province": stateProvincePtr,
		"postal_code":    input.PostalCode,
		"country":        input.Country,
		"updated_at":     time.Now(), // Set current time as updated_at
	}
	
	// Perform update explicitly avoiding created_at
	result := tx.Model(&models.Address{}).
		Where("address_id = ?", id).
		Omit("created_at"). // Important: don't touch created_at
		Updates(updates)
	
	if result.Error != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to update address",
			"details": result.Error.Error(),
		})
		return
	}
	
	if result.RowsAffected == 0 {
		tx.Rollback()
		c.JSON(http.StatusConflict, gin.H{"error": "No changes detected"})
		return
	}
	
	// Commit transaction
	if err := tx.Commit().Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Transaction failed"})
		return
	}
	
	// Fetch the updated address
	var updatedAddress models.Address
	h.DB.First(&updatedAddress, id)
	
	// Ensure the original created_at is preserved
	updatedAddress.CreatedAt = originalCreatedAt
	
	c.JSON(http.StatusOK, updatedAddress)
}
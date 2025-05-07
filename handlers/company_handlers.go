package handlers

import (
	"net/http"
	"github.com/gin-gonic/gin"
	"time"
	"gorm.io/gorm"
	// "invoice-go/utils"
	"invoice-go/models"
)

type CompanyHandler struct {
	DB *gorm.DB
}

type CreateCompanyInput struct {
	CompanyName              string  `json:"company_name" binding:"required"`
	ContactPerson            *string `json:"contact_person,omitempty"`
	Email                    *string `json:"email,omitempty"`
	Phone                    *string `json:"phone,omitempty"`
	IsCustomer               bool    `json:"is_customer"`
	IsVendor                 bool    `json:"is_vendor"`
	DefaultBillingAddressID  *uint   `json:"default_billing_address_id,omitempty"`
	DefaultShippingAddressID *uint   `json:"default_shipping_address_id,omitempty"`
}

type UpdateCompanyInput struct {
	CompanyName              *string `json:"company_name,omitempty"`
	ContactPerson            *string `json:"contact_person,omitempty"`
	Email                    *string `json:"email,omitempty" binding:"omitempty,email"`
	Phone                    *string `json:"phone,omitempty" binding:"omitempty,min=7,max=20"`
	IsCustomer               *bool   `json:"is_customer,omitempty"`
	IsVendor                 *bool   `json:"is_vendor,omitempty"`
	DefaultBillingAddressID  *uint   `json:"default_billing_address_id,omitempty"`
	DefaultShippingAddressID *uint   `json:"default_shipping_address_id,omitempty"`
}

// GetAllCompanies returns all companies
func (h *CompanyHandler) GetAllCompanies(c *gin.Context) {
	var companies []models.Company
	result := h.DB.Preload("DefaultBillingAddress").
		Preload("DefaultShippingAddress").
		Find(&companies)

	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve companies"})
		return
	}
	c.JSON(http.StatusOK, companies)
}

// GetCompanyByID returns a single company by ID
func (h *CompanyHandler) GetCompanyByID(c *gin.Context) {
	id := c.Param("id")
	var company models.Company
	result := h.DB.Preload("DefaultBillingAddress").
		Preload("DefaultShippingAddress").
		Preload("Addresses").
		First(&company, id)

	if result.Error != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Company not found"})
		return
	}
	c.JSON(http.StatusOK, company)
}

// CreateCompany creates a new company
func (h *CompanyHandler) CreateCompany(c *gin.Context) {

	// v02

	var input models.CompanyRequest
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Convert strings to pointers for nullable fields
	var contactPersonPtr *string
	var emailPtr *string
	var phonePtr *string
	
	if input.ContactPerson != "" {
		contactPersonPtr = &input.ContactPerson
	}
	
	if input.Email != "" {
		emailPtr = &input.Email
	}
	
	if input.Phone != "" {
		phonePtr = &input.Phone
	}

	company := models.Company{
		CompanyName:              input.CompanyName,
		ContactPerson:            contactPersonPtr,
		Email:                    emailPtr,
		Phone:                    phonePtr,
		IsCustomer:               input.IsCustomer,
		IsVendor:                 input.IsVendor,
		DefaultBillingAddressID:  nil,
		DefaultShippingAddressID: nil,
	}

	tx := h.DB.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// Validate default addresses if provided

	/**
	if !utils.ValidateAddress(tx, c, company.DefaultBillingAddressID, nil, "Invalid billing address ID") ||
		!utils.ValidateAddress(tx, c, company.DefaultShippingAddressID, nil, "Invalid shipping address ID") {
		return // ValidateAddress handles rollback and error response
	}

	if err := tx.Create(&company).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create company"})
		return
	}

	if err := tx.Commit().Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Transaction failed"})
		return
	}
	*/

	if err := h.DB.Create(&company).Error; err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{
            "error": "Failed to create company",
            "details": err.Error(),
        })
        return
    }

	c.JSON(http.StatusCreated, company)
}

// UpdateCompany updates an existing company
func (h *CompanyHandler) UpdateCompany(c *gin.Context) {
	id := c.Param("id")
	var company models.Company

	// Use Select to explicitly choose updatable fields
    // First get existing timestamps
    if err := h.DB.Select("created_at", "updated_at").First(&company, id).Error; err != nil {
        c.JSON(http.StatusNotFound, gin.H{"error": "Company not found"})
        return
    }

    var input UpdateCompanyInput
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

    // Update fields but preserve original timestamps
    updates := map[string]interface{}{
        "company_name":               input.CompanyName,
        "contact_person":             input.ContactPerson,
        "email":                      input.Email,
        "phone":                      input.Phone,
        "is_customer":                input.IsCustomer,
        "is_vendor":                  input.IsVendor,
        "default_billing_address_id": input.DefaultBillingAddressID,
        "default_shipping_address_id": input.DefaultShippingAddressID,
        "updated_at":                 time.Now().Format("2006-01-02 15:04:05"), // MySQL format
    }

    // Remove nil values from updates
    cleanUpdates := make(map[string]interface{})
    for k, v := range updates {
        if v != nil {
            cleanUpdates[k] = v
        }
    }

    // Explicit update with timestamp control
    result := tx.Model(&models.Company{}).
        Where("company_id = ?", id).
        Omit("created_at").
        Updates(cleanUpdates)

    if result.Error != nil {
        tx.Rollback()
        c.JSON(http.StatusInternalServerError, gin.H{
            "error": "Database update failed",
            "details": result.Error.Error(),
        })
        return
    }

    if result.RowsAffected == 0 {
        tx.Rollback()
        c.JSON(http.StatusConflict, gin.H{"error": "No changes detected"})
        return
    }

    if err := tx.Commit().Error; err != nil {
        tx.Rollback()
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Transaction commit failed"})
        return
    }

    // Return updated company with original created_at
    var updatedCompany models.Company
    h.DB.First(&updatedCompany, id)
    updatedCompany.CreatedAt = company.CreatedAt // Preserve original created_at
    
    c.JSON(http.StatusOK, updatedCompany)
}

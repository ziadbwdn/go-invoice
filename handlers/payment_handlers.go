package handlers

import (
	"net/http"
	"strconv"
	"time"
	"invoice-go/models"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type PaymentHandler struct {
	DB *gorm.DB
}

// PaymentRequest is used for creating a payment
type PaymentRequest struct {
	InvoiceID           	uint      	`json:"invoice_id" binding:"required"`
	PaymentDate         	time.Time 	`json:"payment_date" binding:"required"`
	Amount              	float64   	`json:"amount" binding:"required,min=0.01"`
	Method              	string    	`json:"method" binding:"required,min=1,max=50"`
	TransactionReference 	string    	`json:"transaction_reference" binding:"max=100"`
}


// POST /payment/:id – record a new payment
func (h *PaymentHandler) CreatePayment(c *gin.Context) {
    // 1. Parse invoice ID from path
    invoiceID, err := strconv.ParseUint(c.Param("id"), 10, 64)
    if err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": "invalid invoice ID"})
        return
    }

    // 2. Bind and validate JSON payload
    var input struct {
        Amount  float64 `json:"amount" binding:"required"`
        Method  string  `json:"method"  binding:"required"`
        Status  string  `json:"status"  binding:"required"`
        Ref     *string `json:"transaction_reference"`
    }
    if err := c.ShouldBindJSON(&input); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request payload"})
        return
    }
    if input.Amount <= 0 {
        c.JSON(http.StatusBadRequest, gin.H{"error": "amount must be positive"})
        return
    }

    // 3. Build the model
    payment := models.Payment{
        InvoiceID:           	uint(invoiceID),
        PaymentDate:         	time.Now(),
        Amount:              	input.Amount,
        Method:              	&input.Method,
        Status:              	input.Status,
        TransactionReference: 	input.Ref,
        CreatedAt:           	time.Now(),
    }

    // 4. Persist to DB
    if err := h.DB.Create(&payment).Error; err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to record payment"})
        return
    }

    // 5. Return success
    c.JSON(http.StatusCreated, gin.H{"payment": payment})
}

// GET /payment/:id[?status=…] – list payments for an invoice, optionally filtering by status
func (h *PaymentHandler) GetPayments(c *gin.Context) {
    // 1. Parse invoice ID
    invoiceID, err := strconv.ParseUint(c.Param("id"), 10, 64)
    if err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": "invalid invoice ID"})
        return
    }

    // 2. Optional status filter
    status := c.Query("status")

    // 3. Build and execute query
    var payments []models.Payment
    q := h.DB.Where("invoice_id = ?", invoiceID)
    if status != "" {
        q = q.Where("status = ?", status)
    }
    if err := q.Find(&payments).Error; err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch payments"})
        return
    }

    // 4. Return list
    c.JSON(http.StatusOK, gin.H{"payments": payments})
}

// GET /payment/:id/details – fetch one payment by its ID
func (h *PaymentHandler) GetPaymentDetails(c *gin.Context) {
    // 1. Parse payment ID
    paymentID, err := strconv.ParseUint(c.Param("id"), 10, 64)
    if err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": "invalid payment ID"})
        return
    }

    // 2. Retrieve
    var payment models.Payment
    if err := h.DB.First(&payment, paymentID).Error; err != nil {
        if err == gorm.ErrRecordNotFound {
            c.JSON(http.StatusNotFound, gin.H{"error": "payment not found"})
        } else {
            c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch payment details"})
        }
        return
    }

    // 3. Return detail
    c.JSON(http.StatusOK, gin.H{"payment": payment})
}

// GET /payment/:id/details – fetch one payment by its ID
func (h *PaymentHandler) GetPaymentStatus(c *gin.Context) {
    // 1. Parse payment ID
    paymentID, err := strconv.ParseUint(c.Param("id"), 10, 64)
    if err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": "invalid payment ID"})
        return
    }

    // 2. Retrieve
    var payment models.Payment
    if err := h.DB.First(&payment, paymentID).Error; err != nil {
        if err == gorm.ErrRecordNotFound {
            c.JSON(http.StatusNotFound, gin.H{"error": "payment not found"})
        } else {
            c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch payment details"})
        }
        return
    }

    // 3. Return detail
    c.JSON(http.StatusOK, gin.H{"payment": payment})
}


// PUT /payment/:id/status - update payment status
// PUT /payment/:id/status – update only the status of a payment
func (h *PaymentHandler) UpdatePaymentStatus(c *gin.Context) {
    // 1. Parse payment ID
    paymentID, err := strconv.ParseUint(c.Param("id"), 10, 64)
    if err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": "invalid payment ID"})
        return
    }

    // 2. Bind new status
    var payload struct {
        Status string `json:"status" binding:"required"`
    }
    if err := c.ShouldBindJSON(&payload); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": "status field is required"})
        return
    }
    if payload.Status == "" {
        c.JSON(http.StatusBadRequest, gin.H{"error": "status cannot be empty"})
        return
    }

    // 3. Update and fetch updated record in a transaction
    var updated models.Payment
    err = h.DB.Transaction(func(tx *gorm.DB) error {
        // 3a. Update
        result := tx.Model(&models.Payment{}).
            Where("payment_id = ?", paymentID).
            Updates(map[string]interface{}{
                "status":     payload.Status,
                "updated_at": time.Now(),
            })
        if result.Error != nil {
            return result.Error
        }
        if result.RowsAffected == 0 {
            return gorm.ErrRecordNotFound
        }
        // 3b. Read back
        return tx.First(&updated, paymentID).Error
    })
    if err != nil {
        if err == gorm.ErrRecordNotFound {
            c.JSON(http.StatusNotFound, gin.H{"error": "payment not found"})
        } else {
            c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update status"})
        }
        return
    }

    // 4. Return updated resource
    c.JSON(http.StatusOK, gin.H{"payment": updated})
}

package handlers

import (
	"net/http"
	"time"
	"strconv"
	"invoice-go/database"
	"invoice-go/models"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type InvoiceHandler struct {
    DB *gorm.DB
}

// InvoiceRequest is used for creating an invoice
type InvoiceRequest struct {
	SenderCompanyID    uint      `gorm:"type:int unsigned;column:sender_company_id;not null" json:"sender_company_id"`
	RecipientCompanyID uint      `gorm:"type:int unsigned;column:recipient_company_id;not null;index" json:"recipient_company_id"`
	BillingAddressID   uint      `gorm:"type:int unsigned;column:billing_address_id;not null" json:"billing_address_id"`
	ShippingAddressID  *uint     `gorm:"type:int unsigned;column:shipping_address_id" json:"shipping_address_id,omitempty"`
	OrderID            *uint     `gorm:"type:int unsigned;column:order_id" json:"order_id,omitempty"`
	InvoiceNumber      string    `json:"invoice_number" binding:"required,min=1,max=50"`
	InvoiceDate        time.Time `json:"invoice_date" binding:"required"`
	DueDate            time.Time `json:"due_date" binding:"required"`
	InvoiceSubject     *string    `json:"invoice_subject" binding:"max=200"`
	Notes              *string    `json:"notes" binding:"max=500"`
}

type InvoiceReportResponse struct {
	Invoice          models.Invoice       `json:"invoice"`
	SenderCompany    models.Company       `json:"sender_company"`
	RecipientCompany models.Company       `json:"recipient_company"`
	BillingAddress   models.Address       `json:"billing_address"`
	ShippingAddress  models.Address       `json:"shipping_address"`
	Order            models.Order         `json:"order"`
	Items            []models.InvoiceItem `json:"items"`
	Payments         []models.Payment     `json:"payments"`
	PaymentStatus    string        `json:"payment_status"`
}

// POST /invoice/:id - create new invoice
func (h *InvoiceHandler) CreateInvoice(c *gin.Context) {
    // 1. Parse path param (e.g., maybe this is an order ID or parent resource)
    _, err := strconv.ParseUint(c.Param("id"), 10, 64)
    if err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": "invalid invoice ID"})
        return
    }

    // 2. Bind and validate request body
    var input InvoiceRequest
    if err := c.ShouldBindJSON(&input); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request payload"})
        return
    }
    // TODO: parse dates (invoice_date, due_date) properly – here we just set now
    now := time.Now()
    inv := models.Invoice{
        SenderCompanyID:    input.SenderCompanyID,
        RecipientCompanyID: input.RecipientCompanyID,
        BillingAddressID:   input.BillingAddressID,
        ShippingAddressID:  input.ShippingAddressID,
        OrderID:            input.OrderID,
        InvoiceNumber:      input.InvoiceNumber,
        InvoiceDate:        now,
        DueDate:            now.AddDate(0, 0, 30),
        InvoiceSubject:     input.InvoiceSubject,
        Notes:              input.Notes,
        Subtotal:           0,
        TaxTotal:           0,
        GrandTotal:         0,
        AmountPaid:         0,
        AmountDue:          0,
        Status:             "unpaid",
        CreatedAt:          now,
        UpdatedAt:          now,
    }

    if err := h.DB.Create(&inv).Error; err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create invoice"})
        return
    }
    c.JSON(http.StatusCreated, gin.H{"invoice": inv})
}

// GET /invoice/:id - filter invoices by status
// GET /invoice/:id[?status=…] – fetch invoices filtered by status, or single by ID if no status
func (h *InvoiceHandler) GetInvoices(c *gin.Context) {
    idStr := c.Param("id")
    status := c.Query("status")

    // If status query is present, ignore the path‐ID and filter all invoices
    if status != "" {
        var list []models.Invoice
        if err := h.DB.Where("status = ?", status).Find(&list).Error; err != nil {
            c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch invoices"})
            return
        }
        c.JSON(http.StatusOK, gin.H{"invoices": list})
        return
    }

    // Otherwise treat path ID as invoice_id
    invID, err := strconv.ParseUint(idStr, 10, 64)
    if err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": "invalid invoice ID"})
        return
    }
    var inv models.Invoice
    if err := h.DB.First(&inv, invID).Error; err != nil {
        if err == gorm.ErrRecordNotFound {
            c.JSON(http.StatusNotFound, gin.H{"error": "invoice not found"})
        } else {
            c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to retrieve invoice"})
        }
        return
    }
    c.JSON(http.StatusOK, gin.H{"invoice": inv})
}

// GET /invoice/:id/details - get detailed invoice data
func (h *InvoiceHandler) GetInvoiceDetails(c *gin.Context) {
    idStr := c.Param("id")
    status := c.Query("status")

    // If status query is present, ignore the path‐ID and filter all invoices
    if status != "" {
        var list []models.Invoice
        if err := h.DB.Where("status = ?", status).Find(&list).Error; err != nil {
            c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch invoices"})
            return
        }
        c.JSON(http.StatusOK, gin.H{"invoices": list})
        return
    }

    // Otherwise treat path ID as invoice_id
    invID, err := strconv.ParseUint(idStr, 10, 64)
    if err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": "invalid invoice ID"})
        return
    }
    // Use your existing queries.GenerateInvoiceReport to fetch everything
    report, err := database.GenerateInvoiceDetails(h.DB, uint(invID))
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to generate invoice details"})
        return
    }
    c.JSON(http.StatusOK, report)
}

// GET /invoice/:id/reports - get detailed invoice data
func (h *InvoiceHandler) GetInvoiceReports (c *gin.Context) {
    idStr := c.Param("id")
    status := c.Query("status")

    // If status query is present, ignore the path‐ID and filter all invoices
    if status != "" {
        var list []models.Invoice
        if err := h.DB.Where("status = ?", status).Find(&list).Error; err != nil {
            c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch invoices"})
            return
        }
        c.JSON(http.StatusOK, gin.H{"invoices": list})
        return
    }

    // Otherwise treat path ID as invoice_id
    invID, err := strconv.ParseUint(idStr, 10, 64)
    if err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": "invalid invoice ID"})
        return
    }

    // Use your existing queries.GenerateInvoiceReport to fetch everything
    report, err := database.GenerateInvoiceReport(h.DB, uint(invID))
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to generate invoice reports"})
        return
    }
    c.JSON(http.StatusOK, report)
}

// handlers/invoice_handlers.go

func (h *InvoiceHandler) GetInvoiceStatus(c *gin.Context) {
    // Parse invoice ID from URL
    idParam := c.Param("id")
    invID, err := strconv.ParseUint(idParam, 10, 64)
    if err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": "invalid invoice ID"})
        return
    }

    // Call business logic
    status, due, err := database.GetPaymentStatus(h.DB, uint(invID))
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get payment status"})
        return
    }

    // Return JSON
    c.JSON(http.StatusOK, gin.H{
        "status": status,
        "due":    due,
    })
}


// PUT /invoice/:id/status – update only the invoice status
func (h *InvoiceHandler) UpdateInvoiceStatus(c *gin.Context) {
    invID, err := strconv.ParseUint(c.Param("id"), 10, 64)
    if err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": "invalid invoice ID"})
        return
    }

    var payload struct {
        Status string `json:"status" binding:"required"`
    }
    if err := c.ShouldBindJSON(&payload); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": "status is required"})
        return
    }

    var updated models.Invoice
    err = h.DB.Transaction(func(tx *gorm.DB) error {
        res := tx.Model(&models.Invoice{}).
            Where("invoice_id = ?", invID).
            Updates(map[string]interface{}{
                "status":     payload.Status,
                "updated_at": time.Now(),
            })
        if res.Error != nil {
            return res.Error
        }
        if res.RowsAffected == 0 {
            return gorm.ErrRecordNotFound
        }
        return tx.First(&updated, invID).Error
    })
    if err != nil {
        if err == gorm.ErrRecordNotFound {
            c.JSON(http.StatusNotFound, gin.H{"error": "invoice not found"})
        } else {
            c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update invoice status"})
        }
        return
    }

    c.JSON(http.StatusOK, gin.H{"invoice": updated})
}

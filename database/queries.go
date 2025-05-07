package database

import (
	"fmt"
	"invoice-go/models"
	"gorm.io/gorm"
	"time"
)

// GetInvoiceAddressExtraction retrieves address information for an invoice
func GetInvoiceAddressExtraction(db *gorm.DB, invoiceID uint) (*models.Invoice, error) {
	var invoice models.Invoice
	
	// Preload relevant address information
	err := db.Preload("BillingAddress").
		Preload("ShippingAddress").
		Preload("SenderCompany").
		Preload("RecipientCompany").
		Where("invoice_id = ?", invoiceID).
		First(&invoice).Error
	
	if err != nil {
		return nil, fmt.Errorf("failed to extract invoice address information: %w", err)
	}
	
	return &invoice, nil
}

// GetTransactionDetails (v02)
func GetTransactionDetails(db *gorm.DB, invoiceID uint) (*models.Invoice, error) {
	var invoice models.Invoice
	
	// Preload all relevant transaction information
	err := db.Preload("InvoiceItems.Item").
		Preload("Order.OrderItems.Item").
		Where("invoice_id = ?", invoiceID).
		First(&invoice).Error
	
	if err != nil {
		return nil, fmt.Errorf("failed to get transaction details: %w", err)
	}
	
	return &invoice, nil
}

// GetSubtotalCalculation (v02)
func GetSubtotalCalculation(db *gorm.DB, invoiceID uint) (float64, error) {
	var subtotal float64
	
	// Calculate the sum of all invoice items
	err := db.Model(&models.InvoiceItem{}).
		Select("SUM(item_total)").
		Where("invoice_id = ?", invoiceID).
		Scan(&subtotal).Error
	
	if err != nil {
		return 0, fmt.Errorf("failed to calculate subtotal: %w", err)
	}
	
	return subtotal, nil
}

// GetPaymentStatus returns the payment status and amount due for an invoice - V02
func GetPaymentStatus(db *gorm.DB, invoiceID uint) (string, float64, error) {
	var invoice models.Invoice
	var totalPaid float64
	
	// Get the invoice
	if err := db.Where("invoice_id = ?", invoiceID).First(&invoice).Error; err != nil {
		return "", 0, fmt.Errorf("failed to find invoice: %w", err)
	}
	
	// Calculate total payments
	if err := db.Model(&models.Payment{}).
		Select("COALESCE(SUM(amount), 0)").
		Where("invoice_id = ? AND status = 'completed'", invoiceID).
		Scan(&totalPaid).Error; err != nil {
		return "", 0, fmt.Errorf("failed to calculate payments: %w", err)
	}
	
	// Calculate amount due
	amountDue := invoice.GrandTotal - totalPaid
	
	// Determine payment status
	var status string
	switch {
	case amountDue <= 0:
		status = "paid"
	case totalPaid > 0:
		status = "partial"
	default:
		status = "unpaid"
	}
	
	// Update invoice payment fields if they're out of sync
	if totalPaid != invoice.AmountPaid || amountDue != invoice.AmountDue || status != invoice.Status {
		db.Model(&invoice).Updates(map[string]interface{}{
			"amount_paid": totalPaid,
			"amount_due":  amountDue,
			"status":      status,
			"updated_at":  time.Now(),
		})
	}
	
	return status, amountDue, nil
}

// Invoice Details Function
func GenerateInvoiceDetails(db *gorm.DB, invoiceID uint) ([]models.InvoiceDetail, error) {
    var details []models.InvoiceDetail
    query := `
        WITH details AS (
            SELECT
                ii.invoice_id,
                ii.invoice_item_id,
                itm.name,
                ii.description,
                ii.quantity,
                ii.unit_price,
                ii.item_total,
                ii.tax_rate_percentage
            FROM
                invoice_items AS ii
            JOIN
                items AS itm ON ii.item_id = itm.item_id
            WHERE
                ii.invoice_id = ?
        )
        SELECT
            i.invoice_id,
            i.status,
            d.invoice_item_id,
            d.name,
            d.description,
            d.quantity,
            d.unit_price,
            d.item_total,
            d.tax_rate_percentage
        FROM
            details AS d
        JOIN
            invoices i ON d.invoice_id = i.invoice_id`
    if err := db.Raw(query, invoiceID).Scan(&details).Error; err != nil {
        return nil, fmt.Errorf("failed to fetch invoice details: %w", err)
    }
    return details, nil
}


func GenerateInvoiceReport(db *gorm.DB, invoiceID uint) (*models.InvoiceReportResponse, error) {
    var report models.InvoiceReportResponse

    // Fetch basic invoice data with addresses
    invoiceAddress, err := GetInvoiceAddressExtraction(db, invoiceID)
    if err != nil {
        return nil, err
    }

    // Fetch transaction details (if necessary)
    if _, err := GetTransactionDetails(db, invoiceID); err != nil {
        return nil, err
    }

    // Calculate subtotal
    if _, err := GetSubtotalCalculation(db, invoiceID); err != nil {
        return nil, err
    }

    // Get payment status
    paymentStatus, _, err := GetPaymentStatus(db, invoiceID)
    if err != nil {
        return nil, err
    }

    // Fetch sender and recipient companies
    var senderCompany, recipientCompany models.Company
    if err := db.First(&senderCompany, "company_id = ?", invoiceAddress.SenderCompanyID).Error; err != nil {
        return nil, fmt.Errorf("sender company not found: %w", err)
    }
    if err := db.First(&recipientCompany, "company_id = ?", invoiceAddress.RecipientCompanyID).Error; err != nil {
        return nil, fmt.Errorf("recipient company not found: %w", err)
    }

    // Fetch order details
    var order models.Order
    if err := db.Preload("OrderItems.Item").First(&order, "order_id = ?", invoiceAddress.OrderID).Error; err != nil {
        return nil, fmt.Errorf("order details not found: %w", err)
    }

    // Fetch invoice items using optimized query
    details, err := GenerateInvoiceDetails(db, invoiceID)
    if err != nil {
        return nil, fmt.Errorf("invoice items fetch failed: %w", err)
    }

    // Convert details to InvoiceItem structs
    var invoiceItems []models.InvoiceItem
    for _, d := range details {
        item := &models.Item{Name: d.ItemName} // Populate item name
        invoiceItems = append(invoiceItems, models.InvoiceItem{
            InvoiceItemID:     d.InvoiceItemID,
            InvoiceID:         d.InvoiceID,
            Description:       d.Description,
            Quantity:          d.Quantity,
            UnitPrice:         d.UnitPrice,
            ItemTotal:         d.ItemTotal,
            TaxRatePercentage: d.TaxRatePercentage,
            Item:              item,
        })
    }

    // Fetch payments
    var payments []models.Payment
    if err := db.Where("invoice_id = ?", invoiceID).Find(&payments).Error; err != nil {
        return nil, fmt.Errorf("payments fetch failed: %w", err)
    }

    // Populate the final report
    report = models.InvoiceReportResponse{
        Invoice:          *invoiceAddress,
        SenderCompany:    senderCompany,
        RecipientCompany: recipientCompany,
        BillingAddress:   &invoiceAddress.BillingAddress,
        ShippingAddress:  invoiceAddress.ShippingAddress,
        Order:            order,
        Items:            invoiceItems,
        Payments:         payments,
        PaymentStatus:    paymentStatus,
    }

    return &report, nil
}
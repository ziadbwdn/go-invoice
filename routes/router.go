// router.go
package routes

import (
	"invoice-go/handlers"
	"invoice-go/utils"
	"net/http"
	"log"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type InvoiceDetail struct {
    InvoiceID         uint    `gorm:"column:invoice_id"`
    Status            string  `gorm:"column:status"`
    InvoiceItemID     uint    `gorm:"column:invoice_item_id"`
    ItemName          string  `gorm:"column:name"`
    Description       string  `gorm:"column:description"`
    Quantity          float64 `gorm:"column:quantity"`
    UnitPrice         float64 `gorm:"column:unit_price"`
    ItemTotal         float64 `gorm:"column:item_total"`
    TaxRatePercentage float64 `gorm:"column:tax_rate_percentage"`
}

// SetupRouter configures all the routes for our application
func SetupRouter(db *gorm.DB) *gin.Engine {
	r := gin.Default()

	// Middleware for CORS
	r.Use(func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, PATCH, OPTIONS")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
		
		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}
		
		c.Next()
	})

	// Initialize handlers
	itemHandler := &handlers.ItemHandler{DB: db} 
	orderHandler := &handlers.OrderHandler{DB: db}
	imageHandler := &handlers.ImageHandler{DB: db}
	companyHandler := &handlers.CompanyHandler{DB: db}
	addressHandler := &handlers.AddressHandler{DB: db}
	invoiceHandler := &handlers.InvoiceHandler{DB: db}
	paymentHandler := &handlers.PaymentHandler{DB: db}

	// Static file serving
	r.Static("/uploads", "./uploads")

	// Health check endpoint
	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "up"})
	})

	// test upload
	r.GET("/test-upload", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "Test upload endpoint working"})
	})

	// Product routes - nanti yang ini diganti
	itemRoutes := r.Group("/items") 
	{
		itemRoutes.GET("", itemHandler.GetItems) 
		itemRoutes.GET("/:id", itemHandler.GetItem) 
		itemRoutes.POST("", itemHandler.CreateItem) 
		itemRoutes.PUT("/:id", itemHandler.UpdateItem) 

	// Image upload/download routes with path traversal protection
		itemRoutes.POST("/:id/upload", utils.PathTraversalMiddleware(), imageHandler.UploadItemImage) 
		itemRoutes.GET("/:id/image", utils.PathTraversalMiddleware(), imageHandler.DownloadItemImage) 
	}
  
	// Company routes
	companyRoutes := r.Group("/companies")
	{
		companyRoutes.GET("", companyHandler.GetAllCompanies)
		companyRoutes.GET("/:id", companyHandler.GetCompanyByID)
		companyRoutes.POST("", companyHandler.CreateCompany)
		companyRoutes.PUT("/:id", companyHandler.UpdateCompany)
	}

	// Address routes
	addressRoutes := r.Group("/addresses")
	{
		addressRoutes.GET("", addressHandler.GetAddresses)
		addressRoutes.GET("/:id", addressHandler.GetAddress)
		addressRoutes.POST("", addressHandler.CreateAddress)
		addressRoutes.PUT("/:id", addressHandler.UpdateAddress)
	}


	// Order routes
	orderRoutes := r.Group("/orders")
	{
		orderRoutes.GET("", orderHandler.GetOrders)
		orderRoutes.GET("/:id", orderHandler.GetOrder)
		orderRoutes.POST("", orderHandler.CreateOrder)
		// orderRoutes.GET("/revenue", orderHandler.GetRevenueByCategory)
	}

	invoices := r.Group("/invoice")
	{
		invoices.POST("/:id",          invoiceHandler.CreateInvoice)
		invoices.GET("/:id",           invoiceHandler.GetInvoices)
		invoices.GET("/:id/details",   invoiceHandler.GetInvoiceDetails)
		invoices.GET("/:id/reports",   invoiceHandler.GetInvoiceReports)
		invoices.GET("/:id/status",   invoiceHandler.GetInvoiceStatus)
		invoices.PATCH("/:id/status",    invoiceHandler.UpdateInvoiceStatus)
	}

	payments := r.Group("/payment")
	{
		payments.POST("/:id",       paymentHandler.CreatePayment)
		payments.GET("/:id",        paymentHandler.GetPayments)
		payments.GET("/:id/details",paymentHandler.GetPaymentDetails)
		payments.PUT("/:id/status", paymentHandler.UpdatePaymentStatus)
	}


	// Debug: Print all registered routes
    for _, route := range r.Routes() {
        log.Printf("Route registered: %s %s", route.Method, route.Path)
    }

	return r
}
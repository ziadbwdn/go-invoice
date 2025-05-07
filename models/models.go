package models

import (
	"time"
)

// Company represents the companies table.
type Company struct {
    CompanyID                uint     `gorm:"primaryKey;autoIncrement;column:company_id" json:"company_id"`
    CompanyName              string   `gorm:"column:company_name;not null;index" json:"company_name"`
    ContactPerson            *string  `gorm:"column:contact_person" json:"contact_person,omitempty"`
    Email                    *string  `gorm:"column:email;unique" json:"email,omitempty"`
    Phone                    *string  `gorm:"column:phone" json:"phone,omitempty"`
    IsCustomer               bool     `gorm:"column:is_customer;not null;default:false;index" json:"is_customer"`
    IsVendor                 bool     `gorm:"column:is_vendor;not null;default:false;index" json:"is_vendor"`
    DefaultBillingAddressID  *uint    `gorm:"column:default_billing_address_id" json:"default_billing_address_id,omitempty"`
    DefaultShippingAddressID *uint    `gorm:"column:default_shipping_address_id" json:"default_shipping_address_id,omitempty"`
    CreatedAt                time.Time `gorm:"column:created_at;autoCreateTime" json:"created_at"`
    UpdatedAt                time.Time `gorm:"column:updated_at;autoUpdateTime" json:"updated_at"`

    // Relations (without creating additional foreign keys in DB)
    /**
    DefaultBillingAddress    *Address  `gorm:"foreignKey:AddressID;references:DefaultBillingAddressID;constraint:false" json:"default_billing_address,omitempty"`
    DefaultShippingAddress   *Address  `gorm:"foreignKey:AddressID;references:DefaultShippingAddressID;constraint:false" json:"default_shipping_address,omitempty"`
    */
    DefaultBillingAddress    *Address  `gorm:"foreignKey:DefaultBillingAddressID;references:AddressID;constraint:false" json:"default_billing_address,omitempty"`
    DefaultShippingAddress   *Address  `gorm:"foreignKey:DefaultShippingAddressID;references:AddressID;constraint:false" json:"default_shipping_address,omitempty"`
    Addresses              []Address   `gorm:"foreignKey:CompanyID;references:CompanyID;constraint:false  json:"addresses,omitempty"`
}

// Address represents the unified addresses table.
type Address struct {
    AddressID     uint      `gorm:"primaryKey;autoIncrement;column:address_id" json:"address_id"`
    CompanyID     uint      `gorm:"column:company_id;not null;index" json:"company_id"`
    AddressType   string    `gorm:"column:address_type;not null" json:"address_type"` // e.g., Billing, Shipping, Office, Other
    Street        string    `gorm:"column:street;not null" json:"street"`
    City          string    `gorm:"type:varchar(200);not null"`
    StateProvince *string   `gorm:"column:state_province" json:"state_province,omitempty"`
    PostalCode    string    `gorm:"column:postal_code;not null" json:"postal_code"`
    Country       string    `gorm:"type:varchar(200);not null"`
    CreatedAt     time.Time `gorm:"column:created_at;autoCreateTime" json:"created_at"`
    UpdatedAt     time.Time `gorm:"column:updated_at;autoUpdateTime" json:"updated_at"`
    // If needed, you can preload the corresponding company:
    Company       Company   `gorm:"foreignKey:CompanyID;references:CompanyID;constraint:OnDelete:CASCADE" json:"company"`
}

type Item struct {
    ItemID      uint		`json:"item_id" gorm:"primaryKey;type:int unsigned"`
    Name        string    	`json:"name" gorm:"type:varchar(100);not null"`
    Description string    	`json:"description" gorm:"type:text"`
    UnitPrice   float64   	`json:"unit_price" gorm:"type:decimal(10,2);not null"`
	Type    	string    	`json:"category" gorm:"size:50;not null"`
    Stock       int       	`json:"stock" gorm:"not null;default:0"`
    ImagePath   string    	`json:"image_path" gorm:"type:varchar(255)"`
    CreatedAt   time.Time 	`json:"created_at"`
    UpdatedAt   time.Time 	`json:"updated_at"`
}


// Order represents a customer order for a specific product
type Order struct {
    OrderID           uint       `gorm:"primaryKey;autoIncrement;column:order_id" json:"order_id"`
    CustomerCompanyID uint       `gorm:"column:customer_company_id;not null;index" json:"customer_company_id"`
    OrderDate         time.Time  `gorm:"column:order_date;not null" json:"order_date"` // Use DATE type mapping as needed
    TotalPrice        *float64   `gorm:"column:total_price" json:"total_price,omitempty"`
    Status            string     `gorm:"column:status;not null;default:'Pending';index" json:"status"`
    CreatedAt         time.Time  `gorm:"column:created_at;autoCreateTime" json:"created_at"`
    UpdatedAt         time.Time  `gorm:"column:updated_at;autoUpdateTime" json:"updated_at"`
    // Associations
    OrderItems        []OrderItem `gorm:"foreignKey:OrderID" json:"order_items"`
    CustomerCompany   Company     `gorm:"foreignKey:CustomerCompanyID;references:CompanyID" json:"customer"`
}

// OrderItem represents the order_items table.
type OrderItem struct {
    OrderItemID uint     `gorm:"primaryKey;autoIncrement;column:order_item_id" json:"order_item_id"`
    OrderID     uint     `gorm:"column:order_id;not null;index" json:"order_id"`
    ItemID      uint     `gorm:"column:item_id;not null" json:"item_id"`
    Quantity    float64  `gorm:"column:quantity;not null" json:"quantity"`        // DECIMAL(10,2) for flexibility
    UnitPrice   float64  `gorm:"column:unit_price;not null" json:"unit_price"`
    ItemTotal   float64  `gorm:"column:item_total;not null" json:"item_total"`
    // Associations
    Order       Order    `gorm:"foreignKey:OrderID;references:OrderID" json:"order"`
    Item        Item     `gorm:"foreignKey:ItemID;references:ItemID" json:"item"`
}


// Invoice represents the invoices table.
type Invoice struct {
    InvoiceID          uint         `gorm:"type:int unsigned;primaryKey;autoIncrement;column:invoice_id" json:"invoice_id"`
    SenderCompanyID    uint         `gorm:"type:int unsigned;column:sender_company_id;not null" json:"sender_company_id"`
    RecipientCompanyID uint         `gorm:"type:int unsigned;column:recipient_company_id;not null;index" json:"recipient_company_id"`
    BillingAddressID   uint         `gorm:"type:int unsigned;column:billing_address_id;not null" json:"billing_address_id"`
    ShippingAddressID  *uint        `gorm:"type:int unsigned;column:shipping_address_id" json:"shipping_address_id,omitempty"`
    OrderID            *uint        `gorm:"type:int unsigned;column:order_id" json:"order_id,omitempty"`
    InvoiceNumber      string       `gorm:"column:invoice_number;not null;unique" json:"invoice_number"`
    InvoiceDate        time.Time    `gorm:"column:invoice_date;not null" json:"invoice_date"`
    DueDate            time.Time    `gorm:"column:due_date;not null" json:"due_date"`
    InvoiceSubject     *string      `gorm:"column:invoice_subject" json:"invoice_subject,omitempty"`
    Subtotal           float64      `gorm:"column:subtotal;not null;default:0.00" json:"subtotal"`
    TaxTotal           float64      `gorm:"column:tax_total;not null;default:0.00" json:"tax_total"`
    GrandTotal         float64      `gorm:"column:grand_total;not null;default:0.00" json:"grand_total"`
    AmountPaid         float64      `gorm:"column:amount_paid;not null;default:0.00" json:"amount_paid"`
    AmountDue          float64      `gorm:"column:amount_due;not null;default:0.00" json:"amount_due"`
    Status             string       `gorm:"column:status;not null;default:'Draft';index" json:"status"`
    Notes              *string      `gorm:"column:notes" json:"notes,omitempty"`
    CreatedAt          time.Time    `gorm:"column:created_at;autoCreateTime" json:"created_at"`
    UpdatedAt          time.Time    `gorm:"column:updated_at;autoUpdateTime" json:"updated_at"`
    // Associations
    SenderCompany      Company      `gorm:"foreignKey:SenderCompanyID;references:CompanyID;constraint:OnDelete:RESTRICT" json:"sender_company"`
    RecipientCompany   Company      `gorm:"foreignKey:RecipientCompanyID;references:CompanyID;constraint:OnDelete:RESTRICT" json:"recipient_company"`
    BillingAddress     Address      `gorm:"foreignKey:BillingAddressID;references:AddressID;constraint:OnDelete:RESTRICT" json:"billing_address"`
    ShippingAddress    *Address     `gorm:"foreignKey:ShippingAddressID;references:AddressID;constraint:OnDelete:RESTRICT" json:"shipping_address,omitempty"`
    Order              *Order       `gorm:"foreignKey:OrderID;references:OrderID;constraint:OnDelete:RESTRICT" json:"order,omitempty"`
    InvoiceItems       []InvoiceItem `gorm:"foreignKey:InvoiceID;references:InvoiceID;constraint:OnDelete:CASCADE" json:"invoice_items"`
}

// InvoiceItem represents the invoice_items table.
type InvoiceItem struct {
    InvoiceItemID     uint     `gorm:"primaryKey;autoIncrement;column:invoice_item_id" json:"invoice_item_id"`
    InvoiceID         uint     `gorm:"column:invoice_id;not null;index" json:"invoice_id"`
    ItemID            *uint    `gorm:"column:item_id" json:"item_id,omitempty"` // Nullable: custom line items allowed
    Description       string   `gorm:"column:description;not null" json:"description"`
    Quantity          float64  `gorm:"column:quantity;not null" json:"quantity"`
    UnitPrice         float64  `gorm:"column:unit_price;not null" json:"unit_price"`
    ItemTotal         float64  `gorm:"column:item_total;not null" json:"item_total"`
    TaxRatePercentage float64  `gorm:"column:tax_rate_percentage;default:0.00" json:"tax_rate_percentage"`
    // Associations
    Invoice           Invoice  `gorm:"foreignKey:InvoiceID;references:InvoiceID" json:"invoice"`
    Item              *Item    `gorm:"foreignKey:ItemID;references:ItemID" json:"item,omitempty"`
}

// Payment represents the payments table.
type Payment struct {
    PaymentID               uint       `gorm:"primaryKey;autoIncrement;column:payment_id" json:"payment_id"`
    InvoiceID               uint       `gorm:"column:invoice_id;not null;index" json:"invoice_id"`
    PaymentDate             time.Time  `gorm:"column:payment_date;not null" json:"payment_date"`
    Amount                  float64    `gorm:"column:amount;not null" json:"amount"`
    Method                  *string    `gorm:"column:method" json:"method,omitempty"` // e.g., Credit Card, Bank Transfer, etc.
    TransactionReference    *string    `gorm:"column:transaction_references";"type:varchar(255)" json:"transaction_references"`
    Status                  string     `gorm:"column:status;not null;default:'Completed'" json:"status"`
    CreatedAt               time.Time  `gorm:"column:created_at;autoCreateTime" json:"created_at"`
    // Associations
    Invoice                 Invoice    `gorm:"foreignKey:InvoiceID;references:InvoiceID" json:"invoice"`
}

/**
Refactor request
*/
// CompanyRequest is used for creating/updating a company
type CompanyRequest struct {
	CompanyName             string `json:"company_name" binding:"required,min=1,max=100"`
	ContactPerson           string `json:"contact_person" binding:"max=100"`
	Email                   string `json:"email" binding:"required,email,max=100"`
	Phone                   string `json:"phone" binding:"max=20"`
	IsCustomer              bool   `json:"is_customer"`
	IsVendor                bool   `json:"is_vendor"`
	DefaultBillingAddressID *uint  `json:"default_billing_address_id"`
	DefaultShippingAddressID *uint `json:"default_shipping_address_id"`
}

// AddressRequest is used for creating/updating an address
type AddressRequest struct {
	CompanyID     uint   `json:"company_id" binding:"required"`
	AddressType   string `json:"address_type" binding:"required,min=1,max=20"`
	Street        string `json:"street" binding:"required,min=1,max=200"`
	City          string `json:"city" binding:"required,min=1,max=100"`
	StateProvince string `json:"state_province" binding:"required,min=1,max=100"`
	PostalCode    string `json:"postal_code" binding:"required,min=1,max=20"`
	Country       string `json:"country" binding:"required,min=1,max=100"`
}

// ItemRequest is used for creating/updating an item
type ItemRequest struct {
	ItemName        string  `json:"item_name" binding:"required,min=1,max=100"`
	ItemDescription string  `json:"item_description" binding:"max=500"`
	UnitPrice       float64 `json:"unit_price" binding:"required,min=0"`
	ItemType        string  `json:"item_type" binding:"required,min=1,max=50"`
	ImagePath       string  `json:"image_path" binding:"max=255"`
}

// OrderRequest is used for creating an order
type OrderRequest struct {
	CustomerCompanyID uint                 `json:"customer_company_id" binding:"required"`
	OrderDate         time.Time            `json:"order_date" binding:"required"`
	Status            string               `json:"status" binding:"required,min=1,max=20"`
	OrderItems        []OrderItemRequest   `json:"order_items" binding:"required,dive"`
}

// OrderItemRequest is used for creating order items
type OrderItemRequest struct {
	ProductID uint32    `json:"product_id" binding:"required"`
	Quantity  int     `json:"quantity" binding:"required,min=1"`
}

// Invoice Details
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

// InvoiceRequest is used for creating an invoice
type InvoiceRequest struct {
	SenderCompanyID    uint      `json:"sender_company_id" binding:"required"`
	RecipientCompanyID uint      `json:"recipient_company_id" binding:"required"`
	BillingAddressID   uint      `json:"billing_address_id" binding:"required"`
	ShippingAddressID  uint      `json:"shipping_address_id" binding:"required"`
	OrderID            uint      `json:"order_id" binding:"required"`
	InvoiceNumber      string    `json:"invoice_number" binding:"required,min=1,max=50"`
	InvoiceDate        time.Time `json:"invoice_date" binding:"required"`
	DueDate            time.Time `json:"due_date" binding:"required"`
	InvoiceSubject     string    `json:"invoice_subject" binding:"max=200"`
	Notes              string    `json:"notes" binding:"max=500"`
}

// PaymentRequest is used for creating a payment
type PaymentRequest struct {
	InvoiceID           uint      `json:"invoice_id" binding:"required"`
	PaymentDate         time.Time `json:"payment_date" binding:"required"`
	Amount              float64   `json:"amount" binding:"required,min=0.01"`
	Method              string    `json:"method" binding:"required,min=1,max=50"`
	TransactionReference string    `json:"transaction_reference" binding:"max=100"`
}

// InvoiceReportResponse provides detailed invoice information
type InvoiceReportResponse struct {
	Invoice          Invoice       `json:"invoice"`
	SenderCompany    Company       `json:"sender_company"`
	RecipientCompany Company       `json:"recipient_company"`
	BillingAddress   *Address       `json:"billing_address"`
	ShippingAddress  *Address       `json:"shipping_address"`
	Order            Order         `json:"order"`
	Items            []InvoiceItem `json:"items"`
	Payments         []Payment     `json:"payments"`
	PaymentStatus    string        `json:"payment_status"`
}
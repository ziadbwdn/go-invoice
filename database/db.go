// database
package database

import (
	"fmt"
	"log"
	"os"
	"time"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	// "invoice-go/models"
)

var DB *gorm.DB

// InitDB establishes a connection to the database and configures the schema
func InitDB() *gorm.DB {
	// Get database connection parameters from environment variables
	// or use defaults for local development
	dbUser := getEnv("DB_USER", "root")
	dbHost := getEnv("DB_HOST", "localhost")
	dbPort := getEnv("DB_PORT", "3306")
	dbName := getEnv("DB_NAME", "invoice-go")

	var dbPassword string
	fmt.Print("Enter database password: ") // Prompt the user
	_, errScan := fmt.Scan(&dbPassword)     // Read input into dbPassword
	if errScan != nil {
		log.Fatalf("Failed to read password: %v", errScan)
	}

	// Construct the connection DSN
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		dbUser, dbPassword, dbHost, dbPort, dbName)

	// Configure GORM logger to display SQL queries during development
	newLogger := logger.New(
		log.New(os.Stdout, "\r\n", log.LstdFlags),
		logger.Config{
			SlowThreshold: time.Second,
			LogLevel:      logger.Info,
			Colorful:      true,
		},
	)

	// Connect to the database
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{
		Logger: newLogger,
	})
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	// Enable connection pooling
	sqlDB, err := db.DB()
	if err != nil {
		log.Fatalf("Failed to get database connection: %v", err)
	}

	// Set connection pool settings
	sqlDB.SetMaxIdleConns(10)
	sqlDB.SetMaxOpenConns(100)
	sqlDB.SetConnMaxLifetime(time.Hour)

	// Set up the database schema using the mixed migration approach
	if err := setupDatabaseSchema(db); err != nil {
		log.Fatalf("Failed to set up database schema: %v", err)
	}

	DB = db
	return db
}

// setupDatabaseSchema handles the database migration process with proper foreign key handling
func setupDatabaseSchema(db *gorm.DB) error {
	log.Println("Starting database schema setup...")
	
	// STEP 1: Disable foreign key checks during the entire migration process
	if err := db.Exec("SET FOREIGN_KEY_CHECKS = 0").Error; err != nil {
		return fmt.Errorf("failed to disable foreign key checks: %w", err)
	}
	log.Println("Foreign key checks disabled")
	
	// Ensure we re-enable foreign key checks at the end, even if there's an error
	defer func() {
		db.Exec("SET FOREIGN_KEY_CHECKS = 1")
		log.Println("Foreign key checks re-enabled")
	}()
	
	// STEP 2: Drop all existing tables to start with a clean slate
	log.Println("Dropping existing tables...")
	
	tablesToDrop := []string{
		"payments", "invoice_items", "invoices", "order_items", "orders", 
		"addresses", "items", "companies",
	}
	
	for _, table := range tablesToDrop {
		if err := db.Exec(fmt.Sprintf("DROP TABLE IF EXISTS %s", table)).Error; err != nil {
			log.Printf("Warning: Failed to drop table %s: %v", table, err)
		}
	}
	
	// STEP 3: Create all tables WITHOUT any foreign keys or constraints
	log.Println("Creating tables without constraints...")
	
	// Companies table - aligned with Company struct
	if err := db.Exec(`
		CREATE TABLE companies (
			company_id INT UNSIGNED NOT NULL AUTO_INCREMENT,
			company_name VARCHAR(255) NOT NULL,
			contact_person VARCHAR(255),
			email VARCHAR(255),
			phone VARCHAR(50),
			is_customer BOOLEAN NOT NULL DEFAULT FALSE,
			is_vendor BOOLEAN NOT NULL DEFAULT FALSE,
			default_billing_address_id INT UNSIGNED,
			default_shipping_address_id INT UNSIGNED,
			created_at TIMESTAMP NULL,
			updated_at TIMESTAMP NULL,
			PRIMARY KEY (company_id),
			INDEX idx_companies_company_name (company_name),
			INDEX idx_companies_is_customer (is_customer),
			INDEX idx_companies_is_vendor (is_vendor),
			INDEX idx_companies_default_billing_address (default_billing_address_id),
			INDEX idx_companies_default_shipping_address (default_shipping_address_id)
		)
	`).Error; err != nil {
		return fmt.Errorf("failed to create companies table: %w", err)
	}
	
	// Addresses table - aligned with Address struct
	if err := db.Exec(`
		CREATE TABLE addresses (
			address_id INT UNSIGNED NOT NULL AUTO_INCREMENT,
			company_id INT UNSIGNED NOT NULL,
			address_type VARCHAR(50) NOT NULL,
			street VARCHAR(255) NOT NULL,
			city VARCHAR(200) NOT NULL,
			state_province VARCHAR(100),
			postal_code VARCHAR(20) NOT NULL,
			country VARCHAR(200) NOT NULL,
			created_at TIMESTAMP NULL,
			updated_at TIMESTAMP NULL,
			PRIMARY KEY (address_id),
			INDEX idx_addresses_company (company_id)
		)
	`).Error; err != nil {
		return fmt.Errorf("failed to create addresses table: %w", err)
	}
	
	// Items table - aligned with Item struct
	if err := db.Exec(`
		CREATE TABLE items (
			item_id INT UNSIGNED NOT NULL AUTO_INCREMENT,
			name VARCHAR(100) NOT NULL,
			description TEXT,
			unit_price DECIMAL(10,2) NOT NULL,
			type VARCHAR(50) NOT NULL,
			stock INT NOT NULL DEFAULT 0,
			image_path VARCHAR(255),
			created_at TIMESTAMP NULL,
			updated_at TIMESTAMP NULL,
			PRIMARY KEY (item_id)
		)
	`).Error; err != nil {
		return fmt.Errorf("failed to create items table: %w", err)
	}
	
	// Orders table - aligned with Order struct
	if err := db.Exec(`
		CREATE TABLE orders (
			order_id INT UNSIGNED NOT NULL AUTO_INCREMENT,
			customer_company_id INT UNSIGNED NOT NULL,
			order_date TIMESTAMP NOT NULL,
			total_price DECIMAL(10,2),
			status VARCHAR(50) NOT NULL DEFAULT 'Pending',
			created_at TIMESTAMP NULL,
			updated_at TIMESTAMP NULL,
			PRIMARY KEY (order_id),
			INDEX idx_orders_customer (customer_company_id),
			INDEX idx_orders_status (status)
		)
	`).Error; err != nil {
		return fmt.Errorf("failed to create orders table: %w", err)
	}
	
	// OrderItems table - aligned with OrderItem struct
	if err := db.Exec(`
		CREATE TABLE order_items (
			order_item_id INT UNSIGNED NOT NULL AUTO_INCREMENT,
			order_id INT UNSIGNED NOT NULL,
			item_id INT UNSIGNED NOT NULL,
			quantity DECIMAL(10,2) NOT NULL,
			unit_price DECIMAL(10,2) NOT NULL,
			item_total DECIMAL(10,2) NOT NULL,
			PRIMARY KEY (order_item_id),
			INDEX idx_order_items_order (order_id),
			INDEX idx_order_items_item (item_id)
		)
	`).Error; err != nil {
		return fmt.Errorf("failed to create order_items table: %w", err)
	}
	
	// Invoices table - aligned with Invoice struct
	if err := db.Exec(`
		CREATE TABLE invoices (
			invoice_id INT UNSIGNED NOT NULL AUTO_INCREMENT,
			sender_company_id INT UNSIGNED NOT NULL,
			recipient_company_id INT UNSIGNED NOT NULL,
			billing_address_id INT UNSIGNED NOT NULL,
			shipping_address_id INT UNSIGNED,
			order_id INT UNSIGNED,
			invoice_number VARCHAR(50) NOT NULL,
			invoice_date TIMESTAMP NOT NULL,
			due_date TIMESTAMP NOT NULL,
			invoice_subject VARCHAR(255),
			subtotal DECIMAL(10,2) NOT NULL DEFAULT 0.00,
			tax_total DECIMAL(10,2) NOT NULL DEFAULT 0.00,
			grand_total DECIMAL(10,2) NOT NULL DEFAULT 0.00,
			amount_paid DECIMAL(10,2) NOT NULL DEFAULT 0.00,
			amount_due DECIMAL(10,2) NOT NULL DEFAULT 0.00,
			status VARCHAR(50) NOT NULL DEFAULT 'Draft',
			notes TEXT,
			created_at TIMESTAMP NULL,
			updated_at TIMESTAMP NULL,
			PRIMARY KEY (invoice_id),
			UNIQUE KEY unique_invoice_number (invoice_number),
			INDEX idx_invoices_sender (sender_company_id),
			INDEX idx_invoices_recipient (recipient_company_id),
			INDEX idx_invoices_billing (billing_address_id),
			INDEX idx_invoices_shipping (shipping_address_id),
			INDEX idx_invoices_order (order_id),
			INDEX idx_invoices_status (status)
		)
	`).Error; err != nil {
		return fmt.Errorf("failed to create invoices table: %w", err)
	}
	
	// InvoiceItems table - aligned with InvoiceItem struct
	if err := db.Exec(`
		CREATE TABLE invoice_items (
			invoice_item_id INT UNSIGNED NOT NULL AUTO_INCREMENT,
			invoice_id INT UNSIGNED NOT NULL,
			item_id INT UNSIGNED,
			description VARCHAR(255) NOT NULL,
			quantity DECIMAL(10,2) NOT NULL,
			unit_price DECIMAL(10,2) NOT NULL,
			item_total DECIMAL(10,2) NOT NULL,
			tax_rate_percentage DECIMAL(5,2) DEFAULT 0.00,
			PRIMARY KEY (invoice_item_id),
			INDEX idx_invoice_items_invoice (invoice_id),
			INDEX idx_invoice_items_item (item_id)
		)
	`).Error; err != nil {
		return fmt.Errorf("failed to create invoice_items table: %w", err)
	}
	
	// Payments table - aligned with Payment struct
	if err := db.Exec(`
		CREATE TABLE payments (
			payment_id INT UNSIGNED NOT NULL AUTO_INCREMENT,
			invoice_id INT UNSIGNED NOT NULL,
			payment_date TIMESTAMP NOT NULL,
			amount DECIMAL(10,2) NOT NULL,
			method VARCHAR(50),
			transaction_reference VARCHAR(255),
			status VARCHAR(50) NOT NULL DEFAULT 'Completed',
			created_at TIMESTAMP NULL,
			PRIMARY KEY (payment_id),
			INDEX idx_payments_invoice (invoice_id)
		)
	`).Error; err != nil {
		return fmt.Errorf("failed to create payments table: %w", err)
	}
	
	// STEP 4: Add all foreign key constraints
	log.Println("Adding foreign key constraints...")
	
	fkConstraints := []string{
		// Companies and Addresses circular references
		"ALTER TABLE companies ADD CONSTRAINT fk_company_billing_addr FOREIGN KEY (default_billing_address_id) REFERENCES addresses(address_id) ON DELETE SET NULL",
		"ALTER TABLE companies ADD CONSTRAINT fk_company_shipping_addr FOREIGN KEY (default_shipping_address_id) REFERENCES addresses(address_id) ON DELETE SET NULL",
		
		// Addresses → Companies
		"ALTER TABLE addresses ADD CONSTRAINT fk_address_company FOREIGN KEY (company_id) REFERENCES companies(company_id) ON DELETE CASCADE",
		
		// Orders → Companies
		"ALTER TABLE orders ADD CONSTRAINT fk_order_customer FOREIGN KEY (customer_company_id) REFERENCES companies(company_id) ON DELETE RESTRICT",
		
		// OrderItems → Orders
		"ALTER TABLE order_items ADD CONSTRAINT fk_orderitem_order FOREIGN KEY (order_id) REFERENCES orders(order_id) ON DELETE CASCADE",
		
		// OrderItems → Items
		"ALTER TABLE order_items ADD CONSTRAINT fk_orderitem_item FOREIGN KEY (item_id) REFERENCES items(item_id) ON DELETE RESTRICT",
		
		// Invoices → Companies (sender and recipient)
		"ALTER TABLE invoices ADD CONSTRAINT fk_invoice_sender FOREIGN KEY (sender_company_id) REFERENCES companies(company_id) ON DELETE RESTRICT",
		"ALTER TABLE invoices ADD CONSTRAINT fk_invoice_recipient FOREIGN KEY (recipient_company_id) REFERENCES companies(company_id) ON DELETE RESTRICT",
		
		// Invoices → Addresses (billing and shipping)
		"ALTER TABLE invoices ADD CONSTRAINT fk_invoice_billing FOREIGN KEY (billing_address_id) REFERENCES addresses(address_id) ON DELETE RESTRICT",
		"ALTER TABLE invoices ADD CONSTRAINT fk_invoice_shipping FOREIGN KEY (shipping_address_id) REFERENCES addresses(address_id) ON DELETE RESTRICT",
		
		// Invoices → Orders
		"ALTER TABLE invoices ADD CONSTRAINT fk_invoice_order FOREIGN KEY (order_id) REFERENCES orders(order_id) ON DELETE RESTRICT",
		
		// InvoiceItems → Invoices
		"ALTER TABLE invoice_items ADD CONSTRAINT fk_invoiceitem_invoice FOREIGN KEY (invoice_id) REFERENCES invoices(invoice_id) ON DELETE CASCADE",
		
		// InvoiceItems → Items (optional relationship)
		"ALTER TABLE invoice_items ADD CONSTRAINT fk_invoiceitem_item FOREIGN KEY (item_id) REFERENCES items(item_id) ON DELETE RESTRICT",
		
		// Payments → Invoices
		"ALTER TABLE payments ADD CONSTRAINT fk_payment_invoice FOREIGN KEY (invoice_id) REFERENCES invoices(invoice_id) ON DELETE RESTRICT",
	}
	
	for _, constraint := range fkConstraints {
		if err := db.Exec(constraint).Error; err != nil {
			log.Printf("Warning: Failed to add constraint: %s. Error: %v", constraint, err)
			// Continue with other constraints instead of returning error
		}
	}
	
	log.Println("Database schema setup completed successfully")
	return nil
}

// GetDB returns the database connection instance
func GetDB() *gorm.DB {
	return DB
}

// Helper function to get environment variables with defaults
func getEnv(key, defaultValue string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return defaultValue
}

/**
func SetupDatabase(db *gorm.DB) error {
    // Step 1: Drop any existing foreign key constraints
    db.Exec("SET FOREIGN_KEY_CHECKS=0")
    
    // Step 2: Create or modify tables without foreign keys
    if err := db.AutoMigrate(&Company{}, &Address{}, &Item{}, &Order{}, &OrderItem{}, &Invoice{}, &InvoiceItem{}, &Payment{}); err != nil {
        return fmt.Errorf("failed to migrate tables: %w", err)
    }
    
    // Step 3: Add foreign keys manually
    queries := []string{
        // Companies → Addresses
        "ALTER TABLE companies ADD CONSTRAINT fk_companies_billing_address FOREIGN KEY (default_billing_address_id) REFERENCES addresses(address_id) ON DELETE RESTRICT",
        "ALTER TABLE companies ADD CONSTRAINT fk_companies_shipping_address FOREIGN KEY (default_shipping_address_id) REFERENCES addresses(address_id) ON DELETE RESTRICT",
        
        // Addresses → Companies
        "ALTER TABLE addresses ADD CONSTRAINT fk_addresses_company FOREIGN KEY (company_id) REFERENCES companies(company_id) ON DELETE CASCADE",
        
        // Orders → Companies
        "ALTER TABLE orders ADD CONSTRAINT fk_orders_customer FOREIGN KEY (customer_company_id) REFERENCES companies(company_id) ON DELETE RESTRICT",
        
        // OrderItems → Orders, Items
        "ALTER TABLE order_items ADD CONSTRAINT fk_order_items_order FOREIGN KEY (order_id) REFERENCES orders(order_id) ON DELETE RESTRICT",
        "ALTER TABLE order_items ADD CONSTRAINT fk_order_items_item FOREIGN KEY (item_id) REFERENCES items(item_id) ON DELETE RESTRICT",
        
        // Invoices → Companies, Addresses, Orders
        "ALTER TABLE invoices ADD CONSTRAINT fk_invoices_sender FOREIGN KEY (sender_company_id) REFERENCES companies(company_id) ON DELETE RESTRICT",
        "ALTER TABLE invoices ADD CONSTRAINT fk_invoices_recipient FOREIGN KEY (recipient_company_id) REFERENCES companies(company_id) ON DELETE RESTRICT",
        "ALTER TABLE invoices ADD CONSTRAINT fk_invoices_billing_address FOREIGN KEY (billing_address_id) REFERENCES addresses(address_id) ON DELETE RESTRICT",
        "ALTER TABLE invoices ADD CONSTRAINT fk_invoices_shipping_address FOREIGN KEY (shipping_address_id) REFERENCES addresses(address_id) ON DELETE RESTRICT",
        "ALTER TABLE invoices ADD CONSTRAINT fk_invoices_order FOREIGN KEY (order_id) REFERENCES orders(order_id) ON DELETE RESTRICT",
        
        // InvoiceItems → Invoices, Items
        "ALTER TABLE invoice_items ADD CONSTRAINT fk_invoice_items_invoice FOREIGN KEY (invoice_id) REFERENCES invoices(invoice_id) ON DELETE CASCADE",
        "ALTER TABLE invoice_items ADD CONSTRAINT fk_invoice_items_product FOREIGN KEY (product_id) REFERENCES items(item_id) ON DELETE RESTRICT",
        
        // Payments → Invoices
        "ALTER TABLE payments ADD CONSTRAINT fk_payments_invoice FOREIGN KEY (invoice_id) REFERENCES invoices(invoice_id) ON DELETE RESTRICT",
    }
    
    // Execute each ALTER TABLE statement
    for _, query := range queries {
        if err := db.Exec(query).Error; err != nil {
            // If this fails, it's ok to continue with the rest - just log it
            fmt.Printf("Warning: Failed to execute query: %s\nError: %v\n", query, err)
        }
    }
    
    // Re-enable foreign key checks
    db.Exec("SET FOREIGN_KEY_CHECKS=1")
    
    return nil
}

02
func setupDatabaseSchema(db *gorm.DB) error {
	log.Println("Starting database schema setup...")
	
	// STEP 1: Disable foreign key checks during the entire migration process
	if err := db.Exec("SET FOREIGN_KEY_CHECKS = 0").Error; err != nil {
		return fmt.Errorf("failed to disable foreign key checks: %w", err)
	}
	log.Println("Foreign key checks disabled")
	
	// Ensure we re-enable foreign key checks at the end, even if there's an error
	defer func() {
		db.Exec("SET FOREIGN_KEY_CHECKS = 1")
		log.Println("Foreign key checks re-enabled")
	}()
	
	// STEP 2: Drop all existing tables to start with a clean slate
	// This is the most reliable way to avoid constraint conflicts
	log.Println("Dropping existing tables...")
	
	tablesToDrop := []string{
		"payments", "invoice_items", "invoices", "order_items", "orders", 
		"addresses", "items", "companies",
	}
	
	for _, table := range tablesToDrop {
		if err := db.Exec(fmt.Sprintf("DROP TABLE IF EXISTS %s", table)).Error; err != nil {
			log.Printf("Warning: Failed to drop table %s: %v", table, err)
		}
	}
	
	// STEP 3: Create all tables WITHOUT any foreign keys or constraints
	log.Println("Creating tables without constraints...")
	
	// Companies table
	if err := db.Exec(`
		CREATE TABLE companies (
			company_id INT UNSIGNED NOT NULL AUTO_INCREMENT,
			name VARCHAR(255) NOT NULL,
			tax_id VARCHAR(50),
			email VARCHAR(255),
			phone VARCHAR(50),
			website VARCHAR(255),
			default_billing_address_id INT UNSIGNED NULL,
			default_shipping_address_id INT UNSIGNED NULL,
			created_at TIMESTAMP NULL,
			updated_at TIMESTAMP NULL,
			deleted_at TIMESTAMP NULL,
			PRIMARY KEY (company_id),
			INDEX idx_companies_deleted_at (deleted_at),
			INDEX idx_companies_default_billing_address (default_billing_address_id),
			INDEX idx_companies_default_shipping_address (default_shipping_address_id)
		)
	`).Error; err != nil {
		return fmt.Errorf("failed to create companies table: %w", err)
	}
	
	// Addresses table
	if err := db.Exec(`
		CREATE TABLE addresses (
			address_id INT UNSIGNED NOT NULL AUTO_INCREMENT,
			company_id INT UNSIGNED NOT NULL,
			address_type VARCHAR(50),
			street_line1 VARCHAR(255) NOT NULL,
			street_line2 VARCHAR(255),
			city VARCHAR(100) NOT NULL,
			state VARCHAR(100),
			postal_code VARCHAR(20) NOT NULL,
			country VARCHAR(100) NOT NULL,
			is_default BOOLEAN DEFAULT FALSE,
			created_at TIMESTAMP NULL,
			updated_at TIMESTAMP NULL,
			deleted_at TIMESTAMP NULL,
			PRIMARY KEY (address_id),
			INDEX idx_addresses_company (company_id),
			INDEX idx_addresses_deleted_at (deleted_at)
		)
	`).Error; err != nil {
		return fmt.Errorf("failed to create addresses table: %w", err)
	}
	
	// Items table
	if err := db.Exec(`
		CREATE TABLE items (
			item_id INT UNSIGNED NOT NULL AUTO_INCREMENT,
			name VARCHAR(255) NOT NULL,
			description TEXT,
			sku VARCHAR(100),
			unit_price DECIMAL(10,2) NOT NULL,
			currency VARCHAR(3) DEFAULT 'USD',
			tax_rate DECIMAL(5,2) DEFAULT 0.00,
			is_taxable BOOLEAN DEFAULT TRUE,
			created_at TIMESTAMP NULL,
			updated_at TIMESTAMP NULL,
			deleted_at TIMESTAMP NULL,
			PRIMARY KEY (item_id),
			INDEX idx_items_deleted_at (deleted_at)
		)
	`).Error; err != nil {
		return fmt.Errorf("failed to create items table: %w", err)
	}
	
	// Orders table
	if err := db.Exec(`
		CREATE TABLE orders (
			order_id INT UNSIGNED NOT NULL AUTO_INCREMENT,
			order_number VARCHAR(50) NOT NULL,
			customer_company_id INT UNSIGNED NOT NULL,
			order_date DATE NOT NULL,
			status VARCHAR(50) DEFAULT 'New',
			notes TEXT,
			created_at TIMESTAMP NULL,
			updated_at TIMESTAMP NULL,
			deleted_at TIMESTAMP NULL,
			PRIMARY KEY (order_id),
			INDEX idx_orders_customer (customer_company_id),
			INDEX idx_orders_deleted_at (deleted_at)
		)
	`).Error; err != nil {
		return fmt.Errorf("failed to create orders table: %w", err)
	}
	
	// OrderItems table
	if err := db.Exec(`
		CREATE TABLE order_items (
			order_item_id INT UNSIGNED NOT NULL AUTO_INCREMENT,
			order_id INT UNSIGNED NOT NULL,
			item_id INT UNSIGNED NOT NULL,
			quantity INT NOT NULL,
			unit_price DECIMAL(10,2) NOT NULL,
			tax_rate DECIMAL(5,2) DEFAULT 0.00,
			is_taxable BOOLEAN DEFAULT TRUE,
			notes TEXT,
			created_at TIMESTAMP NULL,
			updated_at TIMESTAMP NULL,
			PRIMARY KEY (order_item_id),
			INDEX idx_order_items_order (order_id),
			INDEX idx_order_items_item (item_id)
		)
	`).Error; err != nil {
		return fmt.Errorf("failed to create order_items table: %w", err)
	}
	
	// Invoices table
	if err := db.Exec(`
		CREATE TABLE invoices (
			invoice_id INT UNSIGNED NOT NULL AUTO_INCREMENT,
			invoice_number VARCHAR(50) NOT NULL,
			sender_company_id INT UNSIGNED NOT NULL,
			recipient_company_id INT UNSIGNED NOT NULL,
			billing_address_id INT UNSIGNED NOT NULL,
			shipping_address_id INT UNSIGNED NOT NULL,
			order_id INT UNSIGNED NULL,
			issue_date DATE NOT NULL,
			due_date DATE NOT NULL,
			status VARCHAR(50) DEFAULT 'Draft',
			notes TEXT,
			created_at TIMESTAMP NULL,
			updated_at TIMESTAMP NULL,
			deleted_at TIMESTAMP NULL,
			PRIMARY KEY (invoice_id),
			INDEX idx_invoices_sender (sender_company_id),
			INDEX idx_invoices_recipient (recipient_company_id),
			INDEX idx_invoices_billing (billing_address_id),
			INDEX idx_invoices_shipping (shipping_address_id),
			INDEX idx_invoices_order (order_id),
			INDEX idx_invoices_deleted_at (deleted_at)
		)
	`).Error; err != nil {
		return fmt.Errorf("failed to create invoices table: %w", err)
	}
	
	// InvoiceItems table
	if err := db.Exec(`
		CREATE TABLE invoice_items (
			invoice_item_id INT UNSIGNED NOT NULL AUTO_INCREMENT,
			invoice_id INT UNSIGNED NOT NULL,
			product_id INT UNSIGNED NOT NULL,
			description TEXT,
			quantity INT NOT NULL,
			unit_price DECIMAL(10,2) NOT NULL,
			tax_rate DECIMAL(5,2) DEFAULT 0.00,
			is_taxable BOOLEAN DEFAULT TRUE,
			created_at TIMESTAMP NULL,
			updated_at TIMESTAMP NULL,
			PRIMARY KEY (invoice_item_id),
			INDEX idx_invoice_items_invoice (invoice_id),
			INDEX idx_invoice_items_product (product_id)
		)
	`).Error; err != nil {
		return fmt.Errorf("failed to create invoice_items table: %w", err)
	}
	
	// Payments table
	if err := db.Exec(`
		CREATE TABLE payments (
			payment_id INT UNSIGNED NOT NULL AUTO_INCREMENT,
			invoice_id INT UNSIGNED NOT NULL,
			payment_date DATE NOT NULL,
			amount DECIMAL(10,2) NOT NULL,
			payment_method VARCHAR(50) NOT NULL,
			reference_number VARCHAR(100),
			notes TEXT,
			created_at TIMESTAMP NULL,
			updated_at TIMESTAMP NULL,
			deleted_at TIMESTAMP NULL,
			PRIMARY KEY (payment_id),
			INDEX idx_payments_invoice (invoice_id),
			INDEX idx_payments_deleted_at (deleted_at)
		)
	`).Error; err != nil {
		return fmt.Errorf("failed to create payments table: %w", err)
	}
	
	// STEP 4: Add all foreign key constraints
	log.Println("Adding foreign key constraints...")
	
	fkConstraints := []string{
		// First, add all constraints except the circular ones between companies and addresses
		
		// Addresses → Companies (an address belongs to a company)
		"ALTER TABLE addresses ADD CONSTRAINT fk_address_company FOREIGN KEY (company_id) REFERENCES companies(company_id) ON DELETE CASCADE",
		
		// Orders → Companies (an order is placed by a customer company)
		"ALTER TABLE orders ADD CONSTRAINT fk_order_customer FOREIGN KEY (customer_company_id) REFERENCES companies(company_id) ON DELETE RESTRICT",
		
		// OrderItems → Orders (an order item belongs to an order)
		"ALTER TABLE order_items ADD CONSTRAINT fk_orderitem_order FOREIGN KEY (order_id) REFERENCES orders(order_id) ON DELETE CASCADE",
		
		// OrderItems → Items (an order item references a product/item)
		"ALTER TABLE order_items ADD CONSTRAINT fk_orderitem_item FOREIGN KEY (item_id) REFERENCES items(item_id) ON DELETE RESTRICT",
		
		// Invoices → Companies (sender and recipient)
		"ALTER TABLE invoices ADD CONSTRAINT fk_invoice_sender FOREIGN KEY (sender_company_id) REFERENCES companies(company_id) ON DELETE RESTRICT",
		"ALTER TABLE invoices ADD CONSTRAINT fk_invoice_recipient FOREIGN KEY (recipient_company_id) REFERENCES companies(company_id) ON DELETE RESTRICT",
		
		// Invoices → Addresses (billing and shipping)
		"ALTER TABLE invoices ADD CONSTRAINT fk_invoice_billing FOREIGN KEY (billing_address_id) REFERENCES addresses(address_id) ON DELETE RESTRICT",
		"ALTER TABLE invoices ADD CONSTRAINT fk_invoice_shipping FOREIGN KEY (shipping_address_id) REFERENCES addresses(address_id) ON DELETE RESTRICT",
		
		// Invoices → Orders (an invoice is for an order)
		"ALTER TABLE invoices ADD CONSTRAINT fk_invoice_order FOREIGN KEY (order_id) REFERENCES orders(order_id) ON DELETE RESTRICT",
		
		// InvoiceItems → Invoices (an invoice item belongs to an invoice)
		"ALTER TABLE invoice_items ADD CONSTRAINT fk_invoiceitem_invoice FOREIGN KEY (invoice_id) REFERENCES invoices(invoice_id) ON DELETE CASCADE",
		
		// InvoiceItems → Items (an invoice item references a product)
		"ALTER TABLE invoice_items ADD CONSTRAINT fk_invoiceitem_product FOREIGN KEY (product_id) REFERENCES items(item_id) ON DELETE RESTRICT",
		
		// Payments → Invoices (a payment is for an invoice)
		"ALTER TABLE payments ADD CONSTRAINT fk_payment_invoice FOREIGN KEY (invoice_id) REFERENCES invoices(invoice_id) ON DELETE RESTRICT",
		
		// Finally, add the circular references
		// Companies → Addresses (a company has billing and shipping addresses)
		"ALTER TABLE companies ADD CONSTRAINT fk_company_billing_addr FOREIGN KEY (default_billing_address_id) REFERENCES addresses(address_id) ON DELETE RESTRICT",
		"ALTER TABLE companies ADD CONSTRAINT fk_company_shipping_addr FOREIGN KEY (default_shipping_address_id) REFERENCES addresses(address_id) ON DELETE RESTRICT",
	}
	
	for _, constraint := range fkConstraints {
		if err := db.Exec(constraint).Error; err != nil {
			log.Printf("Warning: Failed to add constraint: %s. Error: %v", constraint, err)
			// Continue with other constraints instead of returning error
		}
	}
	
	log.Println("Database schema setup completed successfully")
	return nil
}

// v03
func setupDatabaseSchema(db *gorm.DB) error {
	log.Println("Starting database schema setup...")
	
	// STEP 1: Disable foreign key checks during the entire migration process
	if err := db.Exec("SET FOREIGN_KEY_CHECKS = 0").Error; err != nil {
		return fmt.Errorf("failed to disable foreign key checks: %w", err)
	}
	log.Println("Foreign key checks disabled")
	
	// Ensure we re-enable foreign key checks at the end, even if there's an error
	defer func() {
		db.Exec("SET FOREIGN_KEY_CHECKS = 1")
		log.Println("Foreign key checks re-enabled")
	}()
	
	// STEP 2: Drop all existing tables to start with a clean slate
	log.Println("Dropping existing tables...")
	
	tablesToDrop := []string{
		"payments", "invoice_items", "invoices", "order_items", "orders", 
		"addresses", "items", "companies",
	}
	
	for _, table := range tablesToDrop {
		if err := db.Exec(fmt.Sprintf("DROP TABLE IF EXISTS %s", table)).Error; err != nil {
			log.Printf("Warning: Failed to drop table %s: %v", table, err)
		}
	}
	
	// STEP 3: Create all tables WITHOUT any foreign keys or constraints
	log.Println("Creating tables without constraints...")
	
	// Companies table - updated to match model structure
	if err := db.Exec(`
		CREATE TABLE companies (
			company_id INT UNSIGNED NOT NULL AUTO_INCREMENT,
			company_name VARCHAR(255) NOT NULL,
			contact_person VARCHAR(255),
			email VARCHAR(255),
			phone VARCHAR(50),
			is_customer BOOLEAN NOT NULL DEFAULT FALSE,
			is_vendor BOOLEAN NOT NULL DEFAULT FALSE,
			default_billing_address_id INT UNSIGNED NULL,
			default_shipping_address_id INT UNSIGNED NULL,
			created_at TIMESTAMP NULL,
			updated_at TIMESTAMP NULL,
			PRIMARY KEY (company_id),
			INDEX idx_companies_company_name (company_name),
			INDEX idx_companies_is_customer (is_customer),
			INDEX idx_companies_is_vendor (is_vendor),
			UNIQUE INDEX idx_companies_email (email)
		)
	`).Error; err != nil {
		return fmt.Errorf("failed to create companies table: %w", err)
	}
	
	// Addresses table - updated to match model structure
	if err := db.Exec(`
		CREATE TABLE addresses (
			address_id INT UNSIGNED NOT NULL AUTO_INCREMENT,
			company_id INT UNSIGNED NOT NULL,
			address_type VARCHAR(50) NOT NULL,
			street VARCHAR(255) NOT NULL,
			city VARCHAR(200) NOT NULL,
			state_province VARCHAR(200),
			postal_code VARCHAR(20) NOT NULL,
			country VARCHAR(200) NOT NULL,
			created_at TIMESTAMP NULL,
			updated_at TIMESTAMP NULL,
			PRIMARY KEY (address_id),
			INDEX idx_addresses_company (company_id)
		)
	`).Error; err != nil {
		return fmt.Errorf("failed to create addresses table: %w", err)
	}
	
	// Items table - updated to match model structure
	if err := db.Exec(`
		CREATE TABLE items (
			item_id INT UNSIGNED NOT NULL AUTO_INCREMENT,
			name VARCHAR(100) NOT NULL,
			description TEXT,
			unit_price DECIMAL(10,2) NOT NULL,
			type VARCHAR(50) NOT NULL,
			stock INT NOT NULL DEFAULT 0,
			image_path VARCHAR(255),
			created_at TIMESTAMP NULL,
			updated_at TIMESTAMP NULL,
			PRIMARY KEY (item_id)
		)
	`).Error; err != nil {
		return fmt.Errorf("failed to create items table: %w", err)
	}
	
	// Orders table - updated to match model structure
	if err := db.Exec(`
		CREATE TABLE orders (
			order_id INT UNSIGNED NOT NULL AUTO_INCREMENT,
			customer_company_id INT UNSIGNED NOT NULL,
			order_date DATETIME NOT NULL,
			total_price DECIMAL(10,2),
			status VARCHAR(50) NOT NULL DEFAULT 'Pending',
			created_at TIMESTAMP NULL,
			updated_at TIMESTAMP NULL,
			PRIMARY KEY (order_id),
			INDEX idx_orders_customer (customer_company_id),
			INDEX idx_orders_status (status)
		)
	`).Error; err != nil {
		return fmt.Errorf("failed to create orders table: %w", err)
	}
	
	// OrderItems table - updated to match model structure
	if err := db.Exec(`
		CREATE TABLE order_items (
			order_item_id INT UNSIGNED NOT NULL AUTO_INCREMENT,
			order_id INT UNSIGNED NOT NULL,
			item_id INT UNSIGNED NOT NULL,
			quantity DECIMAL(10,2) NOT NULL,
			unit_price DECIMAL(10,2) NOT NULL,
			item_total DECIMAL(10,2) NOT NULL,
			PRIMARY KEY (order_item_id),
			INDEX idx_order_items_order (order_id),
			INDEX idx_order_items_item (item_id)
		)
	`).Error; err != nil {
		return fmt.Errorf("failed to create order_items table: %w", err)
	}
	
	// Invoices table - updated to match model structure
	if err := db.Exec(`
		CREATE TABLE invoices (
			invoice_id INT UNSIGNED NOT NULL AUTO_INCREMENT,
			sender_company_id INT UNSIGNED NOT NULL,
			recipient_company_id INT UNSIGNED NOT NULL,
			billing_address_id INT UNSIGNED NOT NULL,
			shipping_address_id INT UNSIGNED,
			order_id INT UNSIGNED,
			invoice_number VARCHAR(50) NOT NULL,
			invoice_date DATETIME NOT NULL,
			due_date DATETIME NOT NULL,
			invoice_subject VARCHAR(255),
			subtotal DECIMAL(10,2) NOT NULL DEFAULT 0.00,
			tax_total DECIMAL(10,2) NOT NULL DEFAULT 0.00,
			grand_total DECIMAL(10,2) NOT NULL DEFAULT 0.00,
			amount_paid DECIMAL(10,2) NOT NULL DEFAULT 0.00,
			amount_due DECIMAL(10,2) NOT NULL DEFAULT 0.00,
			status VARCHAR(50) NOT NULL DEFAULT 'Draft',
			notes TEXT,
			created_at TIMESTAMP NULL,
			updated_at TIMESTAMP NULL,
			PRIMARY KEY (invoice_id),
			UNIQUE INDEX idx_invoices_number (invoice_number),
			INDEX idx_invoices_sender (sender_company_id),
			INDEX idx_invoices_recipient (recipient_company_id),
			INDEX idx_invoices_billing (billing_address_id),
			INDEX idx_invoices_shipping (shipping_address_id),
			INDEX idx_invoices_order (order_id),
			INDEX idx_invoices_status (status)
		)
	`).Error; err != nil {
		return fmt.Errorf("failed to create invoices table: %w", err)
	}
	
	// InvoiceItems table - updated to match model structure
	if err := db.Exec(`
		CREATE TABLE invoice_items (
			invoice_item_id INT UNSIGNED NOT NULL AUTO_INCREMENT,
			invoice_id INT UNSIGNED NOT NULL,
			item_id INT UNSIGNED,
			description VARCHAR(255) NOT NULL,
			quantity DECIMAL(10,2) NOT NULL,
			unit_price DECIMAL(10,2) NOT NULL,
			item_total DECIMAL(10,2) NOT NULL,
			tax_rate_percentage DECIMAL(5,2) DEFAULT 0.00,
			PRIMARY KEY (invoice_item_id),
			INDEX idx_invoice_items_invoice (invoice_id),
			INDEX idx_invoice_items_item (item_id)
		)
	`).Error; err != nil {
		return fmt.Errorf("failed to create invoice_items table: %w", err)
	}
	
	// Payments table - updated to match model structure
	if err := db.Exec(`
		CREATE TABLE payments (
			payment_id INT UNSIGNED NOT NULL AUTO_INCREMENT,
			invoice_id INT UNSIGNED NOT NULL,
			payment_date DATETIME NOT NULL,
			amount DECIMAL(10,2) NOT NULL,
			method VARCHAR(50),
			transaction_reference VARCHAR(255),
			status VARCHAR(50) NOT NULL DEFAULT 'Completed',
			created_at TIMESTAMP NULL,
			PRIMARY KEY (payment_id),
			INDEX idx_payments_invoice (invoice_id)
		)
	`).Error; err != nil {
		return fmt.Errorf("failed to create payments table: %w", err)
	}
	
	// STEP 4: Add all foreign key constraints
	log.Println("Adding foreign key constraints...")
	
	fkConstraints := []string{
		// Companies ↔ Addresses (circular relationship)
		"ALTER TABLE companies ADD CONSTRAINT fk_company_billing_addr FOREIGN KEY (default_billing_address_id) REFERENCES addresses(address_id) ON DELETE RESTRICT",
		"ALTER TABLE companies ADD CONSTRAINT fk_company_shipping_addr FOREIGN KEY (default_shipping_address_id) REFERENCES addresses(address_id) ON DELETE RESTRICT",
		
		// Addresses → Companies
		"ALTER TABLE addresses ADD CONSTRAINT fk_address_company FOREIGN KEY (company_id) REFERENCES companies(company_id) ON DELETE CASCADE",
		
		// Orders → Companies
		"ALTER TABLE orders ADD CONSTRAINT fk_order_customer FOREIGN KEY (customer_company_id) REFERENCES companies(company_id) ON DELETE RESTRICT",
		
		// OrderItems → Orders
		"ALTER TABLE order_items ADD CONSTRAINT fk_orderitem_order FOREIGN KEY (order_id) REFERENCES orders(order_id) ON DELETE CASCADE",
		
		// OrderItems → Items
		"ALTER TABLE order_items ADD CONSTRAINT fk_orderitem_item FOREIGN KEY (item_id) REFERENCES items(item_id) ON DELETE RESTRICT",
		
		// Invoices → Companies (sender and recipient)
		"ALTER TABLE invoices ADD CONSTRAINT fk_invoice_sender FOREIGN KEY (sender_company_id) REFERENCES companies(company_id) ON DELETE RESTRICT",
		"ALTER TABLE invoices ADD CONSTRAINT fk_invoice_recipient FOREIGN KEY (recipient_company_id) REFERENCES companies(company_id) ON DELETE RESTRICT",
		
		// Invoices → Addresses (billing and shipping)
		"ALTER TABLE invoices ADD CONSTRAINT fk_invoice_billing FOREIGN KEY (billing_address_id) REFERENCES addresses(address_id) ON DELETE RESTRICT",
		"ALTER TABLE invoices ADD CONSTRAINT fk_invoice_shipping FOREIGN KEY (shipping_address_id) REFERENCES addresses(address_id) ON DELETE RESTRICT",
		
		// Invoices → Orders
		"ALTER TABLE invoices ADD CONSTRAINT fk_invoice_order FOREIGN KEY (order_id) REFERENCES orders(order_id) ON DELETE RESTRICT",
		
		// InvoiceItems → Invoices
		"ALTER TABLE invoice_items ADD CONSTRAINT fk_invoiceitem_invoice FOREIGN KEY (invoice_id) REFERENCES invoices(invoice_id) ON DELETE CASCADE",
		
		// InvoiceItems → Items
		"ALTER TABLE invoice_items ADD CONSTRAINT fk_invoiceitem_item FOREIGN KEY (item_id) REFERENCES items(item_id) ON DELETE RESTRICT",
		
		// Payments → Invoices
		"ALTER TABLE payments ADD CONSTRAINT fk_payment_invoice FOREIGN KEY (invoice_id) REFERENCES invoices(invoice_id) ON DELETE RESTRICT",
	}
	
	for _, constraint := range fkConstraints {
		if err := db.Exec(constraint).Error; err != nil {
			log.Printf("Warning: Failed to add constraint: %s. Error: %v", constraint, err)
			// Continue with other constraints instead of returning error
		}
	}
	
	log.Println("Database schema setup completed successfully")
	return nil
}
*/

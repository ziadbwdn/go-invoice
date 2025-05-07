package database

import (
	"fmt"
	"log"
	"gorm.io/gorm"
	// "invoice-go/models"
)

// seed database function
func SeedDatabase(db *gorm.DB) error {
	log.Println("Starting database seeding...")

	// Check if tables exist
	// Get current database name
	var dbName string
	if err := db.Raw("SELECT DATABASE()").Scan(&dbName).Error; err != nil {
		return fmt.Errorf("error getting database name: %w", err)
	}
	log.Printf("Checking tables in database: %s", dbName) // Debug log

    tables := []string{"companies", "addresses", "items", "orders", "order_items", "invoices", "invoice_items", "payments"}
    for _, table := range tables {
        var count int64
        if err := db.Raw("SELECT COUNT(*) FROM information_schema.tables WHERE table_schema = ? AND table_name = ?", "invoice-go", table).Count(&count).Error; err != nil {
            return fmt.Errorf("error checking table existence: %w", err)
        }
        if count == 0 {
            return fmt.Errorf("table '%s' does not exist - please run migrations first", table)
        }
    }

	// We'll execute seeding in transaction to ensure consistency
	return db.Transaction(func(tx *gorm.DB) error {
		// Seed companies
		if err := seedCompanies(tx); err != nil {
			return fmt.Errorf("failed to seed companies: %w", err)
		}

		// Seed addresses
		if err := seedAddresses(tx); err != nil {
			return fmt.Errorf("failed to seed addresses: %w", err)
		}

		// Update companies with default addresses
		// if err := updateCompaniesWithDefaultAddresses(tx); err != nil {
			// return fmt.Errorf("failed to update companies with default addresses: %w", err)
		// }

		// Seed items
		if err := seedItems(tx); err != nil {
			return fmt.Errorf("failed to seed items: %w", err)
		}

		// Seed orders
		if err := seedOrders(tx); err != nil {
			return fmt.Errorf("failed to seed orders: %w", err)
		}

		// Seed order items
		if err := seedOrderItems(tx); err != nil {
			return fmt.Errorf("failed to seed order items: %w", err)
		}

		// Seed invoices
		if err := seedInvoices(tx); err != nil {
			return fmt.Errorf("failed to seed invoices: %w", err)
		}

		// Seed invoice items
		if err := seedInvoiceItems(tx); err != nil {
			return fmt.Errorf("failed to seed invoice items: %w", err)
		}

		// Seed payments
		if err := seedPayments(tx); err != nil {
			return fmt.Errorf("failed to seed payments: %w", err)
		}

		log.Println("Database seeding completed successfully")
		return nil
	})
}

// seedCompanies inserts company records
func seedCompanies(db *gorm.DB) error {
	log.Println("Seeding companies...")
	
	query := `
	INSERT INTO companies (company_name, contact_person, email, phone, is_customer, is_vendor)
	VALUES
	-- Your company
	('InvoiceGo Corp', 'John Smith', 'admin@invoicego.com', '555-1000', FALSE, TRUE),
	
	-- Customer companies
	('Alpha Technologies', 'Emma Johnson', 'emma@alphatech.com', '555-1001', TRUE, FALSE),
	('Beta Solutions', 'Michael Chen', 'michael@betasolutions.com', '555-1002', TRUE, FALSE),
	('Gamma Industries', 'Sophia Garcia', 'sophia@gammaindustries.com', '555-1003', TRUE, FALSE),
	('Delta Innovations', 'James Williams', 'james@deltainno.com', '555-1004', TRUE, FALSE),
	('Epsilon Software', 'Olivia Brown', 'olivia@epsilonsoftware.com', '555-1005', TRUE, FALSE),
	('Zeta Consulting', 'William Jones', 'william@zetaconsulting.com', '555-1006', TRUE, FALSE),
	('Eta Manufacturing', 'Ava Miller', 'ava@etamanufacturing.com', '555-1007', TRUE, FALSE),
	('Theta Logistics', 'Alexander Davis', 'alex@thetalogistics.com', '555-1008', TRUE, FALSE),
	('Iota Services', 'Isabella Wilson', 'isabella@iotaservices.com', '555-1009', TRUE, FALSE),
	('Kappa Retail', 'Ethan Moore', 'ethan@kapparetail.com', '555-1010', TRUE, FALSE),
	
	-- Vendor companies
	('Lambda Suppliers', 'Madison Taylor', 'madison@lambdasuppliers.com', '555-1011', FALSE, TRUE),
	('Mu Electronics', 'Jacob Anderson', 'jacob@muelectronics.com', '555-1012', FALSE, TRUE),
	('Nu Packaging', 'Emily Thomas', 'emily@nupackaging.com', '555-1013', FALSE, TRUE),
	('Xi Transportation', 'Noah Jackson', 'noah@xitransportation.com', '555-1014', FALSE, TRUE),
	('Omicron Materials', 'Abigail White', 'abigail@omicronmaterials.com', '555-1015', FALSE, TRUE),
	('Pi Equipment', 'Daniel Harris', 'daniel@piequipment.com', '555-1016', FALSE, TRUE),
	('Rho Furniture', 'Mia Martin', 'mia@rhofurniture.com', '555-1017', FALSE, TRUE),
	('Sigma IT Solutions', 'Matthew Thompson', 'matthew@sigmait.com', '555-1018', FALSE, TRUE),
	('Tau Printing', 'Charlotte Garcia', 'charlotte@tauprinting.com', '555-1019', FALSE, TRUE),
	('Upsilon Distributors', 'Benjamin Martinez', 'benjamin@upsilondist.com', '555-1020', FALSE, TRUE);
	`
	
	if err := db.Exec(query).Error; err != nil {
		return err
	}
	
	return nil
}

// seedAddresses inserts address records
func seedAddresses(db *gorm.DB) error {
    log.Println("Seeding addresses...")
    
    // First batch
    query1 := `
    -- First INSERT
    INSERT INTO addresses (
        company_id, 
        address_type, 
        street, 
        city, 
        state_province, 
        postal_code, 
        country
    ) VALUES 
    (1, 'main', '123 Main Street', 'San Francisco', 'California', '94105', 'United States'),
    (1, 'billing', '123 Main Street, Suite 100', 'San Francisco', 'California', '94105', 'United States'),
    (1, 'shipping', '456 Warehouse Blvd', 'Oakland', 'California', '94607', 'United States');`
    
    if err := db.Exec(query1).Error; err != nil {
        return fmt.Errorf("company addresses failed: %w", err)
    }
    
	// Second batch
    query2 := `
    -- Second INSERT
    INSERT INTO addresses (
        company_id, 
        address_type, 
        street, 
        city, 
        state_province, 
        postal_code, 
        country
    ) VALUES 
    (2, 'billing', '789 Tech Park', 'Austin', 'Texas', '78701', 'United States'),
    (2, 'shipping', '790 Tech Park', 'Austin', 'Texas', '78701', 'United States'),
    (3, 'billing', '456 Innovation Drive', 'Boston', 'Massachusetts', '02110', 'United States'),
    (3, 'shipping', '789 Shipping Lane', 'Boston', 'Massachusetts', '02110', 'United States'),
    (4, 'billing', '101 Industrial Pkwy', 'Chicago', 'Illinois', '60607', 'United States'),
    (4, 'shipping', '102 Warehouse District', 'Chicago', 'Illinois', '60607', 'United States'),
    (5, 'billing', '222 Research Blvd', 'Seattle', 'Washington', '98101', 'United States'),
    (5, 'shipping', '223 Distribution Center', 'Seattle', 'Washington', '98101', 'United States'),
    (6, 'billing', '333 Coding Lane', 'Portland', 'Oregon', '97201', 'United States'),
    (6, 'shipping', '334 Download Drive', 'Portland', 'Oregon', '97201', 'United States'),
    (7, 'billing', '444 Advisory Ave', 'New York', 'New York', '10001', 'United States'),
    (7, 'shipping', '445 Materials Dept', 'New York', 'New York', '10001', 'United States'),
    (8, 'billing', '555 Factory Rd', 'Detroit', 'Michigan', '48201', 'United States'),
    (8, 'shipping', '556 Assembly Line', 'Detroit', 'Michigan', '48201', 'United States'),
    (9, 'billing', '666 Shipping Lane', 'Miami', 'Florida', '33101', 'United States'),
    (9, 'shipping', '667 Port Access Rd', 'Miami', 'Florida', '33101', 'United States'),
    (10, 'billing', '777 Service Street', 'Denver', 'Colorado', '80202', 'United States'),
    (10, 'shipping', '778 Delivery Drive', 'Denver', 'Colorado', '80202', 'United States'),
    (11, 'billing', '888 Shopping Mall', 'Las Vegas', 'Nevada', '89101', 'United States'),
    (11, 'shipping', '889 Retail Row', 'Las Vegas', 'Nevada', '89101', 'United States');
	`
    
    if err := db.Exec(query2).Error; err != nil {
        return fmt.Errorf("customer addresses failed: %w", err)
    }
    
    return nil
}

// seedItems inserts item/product records
func seedItems(db *gorm.DB) error {
	log.Println("Seeding items...")
	
	query := `
	INSERT INTO items (name, description, unit_price, type, stock, image_path)
	VALUES
	-- Software Products
	('Standard Widget', 'Basic widget for standard use cases', 50.00, 'software', 100, '/uploads/items/standard_widget.png'),
	('Enterprise Widget', 'Advanced widget with premium features', 200.00, 'software', 50, '/uploads/items/enterprise_widget.png'),
	('Widget API Access', 'API access to widget platform - monthly subscription', 100.00, 'subscription', 999, '/uploads/items/widget_api.png'),
	('Mobile Widget', 'Widget optimized for mobile devices', 75.00, 'software', 100, '/uploads/items/mobile_widget.png'),
	('Widget Suite', 'Complete bundle of all widget products', 350.00, 'bundle', 25, '/uploads/items/widget_suite.png'),
	
	-- Physical Products
	('Server Rack', 'Standard 42U server rack', 1200.00, 'hardware', 10, '/uploads/items/server_rack.png'),
	('Network Switch', '24-port gigabit ethernet switch', 350.00, 'hardware', 30, '/uploads/items/network_switch.png'),
	('UPS Battery Backup', '1500VA battery backup system', 275.00, 'hardware', 15, '/uploads/items/ups_battery.png'),
	('Cat6 Cable (1m)', 'Category 6 ethernet cable - 1 meter', 12.50, 'hardware', 200, '/uploads/items/cat6_cable.png'),
	('Fiber Optic Cable (5m)', 'Multi-mode fiber optic cable - 5 meters', 35.00, 'hardware', 50, '/uploads/items/fiber_cable.png'),
	
	-- Services
	('Basic Support Plan', '9-5 weekday support - monthly fee', 150.00, 'service', 999, '/uploads/items/basic_support.png'),
	('Premium Support Plan', '24/7 support with 1-hour response time - monthly fee', 500.00, 'service', 999, '/uploads/items/premium_support.png'),
	('System Implementation', 'Professional implementation services - hourly rate', 125.00, 'service', 999, '/uploads/items/implementation.png'),
	('Staff Training', 'On-site staff training - per day', 1000.00, 'service', 999, '/uploads/items/training.png'),
	('System Audit', 'Comprehensive system security audit', 2500.00, 'service', 999, '/uploads/items/audit.png'),
	
	-- Room Rental Services
	('Conference Room A', 'Small conference room (seats 8) - hourly rate', 50.00, 'rental', 1, '/uploads/items/conference_a.png'),
	('Conference Room B', 'Large conference room (seats 20) - hourly rate', 100.00, 'rental', 1, '/uploads/items/conference_b.png'),
	('Executive Boardroom', 'Executive boardroom (seats 12) - hourly rate', 150.00, 'rental', 1, '/uploads/items/boardroom.png'),
	('Training Lab', 'Computer training lab (seats 25) - daily rate', 750.00, 'rental', 1, '/uploads/items/training_lab.png'),
	('Event Space', 'Open event space (capacity 100) - daily rate', 2000.00, 'rental', 1, '/uploads/items/event_space.png');
	`
	
	if err := db.Exec(query).Error; err != nil {
		return err
	}
	
	return nil
}

// seedOrders inserts order records
func seedOrders(db *gorm.DB) error {
	log.Println("Seeding orders...")
	
	query := `
	INSERT INTO orders (customer_company_id, order_date, total_price, status)
	VALUES
	-- Alpha Technologies orders
	(2, '2025-03-01 10:30:00', 650.00, 'completed'),
	(2, '2025-03-15 14:45:00', 500.00, 'completed'),
	-- Beta Solutions orders
	(3, '2025-03-05 09:15:00', 875.00, 'completed'),
	(3, '2025-03-20 11:20:00', 2500.00, 'completed'),
	-- Gamma Industries orders
	(4, '2025-03-10 13:00:00', 1487.50, 'shipped'),
	(4, '2025-03-25 16:30:00', 350.00, 'processing'),
	-- Delta Innovations orders
	(5, '2025-04-02 08:45:00', 2750.00, 'processing'),
	(5, '2025-04-10 10:15:00', 150.00, 'pending'),
	-- Epsilon Software orders
	(6, '2025-04-05 14:00:00', 200.00, 'pending'),
	(6, '2025-04-15 15:30:00', 1000.00, 'pending'),
	-- Example order with 2+ items and quantity < 20
	(2, '2025-04-29 09:00:00', 1250.00, 'pending');
	`
	
	if err := db.Exec(query).Error; err != nil {
		return err
	}
	
	return nil
}

// seedOrderItems inserts order item records
func seedOrderItems(db *gorm.DB) error {
	log.Println("Seeding order items...")
	
	query := `
	INSERT INTO order_items (order_id, item_id, quantity, unit_price, item_total)
	VALUES
	-- Order 1 items: Alpha Technologies
	(1, 1, 5, 50.00, 250.00),
	(1, 4, 2, 75.00, 150.00),
	(1, 11, 1, 150.00, 150.00),
	(1, 9, 8, 12.50, 100.00),

	-- Order 2 items: Alpha Technologies
	(2, 12, 1, 500.00, 500.00),

	-- Order 3 items: Beta Solutions
	(3, 2, 3, 200.00, 600.00),
	(3, 3, 1, 100.00, 100.00),
	(3, 9, 14, 12.50, 175.00),

	-- Order 4 items: Beta Solutions
	(4, 15, 1, 2500.00, 2500.00),

	-- Order 5 items: Gamma Industries
	(5, 7, 4, 350.00, 1400.00),
	(5, 10, 2.5, 35.00, 87.50),

	-- Order 6 items: Gamma Industries
	(6, 7, 1, 350.00, 350.00),

	-- Order 7 items: Delta Innovations
	(7, 6, 2, 1200.00, 2400.00),
	(7, 10, 10, 35.00, 350.00),

	-- Order 8 items: Delta Innovations
	(8, 11, 1, 150.00, 150.00),

	-- Order 9 items: Epsilon Software
	(9, 1, 4, 50.00, 200.00),

	-- Order 10 items: Epsilon Software
	(10, 14, 1, 1000.00, 1000.00),

	-- Order 11 items: Alpha Technologies (Example order with multiple items, all quantity < 20)
	(11, 1, 15, 50.00, 750.00),
	(11, 4, 5, 75.00, 375.00),
	(11, 11, 1, 150.00, 150.00);
	`
	
	if err := db.Exec(query).Error; err != nil {
		return err
	}
	
	return nil
}

// seedInvoices inserts invoice records
func seedInvoices(db *gorm.DB) error {
	log.Println("Seeding invoices...")
	
	query := `
	INSERT INTO invoices (sender_company_id, recipient_company_id, billing_address_id, shipping_address_id, order_id, invoice_number, invoice_date, due_date, invoice_subject, subtotal, tax_total, grand_total, amount_paid, amount_due, status, notes)
	VALUES
	-- Alpha Technologies invoices
	(1, 2, 4, 5, 1, 'INV-2025-0001', '2025-03-01', '2025-03-31', 'March Services and Products', 650.00, 52.00, 702.00, 702.00, 0.00, 'paid', 'Thank you for your business!'),
	(1, 2, 4, 5, 2, 'INV-2025-0002', '2025-03-15', '2025-04-14', 'Premium Support Plan', 500.00, 40.00, 540.00, 540.00, 0.00, 'paid', 'Premium support plan monthly fee'),

	-- Beta Solutions invoices
	(1, 3, 6, 7, 3, 'INV-2025-0003', '2025-03-05', '2025-04-04', 'Software Licenses and Services', 875.00, 70.00, 945.00, 945.00, 0.00, 'paid', ''),
	(1, 3, 6, 7, 4, 'INV-2025-0004', '2025-03-20', '2025-04-19', 'System Audit Services', 2500.00, 200.00, 2700.00, 2700.00, 0.00, 'paid', 'Comprehensive security audit completed'),

	-- Gamma Industries invoices
	(1, 4, 8, 9, 5, 'INV-2025-0005', '2025-03-10', '2025-04-09', 'Network Equipment Order', 1487.50, 119.00, 1606.50, 1606.50, 0.00, 'paid', ''),
	(1, 4, 8, 9, 6, 'INV-2025-0006', '2025-03-25', '2025-04-24', 'Network Switch', 350.00, 28.00, 378.00, 0.00, 378.00, 'sent', ''),

	-- Delta Innovations invoices
	(1, 5, 10, 11, 7, 'INV-2025-0007', '2025-04-02', '2025-05-02', 'Server Equipment Order', 2750.00, 220.00, 2970.00, 1500.00, 1470.00, 'partial', 'Partial payment received'),

	-- Invoice example for presentation
	(1, 2, 4, 5, 11, 'INV-2025-0500', '2025-04-30', '2025-05-30', 'April Products and Services', 1250.00, 100.00, 1350.00, 0.00, 1350.00, 'sent', 'Example invoice for demonstration');
	`
	
	if err := db.Exec(query).Error; err != nil {
		return err
	}
	
	return nil
}

// seedInvoiceItems inserts invoice item records
func seedInvoiceItems(db *gorm.DB) error {
	log.Println("Seeding invoice items...")
	
	query := `
	INSERT INTO invoice_items (invoice_id, item_id, description, quantity, unit_price, item_total, tax_rate_percentage)
	VALUES
	-- Invoice 1 items (Order 1): Alpha Technologies
	(1, 1, 'Standard Widget', 5, 50.00, 250.00, 8.00),
	(1, 4, 'Mobile Widget', 2, 75.00, 150.00, 8.00),
	(1, 11, 'Basic Support Plan', 1, 150.00, 150.00, 8.00),
	(1, 9, 'Cat6 Cable (1m)', 8, 12.50, 100.00, 8.00),

	-- Invoice 2 items (Order 2): Alpha Technologies
	(2, 12, 'Premium Support Plan', 1, 500.00, 500.00, 8.00),

	-- Invoice 3 items (Order 3): Beta Solutions
	(3, 2, 'Enterprise Widget', 3, 200.00, 600.00, 8.00),
	(3, 3, 'Widget API Access', 1, 100.00, 100.00, 8.00),
	(3, 9, 'Cat6 Cable (1m)', 14, 12.50, 175.00, 8.00),

	-- Invoice 4 items (Order 4): Beta Solutions
	(4, 15, 'System Audit', 1, 2500.00, 2500.00, 8.00),

	-- Invoice 5 items (Order 5): Gamma Industries
	(5, 7, 'Network Switch', 4, 350.00, 1400.00, 8.00),
	(5, 10, 'Fiber Optic Cable (5m)', 2.5, 35.00, 87.50, 8.00),

	-- Invoice 6 items (Order 6): Gamma Industries
	(6, 7, 'Network Switch', 1, 350.00, 350.00, 8.00),

	-- Invoice 7 items (Order 7): Delta Innovations
	(7, 6, 'Server Rack', 2, 1200.00, 2400.00, 8.00),
	(7, 10, 'Fiber Optic Cable (5m)', 10, 35.00, 350.00, 8.00),

	-- Invoice 8 items (Order 11): Alpha Technologies example
	(8, 1, 'Standard Widget', 15, 50.00, 750.00, 8.00),
	(8, 4, 'Mobile Widget', 5, 75.00, 375.00, 8.00),
	(8, 11, 'Basic Support Plan', 1, 150.00, 150.00, 8.00);
	`
	
	if err := db.Exec(query).Error; err != nil {
		return err
	}
	
	return nil
}

// seedPayments inserts payment records
func seedPayments(db *gorm.DB) error {
	log.Println("Seeding payments...")
	
	query := `
	INSERT INTO payments (invoice_id, payment_date, amount, method, status, transaction_reference)
	VALUES
	-- Invoice 1 payments: Alpha Technologies
	(1, '2025-03-15 14:30:00', 702.00, 'bank_transfer', 'completed', 'BANK-20250315-A2C1'),

	-- Invoice 2 payments: Alpha Technologies
	(2, '2025-03-20 09:45:00', 540.00, 'credit_card', 'completed', 'CC-20250320-B3D2'),

	-- Invoice 3 payments: Beta Solutions
	(3, '2025-03-25 11:15:00', 945.00, 'bank_transfer', 'completed', 'BANK-20250325-C4E3'),

	-- Invoice 4 payments: Beta Solutions
	(4, '2025-04-01 10:30:00', 2700.00, 'bank_transfer', 'completed', 'BANK-20250401-D5F4'),

	-- Invoice 5 payments: Gamma Industries
	(5, '2025-04-05 16:00:00', 1606.50, 'credit_card', 'completed', 'CC-20250405-E6G5'),

	-- Invoice 7 payments: Delta Innovations (partial payment)
	(7, '2025-04-10 13:20:00', 1500.00, 'bank_transfer', 'completed', 'BANK-20250410-F7H6');
	`
	
	if err := db.Exec(query).Error; err != nil {
		return err
	}
	
	return nil
}
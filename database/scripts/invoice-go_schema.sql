CREATE TABLE companies (
    company_id INT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    company_name VARCHAR(100) NOT NULL,
    contact_person VARCHAR(100),
    email VARCHAR(100) NOT NULL UNIQUE,
    phone VARCHAR(20),
    is_customer BOOLEAN DEFAULT FALSE,
    is_vendor BOOLEAN DEFAULT FALSE,
    default_billing_address_id INT UNSIGNED,
    default_shipping_address_id INT UNSIGNED,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

-- address
CREATE TABLE addresses (
    address_id INT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    company_id INT UNSIGNED NOT NULL,
    address_type VARCHAR(20) NOT NULL,
    street VARCHAR(200) NOT NULL,
    city VARCHAR(100) NOT NULL,
    state_province VARCHAR(100) NOT NULL,
    postal_code VARCHAR(20) NOT NULL,
    country VARCHAR(100) NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    FOREIGN KEY (company_id) REFERENCES companies(company_id) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

-- items
CREATE TABLE items (
    item_id INT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    item_name VARCHAR(100) NOT NULL,
    item_description VARCHAR(500),
    unit_price DECIMAL(10,2) NOT NULL,
    item_type VARCHAR(50) NOT NULL,
    image_path VARCHAR(255),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

-- orders
CREATE TABLE orders (
    order_id INT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    customer_company_id INT UNSIGNED NOT NULL,
    order_date DATETIME NOT NULL,
    total_price DECIMAL(12,2) NOT NULL,
    status VARCHAR(20) NOT NULL DEFAULT 'pending',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    FOREIGN KEY (customer_company_id) REFERENCES companies(company_id) ON DELETE RESTRICT
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

-- Order Items
CREATE TABLE  order_items (
    order_item_id INT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    order_id INT UNSIGNED NOT NULL,
    item_id INT UNSIGNED NOT NULL,
    quantity INT UNSIGNED NOT NULL,
    unit_price DECIMAL(10,2) NOT NULL,
    item_total DECIMAL(12,2) NOT NULL,
    FOREIGN KEY (order_id) REFERENCES orders(order_id) ON DELETE CASCADE,
    FOREIGN KEY (item_id) REFERENCES items(item_id) ON DELETE RESTRICT
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

-- Invoices
CREATE TABLE invoices (
    invoice_id INT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    sender_company_id INT UNSIGNED NOT NULL,
    recipient_company_id INT UNSIGNED NOT NULL,
    billing_address_id INT UNSIGNED NOT NULL,
    shipping_address_id INT UNSIGNED NOT NULL,
    order_id INT UNSIGNED NOT NULL,
    invoice_number VARCHAR(50) NOT NULL UNIQUE,
    invoice_date DATETIME NOT NULL,
    due_date DATETIME NOT NULL,
    invoice_subject VARCHAR(200),
    subtotal DECIMAL(10,2) NOT NULL,
    tax_total DECIMAL(10,2) NOT NULL,
    grand_total DECIMAL(10,2) NOT NULL,
    amount_paid DECIMAL(10,2) DEFAULT 0,
    amount_due DECIMAL(10,2) NOT NULL,
    status VARCHAR(20) NOT NULL DEFAULT 'draft',
    notes VARCHAR(500),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    FOREIGN KEY (sender_company_id) REFERENCES companies(company_id) ON DELETE RESTRICT,
    FOREIGN KEY (recipient_company_id) REFERENCES companies(company_id) ON DELETE RESTRICT,
    FOREIGN KEY (billing_address_id) REFERENCES addresses(address_id) ON DELETE RESTRICT,
    FOREIGN KEY (shipping_address_id) REFERENCES addresses(address_id) ON DELETE RESTRICT,
    FOREIGN KEY (order_id) REFERENCES orders(order_id) ON DELETE RESTRICT
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

-- invoice items
CREATE TABLE invoice_items (
    invoice_item_id INT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    invoice_id INT UNSIGNED NOT NULL,
    product_id INT UNSIGNED NOT NULL,
    description VARCHAR(200) NOT NULL,
    quantity INT NOT NULL,
    unit_price DECIMAL(10,2) NOT NULL,
    item_total DECIMAL(10,2) NOT NULL,
    tax_rate_percentage DECIMAL(5,2) DEFAULT 0,
    FOREIGN KEY (invoice_id) REFERENCES invoices(invoice_id) ON DELETE CASCADE,
    FOREIGN KEY (product_id) REFERENCES items(item_id) ON DELETE RESTRICT
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE payments (
    payment_id INT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    invoice_id INT UNSIGNED NOT NULL,
    payment_date DATETIME NOT NULL,
    amount DECIMAL(10,2) NOT NULL,
    method VARCHAR(50) NOT NULL,
    status VARCHAR(20) NOT NULL DEFAULT 'pending',
    transaction_reference VARCHAR(100),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (invoice_id) REFERENCES invoices(invoice_id) ON DELETE RESTRICT
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;


-- create indexing
-- Create indexes to improve query performance
CREATE INDEX idx_companies_is_customer ON companies(is_customer);
CREATE INDEX idx_companies_is_vendor ON companies(is_vendor);
CREATE INDEX idx_addresses_company_id ON addresses(company_id);
CREATE INDEX idx_addresses_city ON addresses(city);
CREATE INDEX idx_addresses_country ON addresses(country);
CREATE INDEX idx_items_item_type ON items(item_type);
CREATE INDEX idx_orders_customer_company_id ON orders(customer_company_id);
CREATE INDEX idx_orders_status ON orders(status);
CREATE INDEX idx_order_items_order_id ON order_items(order_id);
CREATE INDEX idx_order_items_item_id ON order_items(item_id);
CREATE INDEX idx_invoices_order_id ON invoices(order_id);
CREATE INDEX idx_invoices_status ON invoices(status);
CREATE INDEX idx_invoices_recipient_company_id ON invoices(recipient_company_id);
CREATE INDEX idx_invoice_items_invoice_id ON invoice_items(invoice_id);
CREATE INDEX idx_invoice_items_product_id ON invoice_items(product_id);
CREATE INDEX idx_payments_invoice_id ON payments(invoice_id);
CREATE INDEX idx_payments_status ON payments(status);



-- Insert sample data for companies (first your own company, then customers)
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


-- Insert sample addresses (after companies)
-- First, add addresses for your company
INSERT INTO addresses (company_id, address_type, street, city, state_province, postal_code, country)
VALUES
(1, 'main', '123 Main Street', 'San Francisco', 'California', '94105', 'United States'),
(1, 'billing', '123 Main Street, Suite 100', 'San Francisco', 'California', '94105', 'United States'),
(1, 'shipping', '456 Warehouse Blvd', 'Oakland', 'California', '94607', 'United States');

-- Update your company with default addresses
UPDATE companies SET default_billing_address_id = 2, default_shipping_address_id = 3 WHERE company_id = 1;

-- Add addresses for customer companies
INSERT INTO addresses (company_id, address_type, street, city, state_province, postal_code, country)
VALUES
-- Alpha Technologies
(2, 'billing', '789 Tech Park', 'Austin', 'Texas', '78701', 'United States'),
(2, 'shipping', '790 Tech Park', 'Austin', 'Texas', '78701', 'United States'),

-- Beta Solutions
(3, 'billing', '456 Innovation Drive', 'Boston', 'Massachusetts', '02110', 'United States'),
(3, 'shipping', '789 Shipping Lane', 'Boston', 'Massachusetts', '02110', 'United States'),

-- Gamma Industries
(4, 'billing', '101 Industrial Pkwy', 'Chicago', 'Illinois', '60607', 'United States'),
(4, 'shipping', '102 Warehouse District', 'Chicago', 'Illinois', '60607', 'United States'),

-- Delta Innovations
(5, 'billing', '222 Research Blvd', 'Seattle', 'Washington', '98101', 'United States'),
(5, 'shipping', '223 Distribution Center', 'Seattle', 'Washington', '98101', 'United States'),

-- Epsilon Software
(6, 'billing', '333 Coding Lane', 'Portland', 'Oregon', '97201', 'United States'),
(6, 'shipping', '334 Download Drive', 'Portland', 'Oregon', '97201', 'United States'),

-- Zeta Consulting
(7, 'billing', '444 Advisory Ave', 'New York', 'New York', '10001', 'United States'),
(7, 'shipping', '445 Materials Dept', 'New York', 'New York', '10001', 'United States'),

-- Eta Manufacturing
(8, 'billing', '555 Factory Rd', 'Detroit', 'Michigan', '48201', 'United States'),
(8, 'shipping', '556 Assembly Line', 'Detroit', 'Michigan', '48201', 'United States'),

-- Theta Logistics
(9, 'billing', '666 Shipping Lane', 'Miami', 'Florida', '33101', 'United States'),
(9, 'shipping', '667 Port Access Rd', 'Miami', 'Florida', '33101', 'United States'),

-- Iota Services
(10, 'billing', '777 Service Street', 'Denver', 'Colorado', '80202', 'United States'),
(10, 'shipping', '778 Delivery Drive', 'Denver', 'Colorado', '80202', 'United States'),

-- Kappa Retail
(11, 'billing', '888 Shopping Mall', 'Las Vegas', 'Nevada', '89101', 'United States'),
(11, 'shipping', '889 Retail Row', 'Las Vegas', 'Nevada', '89101', 'United States');


-- Update customer companies with default addresses
UPDATE companies SET default_billing_address_id = 4, default_shipping_address_id = 5 WHERE company_id = 2;
UPDATE companies SET default_billing_address_id = 6, default_shipping_address_id = 7 WHERE company_id = 3;
UPDATE companies SET default_billing_address_id = 8, default_shipping_address_id = 9 WHERE company_id = 4;
UPDATE companies SET default_billing_address_id = 10, default_shipping_address_id = 11 WHERE company_id = 5;
UPDATE companies SET default_billing_address_id = 12, default_shipping_address_id = 13 WHERE company_id = 6;
UPDATE companies SET default_billing_address_id = 14, default_shipping_address_id = 15 WHERE company_id = 7;
UPDATE companies SET default_billing_address_id = 16, default_shipping_address_id = 17 WHERE company_id = 8;
UPDATE companies SET default_billing_address_id = 18, default_shipping_address_id = 19 WHERE company_id = 9;
UPDATE companies SET default_billing_address_id = 20, default_shipping_address_id = 21 WHERE company_id = 10;
UPDATE companies SET default_billing_address_id = 22, default_shipping_address_id = 23 WHERE company_id = 11;

-- Insert sample items/products
INSERT INTO items (item_name, item_description, unit_price, item_type, image_path)
VALUES
-- Software Products
('Standard Widget', 'Basic widget for standard use cases', 50.00, 'software', '/uploads/items/standard_widget.png'),
('Enterprise Widget', 'Advanced widget with premium features', 200.00, 'software', '/uploads/items/enterprise_widget.png'),
('Widget API Access', 'API access to widget platform - monthly subscription', 100.00, 'subscription', '/uploads/items/widget_api.png'),
('Mobile Widget', 'Widget optimized for mobile devices', 75.00, 'software', '/uploads/items/mobile_widget.png'),
('Widget Suite', 'Complete bundle of all widget products', 350.00, 'bundle', '/uploads/items/widget_suite.png'),

-- Physical Products
('Server Rack', 'Standard 42U server rack', 1200.00, 'hardware', '/uploads/items/server_rack.png'),
('Network Switch', '24-port gigabit ethernet switch', 350.00, 'hardware', '/uploads/items/network_switch.png'),
('UPS Battery Backup', '1500VA battery backup system', 275.00, 'hardware', '/uploads/items/ups_battery.png'),
('Cat6 Cable (1m)', 'Category 6 ethernet cable - 1 meter', 12.50, 'hardware', '/uploads/items/cat6_cable.png'),
('Fiber Optic Cable (5m)', 'Multi-mode fiber optic cable - 5 meters', 35.00, 'hardware', '/uploads/items/fiber_cable.png'),

-- Services
('Basic Support Plan', '9-5 weekday support - monthly fee', 150.00, 'service', '/uploads/items/basic_support.png'),
('Premium Support Plan', '24/7 support with 1-hour response time - monthly fee', 500.00, 'service', '/uploads/items/premium_support.png'),
('System Implementation', 'Professional implementation services - hourly rate', 125.00, 'service', '/uploads/items/implementation.png'),
('Staff Training', 'On-site staff training - per day', 1000.00, 'service', '/uploads/items/training.png'),
('System Audit', 'Comprehensive system security audit', 2500.00, 'service', '/uploads/items/audit.png'),

-- Room Rental Services
('Conference Room A', 'Small conference room (seats 8) - hourly rate', 50.00, 'rental', '/uploads/items/conference_a.png'),
('Conference Room B', 'Large conference room (seats 20) - hourly rate', 100.00, 'rental', '/uploads/items/conference_b.png'),
('Executive Boardroom', 'Executive boardroom (seats 12) - hourly rate', 150.00, 'rental', '/uploads/items/boardroom.png'),
('Training Lab', 'Computer training lab (seats 25) - daily rate', 750.00, 'rental', '/uploads/items/training_lab.png'),
('Event Space', 'Open event space (capacity 100) - daily rate', 2000.00, 'rental', '/uploads/items/event_space.png');

-- Insert sample orders
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

-- Insert sample order items
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


-- Insert sample invoices
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


-- Insert sample invoice items
INSERT INTO invoice_items (invoice_id, product_id, description, quantity, unit_price, item_total, tax_rate_percentage)
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
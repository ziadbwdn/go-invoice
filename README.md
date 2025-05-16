# Invoice-Go

A comprehensive invoice management system built with Go, Gin, and GORM.

## Overview

Invoice-Go is a comprehensive invoice management system designed to streamline the creation, tracking, and management of invoices, orders, and payments. Built with Golang and Gin framework, it provides a robust backend infrastructure with a well-designed relational database and intuitive RESTful API endpoints.

## Features

- **Company & Address Management**: Track customers, vendors, and their multiple addresses
- **Product & Inventory Management**: Manage products, services, and stock levels
- **Order Processing**: Create and track orders with line items
- **Invoice Generation**: Create professional invoices with automatic calculations
- **Payment Tracking**: Record and monitor payment status
- **File Management**: Upload and serve product images
- **Comprehensive Reporting**: Generate detailed reports on invoices and orders

## Prerequisites

- Go 1.16 or higher
- MySQL 5.7 or higher
- Git

## Project Structure

```
.
├── database
│   ├── db.go             # Database connection and configuration
│   ├── queries.go        # SQL queries
│   ├── scripts
│   │   └── invoice-go_schema.sql  # Database schema
│   └── seeder.go         # Data seeding functionality
├── docs
│   └── invoice-go-api.postman_collection  # API documentation
├── handlers
│   ├── address_handlers.go    # Address management endpoints
│   ├── company_handlers.go    # Company management endpoints
│   ├── image_handlers.go      # Image upload/download functionality
│   ├── invoice_handlers.go    # Invoice management endpoints
│   ├── item_handlers.go       # Product/Item management endpoints
│   ├── order_handlers.go      # Order management endpoints
│   └── payment_handlers.go    # Payment processing endpoints
├── models
│   └── models.go        # Data models and database structure
├── routes
│   └── router.go        # API route definitions
├── uploads              # Storage for uploaded files
│   ├── items
│   └── products
└── utils
    └── utils.go         # Helper functions and utilities
```

## Database Design

### Entity Relationship Diagram

```
+-------------+     +------------+     +----------+     +-------------+
|  Companies  |-----| Addresses  |     |  Items   |     |   Orders    |
+-------------+     +------------+     +----------+     +-------------+
       |                                    |                 |
       |                                    |                 |
       v                                    v                 v
+-------------+     +---------------+     +-------------+
|  Invoices   |-----| InvoiceItems |-----| OrderItems  |
+-------------+     +---------------+     +-------------+
       |
       |
       v
+-------------+
|  Payments   |
+-------------+
```

## Installation

1. Clone the repository:
   ```
   git clone https://github.com/yourusername/invoice-go.git
   cd invoice-go
   ```

2. Install dependencies:
   ```
   go mod download
   ```

3. Configure your database connection in the `database/db.go` file or through environment variables.

## Running the Application

**Important**: Make sure MySQL is running before starting the application.

To run the application:

```
go run main.go -seed
```

The `-seed` flag initializes the database with sample data. Remove this flag if you don't want to seed the database on startup.

## API Endpoints

### Health Check
- `GET /health` - Check if the service is running

### Items/Products
- `GET /items` - Get all items
- `GET /items/:id` - Get a specific item
- `POST /items` - Create a new item
- `PUT /items/:id` - Update an item
- `POST /items/:id/upload` - Upload an image for an item
- `GET /items/:id/image` - Download an item's image

### Companies
- `GET /companies` - Get all companies
- `GET /companies/:id` - Get a specific company
- `POST /companies` - Create a new company
- `PUT /companies/:id` - Update a company

### Addresses
- `GET /addresses` - Get all addresses
- `GET /addresses/:id` - Get a specific address
- `POST /addresses` - Create a new address
- `PUT /addresses/:id` - Update an address

### Orders
- `GET /orders` - Get all orders
- `GET /orders/:id` - Get a specific order
- `POST /orders` - Create a new order

### Invoices
- `POST /invoice/:id` - Create an invoice
- `GET /invoice/:id` - Get invoices
- `GET /invoice/:id/details` - Get invoice details
- `GET /invoice/:id/reports` - Get invoice reports
- `GET /invoice/:id/status` - Get invoice status
- `PATCH /invoice/:id/status` - Update invoice status

### Payments
- `POST /payment/:id` - Create a payment
- `GET /payment/:id` - Get payments
- `GET /payment/:id/details` - Get payment details
- `PUT /payment/:id/status` - Update payment status

## Development

### Test Upload Endpoint
- `GET /test-upload` - Test if the upload functionality is working

### CORS
The API supports Cross-Origin Resource Sharing (CORS), allowing requests from any origin.

## Security

The application includes middleware for preventing path traversal attacks when handling file uploads and downloads.

## Testing with Postman

The project includes a Postman collection for easy API testing:

1. Import the Postman collection from `docs/invoice-go-api.postman_collection` into your Postman application.

2. Ensure the application is running with `go run main.go -seed`.

3. The default server runs on `http://localhost:8080` (or your configured port).

4. Use the collection to test the various endpoints:
   - Start with the health check endpoint to verify the server is running
   - Test CRUD operations for companies, items, addresses, etc.
   - Create orders and then invoices
   - Test the payment processing endpoints

5. For endpoints requiring file uploads:
   - Use Postman's form-data option
   - Set the key type to "File" 
   - Select your file from the file picker

6. For secured endpoints (if implemented), you may need to add authentication headers.

7. Check response status codes and JSON payloads to verify proper functionality.

## Troubleshooting

If you encounter issues starting the application:

1. Ensure MySQL is running and accessible
2. Check database connection settings
3. Verify that all required Go modules are installed
4. Check for port conflicts if the server fails to start

## Postman Collection

A Postman collection is available in the `docs/` directory for testing the API endpoints. Import the collection into Postman to get started quickly with testing.

## Example Invoice

Below is an example of how invoices are formatted:

```
+-----------------------------------------------------------------------------------------------+
|**INVOICE**                                                                                    |
|                                               From     |Furnitura Romana                      |
|                                                        |32 St Maximius Place                  |
|                                                        |Liverpool LG1 2ER                     |
|                                                        |United Kingdom                        |
|                                                                                               |
|Invoice ID | 0031                           For         |YHA Canterburry Hostel                |
|Issue Date | 02/09/2019                                 |54 New Dover Road                     |
|Due Date   | 02/09/2019 (upon receipt)                  |Canterburry CT1 3DT                   |
|Subject    | Autumn Marketing Campaign                  |United Kingdom                        |
|                                                                                               |
|                                                                                               |
|+---------------+---------------+---------------+---------------+--------------+               |
||Item Type      |Description    |Quantity       |Unit Price     |Amount        |               |
|+---------------+---------------+---------------+---------------+--------------+               |
||Service        |Design         |44.00          |£240.00        |£10,560.00    |               |
||Service        |Development    |59.00          |£310.00        |£18,290.00    |               |
||Service        |Meetings       |5.50           |£70.00         |£385.00       |               |
|+---------------+---------------+---------------+---------------+--------------+               |
|                                                                                               |
|                                                                                               |
|                                           Subtotal        |£29,235.00                         |
|                                           Tax (10%)       |£2,923.50                          |
|                                           Payments        |-£32,158.50                        |
|                                                                                               |
|                                           **Amount Due**  |£0.00                              |
|                                                                                               |
+-----------------------------------------------------------------------------------------------+
```

## License

This project is licensed under the MIT License - see the LICENSE file for details.

## Contributors

Muhammad Ziad

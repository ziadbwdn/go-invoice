// image utils

package utils

import (
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"log"
	"gorm.io/gorm"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"invoice-go/models"
	"time"
	"math/rand"
	"strconv"
)

const (
	// Maximum file size (5MB)
	MaxFileSize = 5 << 20
	// Base upload directory
	UploadDir = "./uploads/items"
)

// ValidateImage checks if the file is a valid image and within size limits
func ValidateImage(file *multipart.FileHeader) error {
	// Check file size
	if file.Size > MaxFileSize {
		return fmt.Errorf("file exceeds 5MB limit")
	}

	// Open the file to check its content type
	src, err := file.Open()
	if err != nil {
		return err
	}
	defer src.Close()

	// Read first 512 bytes to determine content type
	buffer := make([]byte, 512)
	_, err = src.Read(buffer)
	if err != nil && err != io.EOF {
		return err
	}

	// Reset the file reader
	_, err = src.Seek(0, io.SeekStart)
	if err != nil {
		return err
	}

	// Check content type
	contentType := http.DetectContentType(buffer)
	if contentType != "image/jpeg" && contentType != "image/jpg" && contentType != "image/png" {
		return fmt.Errorf("only PNG/JPG/JPEG allowed")
	}

	return nil
}

func ValidateAddress(tx *gorm.DB, c *gin.Context, addressID *uint, companyID *uint, errorMsg string) bool {
	/**
	if addressID == nil {
		return true
	}

	var address models.Address
	query := tx.Model(&models.Address{})
	
	if companyID != nil {
		query = query.Where("address_id = ? AND company_id = ?", *addressID, *companyID)
	} else {
		query = query.Where("address_id = ?", *addressID)
	}

	if err := query.First(&address).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusBadRequest, gin.H{"error": errorMsg})
		return false
	}
	
	return true
	*/

	// v02
	/**
	// Check billing address if provided
	if models.DefaultBillingAddressID != nil {
		valid, err := h.isValidAddressForCompany(tx, *models.DefaultBillingAddressID, models.CompanyID)
		if !valid {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid billing address ID"})
			return false
		}
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Error validating billing address"})
			return false
		}
	}

	// Check shipping address if provided
	if company.DefaultShippingAddressID != nil {
		valid, err := h.isValidAddressForCompany(tx, *company.DefaultShippingAddressID, company.CompanyID)
		if !valid {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid shipping address ID"})
			return false
		}
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Error validating shipping address"})
			return false
		}
	}

	return true
	*/
	var count int64
    tx.Model(&models.Address{}).
        Where("address_id = ? AND company_id = ?", addressID, companyID).
        Count(&count)
    return count > 0
}


func SafeUpdateString(field *string, newValue *string) {
    if newValue != nil {
        *field = *newValue
    }
}

func SafeUpdateStringPtr(field **string, newValue *string) {
    if newValue != nil {
        *field = newValue
    }
}

func SafeUpdateBool(field *bool, newValue *bool) {
    if newValue != nil {
        *field = *newValue
    }
}

// isValidAddressForCompany checks if an address belongs to the company
// For new companies (companyID = 0), it just checks if the address exists
/**
func (h *CompanyHandler) isValidAddressForCompany(tx *gorm.DB, addressID uint, companyID uint) (bool, error) {
	var address models.Address
	result := tx.Where("address_id = ?", addressID)
	
	// If companyID is provided (update case), also check company ownership
	if companyID > 0 {
		result = result.Where("company_id = ?", companyID)
	}
	
	if err := result.First(&address).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return false, nil
		}
		return false, err
	}
	
	return true, nil
}
*/

// saveProductImage saves the uploaded file to the appropriate directory
func SaveItemImage(c *gin.Context, file *multipart.FileHeader, itemID uint ) (string, error) {
	// Generate unique filename with UUID
	fileExt := filepath.Ext(file.Filename)
	newFilename := uuid.New().String() + "-ItemImage" + fileExt

	// Ensure the directory exists
	itemDir := filepath.Join(UploadDir, fmt.Sprintf("%d", itemID))
	if err := os.MkdirAll(itemDir, os.ModePerm); err != nil {
		return "", fmt.Errorf("failed to create directory: %w", err)
	}

	// Full path for the file
	dst := filepath.Join(itemDir, newFilename)

	// Save the file using Gin's utility function
	if err := c.SaveUploadedFile(file, dst); err != nil {
		return "", err
	}

	return dst, nil
}

// GetProductImagePath retrieves the latest image for a product
func GetItemImagePath(itemID uint ) (string, error) {
	itemDir := filepath.Join(UploadDir, fmt.Sprintf("%d", itemID))
	
	// Check if directory exists
	if _, err := os.Stat(itemDir); os.IsNotExist(err) {
		return "", fmt.Errorf("no image found for items")
	}

	// Read directory contents
	files, err := os.ReadDir(itemDir)
	if err != nil {
		return "", err
	}

	// No files found
	if len(files) == 0 {
		return "", fmt.Errorf("no image found for item")
	}

	// Find the most recent file (based on name for simplicity)
	// In a real app, you might want to check file creation dates
	var latestFile string
	for _, file := range files {
		if !file.IsDir() && (strings.HasSuffix(file.Name(), ".jpg") || 
		   strings.HasSuffix(file.Name(), ".jpeg") || 
		   strings.HasSuffix(file.Name(), ".png")) {
			if latestFile == "" || file.Name() > latestFile {
				latestFile = file.Name()
			}
		}
	}

	if latestFile == "" {
		return "", fmt.Errorf("no image found for item")
	}

	return filepath.Join(itemDir, latestFile), nil
}

// DeleteProductImages removes all images associated with a product
func DeleteItemImages(itemID uint ) error {
	itemDir := filepath.Join(UploadDir, fmt.Sprintf("%d", itemID))
	
	// Check if directory exists
	if _, err := os.Stat(itemDir); os.IsNotExist(err) {
		// Directory doesn't exist, nothing to delete
		return nil
	}

	// Remove the directory and all its contents
	return os.RemoveAll(itemDir)
}

// PathTraversalMiddleware prevents path traversal attacks
func PathTraversalMiddleware() gin.HandlerFunc {
    return func(c *gin.Context) {
        paramValue := c.Param("id")
        log.Printf("PathTraversalMiddleware checking param: %s", paramValue)
        
        if strings.Contains(paramValue, "..") {
            c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid path parameter"})
            c.Abort()
            return
        }
        log.Printf("PathTraversalMiddleware passed for param: %s", paramValue)
        c.Next()
    }
}

// GenerateInvoiceNumber creates a unique invoice number with prefix and date
func GenerateInvoiceNumber() string {
	// Format: INV-YYYY-XXXX where XXXX is a random number
	year := time.Now().Format("2006")
	randomNum := 1000 + rand.Intn(9000) // Random 4-digit number
	
	return fmt.Sprintf("INV-%s-%04d", year, randomNum)
}

// sanitizeFilename removes potentially harmful characters from filenames
func sanitizeFilename(filename string) string {
	// Remove path components and special characters
	filename = filepath.Base(filename)
	filename = strings.ReplaceAll(filename, " ", "_")
	
	// Remove any characters that aren't alphanumeric, underscore, hyphen, or dots
	return strings.Map(func(r rune) rune {
		if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') || r == '_' || r == '-' || r == '.' {
			return r
		}
		return -1
	}, filename)
}

// CalculateInvoiceTotals calculates subtotal, tax, and grand total for invoice items
func CalculateInvoiceTotals(items []map[string]interface{}) (float64, float64, float64) {
	var subtotal, taxTotal, grandTotal float64
	
	for _, item := range items {
		quantity := toFloat64(item["quantity"])
		unitPrice := toFloat64(item["unit_price"])
		taxRate := toFloat64(item["tax_rate_percentage"]) / 100.0
		
		itemTotal := quantity * unitPrice
		itemTax := itemTotal * taxRate
		
		subtotal += itemTotal
		taxTotal += itemTax
	}
	
	grandTotal = subtotal + taxTotal
	
	return subtotal, taxTotal, grandTotal
}

// toFloat64 safely converts an interface{} to float64
func toFloat64(v interface{}) float64 {
	switch val := v.(type) {
	case float64:
		return val
	case float32:
		return float64(val)
	case int:
		return float64(val)
	case int64:
		return float64(val)
	case int32:
		return float64(val)
	case string:
		if f, err := strconv.ParseFloat(val, 64); err == nil {
			return f
		}
	default:
		// unsupported type or nil, default to zero
	}
	return 0
}
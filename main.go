// main.go file
// main.go
package main

import (
	"flag"
	"invoice-go/database"
	"invoice-go/routes"
	"log"
	"os"
)

func main() {
	seedDb := flag.Bool("seed", false, "Seed the database with initial data")
	flag.Parse()
	// Create uploads directory if it doesn't exist
	if _, err := os.Stat("uploads"); os.IsNotExist(err) {
		os.Mkdir("uploads", 0755)
	}

	// Initialize database
	db := database.InitDB()


	if *seedDb {
		log.Println("Seeding database...")
		if err := database.SeedDatabase(db); err != nil {
			log.Fatalf("Failed to seed database: %v", err)
		}
		log.Println("Database seeding completed")
	}

	// Create uploads directory if it doesn't exist
	if err := os.MkdirAll("./uploads/products", os.ModePerm); err != nil {
		log.Fatalf("Failed to create uploads directory: %v", err)
	}

	// Setup router
	r := routes.SetupRouter(db)

	// Start server
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	
	log.Printf("Server starting on port %s...", port)
	if err := r.Run(":" + port); err != nil {
		log.Fatal("Failed to start server:", err)
	}
}
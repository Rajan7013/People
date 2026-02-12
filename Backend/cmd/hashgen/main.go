package main

import (
	"fmt"
	"os"

	"github.com/rajanprasaila/PeopleOS/backend/peopleos-api/internal/auth"
)

func main() {
	passwords := []string{"admin123", "hr123", "manager123", "employee123"}
	pepper := os.Getenv("PEPPER_SECRET")
	if pepper == "" {
		pepper = "change-me-to-a-long-random-string-in-prod" // Default dev pepper
		fmt.Println("WARNING: Using default dev pepper. Set PEPPER_SECRET env var for production.")
	}

	fmt.Printf("Using Pepper: %s\n\n", pepper)

	for _, password := range passwords {
		hash, err := auth.HashPassword(password, pepper)
		if err != nil {
			fmt.Printf("Error hashing %s: %v\n", password, err)
			continue
		}
		fmt.Printf("Password: %s\nHash: %s\n\n", password, hash)
	}
}

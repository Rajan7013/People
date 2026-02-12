package main

import (
	"fmt"

	"golang.org/x/crypto/bcrypt"
)

func main() {
	hash := "$2a$10$7X3uovR9/.EFXuCLxcNCOOYLoCI6a1.BWEKg9pObV6nRs/ajADJPq"
	password := "admin123"

	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	if err != nil {
		fmt.Printf("Password verification failed: %v\n", err)
	} else {
		fmt.Printf("Password verification successful!\n")
	}
}

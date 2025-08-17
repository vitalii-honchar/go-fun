package main

import (
	"encoding/json"
	"fmt"

	"github.com/go-playground/validator/v10"
)

type User struct {
	FirstName string `json:"first_name" validate:"required"`
	LastName  string `json:"last_name" validate:"required"`
	Email     string `json:"email" validate:"required,email"`
	Age       int    `json:"age" validate:"gte=0,lte=130"`
}

func main() {
	validate := validator.New(validator.WithRequiredStructEnabled())

	// Example JSON data
	jsonData := `{
		"first_name": "John",
		"last_name": "Doe",
		"email": "john.doeexample.com",
		"age": 25
	}`

	// Unmarshal JSON into struct
	var user User
	err := json.Unmarshal([]byte(jsonData), &user)
	if err != nil {
		fmt.Println("JSON unmarshaling error:", err)
		return
	}

	fmt.Printf("Unmarshaled user: %+v\n", user)

	// Validate the struct
	err = validate.Struct(user)
	if err != nil {
		fmt.Println("Validation errors found:", err)
		return
	}

	fmt.Println("User is valid.")
}

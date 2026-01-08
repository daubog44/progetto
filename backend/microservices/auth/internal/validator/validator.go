package validator

import (
	"errors"
	"unicode"

	"github.com/go-playground/validator/v10"
)

var (
	validate *validator.Validate
)

func init() {
	validate = validator.New()
}

// RegistrationRequest helper for validation tags
type RegistrationRequest struct {
	Email    string `validate:"required,email"`
	Password string `validate:"required,min=8"`
	Username string `validate:"required,min=3,max=32"`
}

// ValidateRegister checks email format and password complexity
func ValidateRegister(email, password, username string) error {
	// 1. Structural validation using go-playground/validator
	req := RegistrationRequest{
		Email:    email,
		Password: password,
		Username: username,
	}

	if err := validate.Struct(req); err != nil {
		return err
	}

	// 2. Custom Password Complexity Check
	if !isComplexPassword(password) {
		return errors.New("password must contain at least one uppercase letter, one number, and one special character")
	}

	return nil
}

func isComplexPassword(pass string) bool {
	var (
		hasUpper   bool
		hasNumber  bool
		hasSpecial bool
	)

	for _, char := range pass {
		switch {
		case unicode.IsUpper(char):
			hasUpper = true
		case unicode.IsDigit(char):
			hasNumber = true
		case unicode.IsPunct(char) || unicode.IsSymbol(char):
			hasSpecial = true
		}
	}

	return hasUpper && hasNumber && hasSpecial
}

package auth_service

import (
	"net/mail"
	"strings"
	"unicode"
)

func validateAuthInput(email, password string) error {
	if err := validateEmail(email); err != nil {
		return err
	}
	if err := validatePassword(password); err != nil {
		return err
	}
	return nil
}

func validateEmail(email string) error {
	if email == "" {
		return ErrInvalidEmail
	}

	if len(email) < 5 || len(email) > 254 {
		return ErrInvalidEmail
	}

	_, err := mail.ParseAddress(email)
	if err != nil {
		return ErrInvalidEmail
	}

	parts := strings.Split(email, "@")
	if len(parts) != 2 {
		return ErrInvalidEmail
	}

	if len(parts[1]) == 0 || len(parts[1]) > 253 {
		return ErrInvalidEmail
	}

	if len(parts[1]) < 3 {
		return ErrInvalidEmail
	}

	return nil
}

func validatePassword(password string) error {
	if len(password) < 8 {
		return ErrInvalidPassword
	}

	hasUpper := false
	hasLower := false
	hasDigit := false
	hasSpecial := false

	for _, char := range password {
		switch {
		case unicode.IsUpper(char):
			hasUpper = true
		case unicode.IsLower(char):
			hasLower = true
		case unicode.IsDigit(char):
			hasDigit = true
		case unicode.IsPunct(char) || unicode.IsSymbol(char):
			hasSpecial = true
		}
	}

	if !hasUpper || !hasLower || !hasDigit || !hasSpecial {
		return ErrInvalidPassword
	}

	return nil
}

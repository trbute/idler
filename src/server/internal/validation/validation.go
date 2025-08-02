package validation

import (
	"errors"
	"net/mail"
	"regexp"
	"strings"
	"unicode"
)

var (
	ErrEmailRequired     = errors.New("email is required")
	ErrEmailInvalid      = errors.New("email is invalid")
	ErrPasswordRequired  = errors.New("password is required")
	ErrPasswordTooShort  = errors.New("password must be at least 8 characters")
	ErrPasswordTooLong   = errors.New("password must be at most 72 characters")
	ErrPasswordWeak      = errors.New("password must contain at least one uppercase letter, one lowercase letter, and one number")
	ErrSurnameRequired   = errors.New("surname is required")
	ErrSurnameTooShort   = errors.New("surname must be at least 3 characters")
	ErrSurnameTooLong    = errors.New("surname must be at most 20 characters")
	ErrSurnameInvalid    = errors.New("surname can only contain letters, numbers, and underscores")
	ErrNameRequired      = errors.New("name is required")
	ErrNameTooShort      = errors.New("name must be at least 3 characters")
	ErrNameTooLong       = errors.New("name must be at most 20 characters")
	ErrNameInvalid       = errors.New("name can only contain letters, numbers, and underscores")
	ErrTargetRequired    = errors.New("target is required")
	ErrTargetInvalid     = errors.New("target is invalid")
	ErrAmountInvalid     = errors.New("amount must be greater than 0")
	ErrItemNameRequired  = errors.New("item name is required")
	ErrItemNameInvalid   = errors.New("item name is invalid")
	ErrQuantityRequired  = errors.New("quantity is required")
	ErrQuantityInvalid   = errors.New("quantity must be greater than 0")
)

var (
	nameRegex = regexp.MustCompile(`^[a-zA-Z0-9_]+$`)
	gameItemRegex = regexp.MustCompile(`^[A-Z0-9\s]+$`)
)

func ValidateEmail(email string) error {
	email = strings.TrimSpace(email)
	if email == "" {
		return ErrEmailRequired
	}
	
	_, err := mail.ParseAddress(email)
	if err != nil {
		return ErrEmailInvalid
	}
	
	if !strings.Contains(email, "@") || !strings.Contains(email, ".") {
		return ErrEmailInvalid
	}
	
	return nil
}

func ValidatePassword(password string) error {
	if password == "" {
		return ErrPasswordRequired
	}
	
	if len(password) < 8 {
		return ErrPasswordTooShort
	}
	
	if len(password) > 72 {
		return ErrPasswordTooLong
	}
	
	var hasUpper, hasLower, hasNumber bool
	for _, char := range password {
		switch {
		case unicode.IsUpper(char):
			hasUpper = true
		case unicode.IsLower(char):
			hasLower = true
		case unicode.IsNumber(char):
			hasNumber = true
		}
	}
	
	if !hasUpper || !hasLower || !hasNumber {
		return ErrPasswordWeak
	}
	
	return nil
}

func ValidateSurname(surname string) error {
	surname = strings.TrimSpace(surname)
	if surname == "" {
		return ErrSurnameRequired
	}
	
	if len(surname) < 3 {
		return ErrSurnameTooShort
	}
	
	if len(surname) > 20 {
		return ErrSurnameTooLong
	}
	
	if !nameRegex.MatchString(surname) {
		return ErrSurnameInvalid
	}
	
	return nil
}

func ValidateCharacterName(name string) error {
	name = strings.TrimSpace(name)
	if name == "" {
		return ErrNameRequired
	}
	
	if len(name) < 3 {
		return ErrNameTooShort
	}
	
	if len(name) > 20 {
		return ErrNameTooLong
	}
	
	if !nameRegex.MatchString(name) {
		return ErrNameInvalid
	}
	
	return nil
}

func ValidateTarget(target string) error {
	if target == "" {
		return ErrTargetRequired
	}
	
	if target == "IDLE" {
		return nil
	}
	
	if len(target) > 50 {
		return ErrTargetInvalid
	}
	
	if !gameItemRegex.MatchString(target) {
		return ErrTargetInvalid
	}
	
	return nil
}

func ValidateAmount(amount *int) error {
	if amount != nil && *amount <= 0 {
		return ErrAmountInvalid
	}
	return nil
}

func ValidateItemName(itemName string) error {
	if itemName == "" {
		return ErrItemNameRequired
	}
	
	if len(itemName) > 50 {
		return ErrItemNameInvalid
	}
	
	if !gameItemRegex.MatchString(itemName) {
		return ErrItemNameInvalid
	}
	
	return nil
}

func ValidateQuantity(quantity int) error {
	if quantity <= 0 {
		return ErrQuantityInvalid
	}
	return nil
}

func SanitizeString(s string, maxLength int) string {
	s = strings.TrimSpace(s)
	if len(s) > maxLength {
		s = s[:maxLength]
	}
	return s
}
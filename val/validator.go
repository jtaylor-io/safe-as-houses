package val

import (
	"errors"
	"fmt"
	"net/mail"
	"regexp"
)

var (
	ErrStringTooShort  = errors.New("string is too short")
	ErrStringTooLong   = errors.New("string is too long")
	ErrInvalidUsername = errors.New("username can only contain letters, numbers or underscores")
	ErrInvalidFullName = errors.New("full_name can only contain letters or spaces")
	ErrInvalidEmail    = errors.New("email is invalid")
	isValidUsername    = regexp.MustCompile("^[a-zA-Z0-9_]+$").MatchString
	isValidFullName    = regexp.MustCompile("^[a-zA-Z\\s]+$").MatchString
)

func ValidateString(value string, minLength, maxLength int) error {
	if len(value) < minLength {
		return ErrStringTooShort
	}
	if len(value) > maxLength {
		return ErrStringTooLong
	}
	return nil
}

func ValidateUsername(value string) error {
	if err := ValidateString(value, 3, 100); err != nil {
		return err
	}
	if !isValidUsername(value) {
		return ErrInvalidUsername
	}
	return nil
}

func ValidateFullName(value string) error {
	if err := ValidateString(value, 3, 100); err != nil {
		return err
	}
	if !isValidFullName(value) {
		return ErrInvalidFullName
	}

	return nil
}

func ValidatePassword(value string) error {
	return ValidateString(value, 6, 100)
}

func ValidateEmail(value string) error {
	if err := ValidateString(value, 3, 200); err != nil {
		return err
	}

	if _, err := mail.ParseAddress(value); err != nil {
		return err
	}
	return nil
}

func ValidateEmailId(value int64) error {
	if value <= 0 {
		return fmt.Errorf("email id must be a positive integer")
	}
	return nil
}

func ValidateSecretCode(value string) error {
	return ValidateString(value, 32, 128)
}

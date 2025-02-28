package passwordvalidator

import (
	"fmt"
	"strings"
	"unicode"
	"unicode/utf8"
)

var (
	ErrPasswordIsTooShort                  = fmt.Errorf("password is too short")
	ErrPasswordContainsNoNumbers           = fmt.Errorf("password does not contain a number")
	ErrPasswordContainsNoSpecialCharacters = fmt.Errorf("password does not contain a special character")
)

type PasswordValidator interface {
	Validate() error // Password is an argument
}

type PasswordMinLengthValidator struct {
	minLength int
	password  string
}

func (v PasswordMinLengthValidator) Validate() error {
	if utf8.RuneCountInString(v.password) < v.minLength { // Use utf8.RuneCountInString
		return ErrPasswordIsTooShort
	}
	return nil
}

type PasswordNumberValidator struct {
	password string
}

func (v PasswordNumberValidator) Validate() error {
	for _, r := range v.password {
		if unicode.IsDigit(r) {
			return nil
		}
	}
	return ErrPasswordContainsNoNumbers
}

type PasswordSpecialCharValidator struct {
	password string
}

func (v PasswordSpecialCharValidator) Validate() error {
	specialChars := "!@#$%^&*()"
	for _, r := range v.password {
		if strings.ContainsRune(specialChars, r) {
			return nil
		}
	}
	return ErrPasswordContainsNoSpecialCharacters
}

func NewPasswordMinLengthValidator(password string, minLength int) PasswordMinLengthValidator {
	return PasswordMinLengthValidator{minLength: minLength, password: password}
}

func NewPasswordNumberValidator(password string) PasswordNumberValidator {
	return PasswordNumberValidator{password: password}
}

func NewPasswordSpecialCharValidator(password string) PasswordSpecialCharValidator {
	return PasswordSpecialCharValidator{password: password}
}

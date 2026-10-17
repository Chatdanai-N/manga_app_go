package validatorutil

import (
	"regexp"
	"unicode"

	"github.com/go-playground/validator/v10"
)

var IsValidEmail validator.Func = func(fl validator.FieldLevel) bool {
	emailRegex := regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)
	email := fl.Field().String()
	return emailRegex.MatchString(email)
}

var LenPassword validator.Func = func(fl validator.FieldLevel) bool {
	password := fl.Field().String()
	return len(password) > 8
}
var HasUpper validator.Func = func(fl validator.FieldLevel) bool {
	password := fl.Field().String()
	hasUpper := false
	for _, char := range password {
		if unicode.IsUpper(char) {
			hasUpper = true
		}
	}
	return hasUpper
}
var HasLower validator.Func = func(fl validator.FieldLevel) bool {
	password := fl.Field().String()
	hasLower := false
	for _, char := range password {
		if unicode.IsLower(char) {
			hasLower = true
		}
	}
	return hasLower
}

var HasNumber validator.Func = func(fl validator.FieldLevel) bool {
	password := fl.Field().String()
	hasNumber := false
	for _, char := range password {
		if unicode.IsNumber(char) {
			hasNumber = true
		}
	}
	return hasNumber
}

var HasSpacial validator.Func = func(fl validator.FieldLevel) bool {
	password := fl.Field().String()
	hasSpacial := false
	for _, char := range password {
		if unicode.IsPunct(char) || unicode.IsSymbol(char) {
			hasSpacial = true
		}
	}
	return hasSpacial
}

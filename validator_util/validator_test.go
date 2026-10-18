package validator_util

import (
	"testing"

	"github.com/go-playground/validator/v10"
)

type TestSuite struct {
	name           string
	input          string
	expectedResult bool
}

func TestIsValidEmail(t *testing.T) {

	ts := []TestSuite{
		{
			name:           "valid email",
			input:          "test@example.com",
			expectedResult: true,
		},
		{
			name:           "invalid email",
			input:          "ser.name+tag@sub.domain.co.uk",
			expectedResult: false,
		}, {
			name:           "invalid email missing domain",
			input:          "test@example",
			expectedResult: false,
		},
	}

	type TestUserEmail struct {
		Email string `validate:"is_valid_email"`
	}

	runValidatorTests(t, ts, "is_valid_email", IsValidEmail, func(val string) any {
		return TestUserEmail{
			Email: val,
		}
	})
}

func TestLenPassword(t *testing.T) {
	ts := []TestSuite{
		{
			name:           "valid length character",
			input:          "123456789",
			expectedResult: true,
		},
		{
			name:           "invalid length character",
			input:          "1234567",
			expectedResult: false,
		},
	}

	type TestUserLenPassword struct {
		Password string `validate:"len_password"`
	}

	runValidatorTests(t, ts, "len_password", LenPassword, func(val string) any {
		return TestUserLenPassword{
			Password: val,
		}
	})
}

func TestHasUpper(t *testing.T) {
	ts := []TestSuite{
		{
			name:           "valid upper character",
			input:          "P@ssword123",
			expectedResult: true,
		},
		{
			name:           "invalid upper character",
			input:          "p@ssword123",
			expectedResult: false,
		},
	}

	type TestUserHasUpper struct {
		Password string `validate:"has_upper"`
	}

	runValidatorTests(t, ts, "has_upper", HasUpper, func(val string) any {
		return TestUserHasUpper{
			Password: val,
		}
	})
}

func TestHasLower(t *testing.T) {
	ts := []TestSuite{
		{
			name:           "valid lower character",
			input:          "P@ssword123",
			expectedResult: true,
		}, {
			name:           "invalid lower character",
			input:          "P@SSW0RD123",
			expectedResult: false,
		},
	}

	type TestUserHasLower struct {
		Password string `validate:"has_lower"`
	}

	runValidatorTests(t, ts, "has_lower", HasLower, func(val string) any {
		return TestUserHasLower{
			Password: val,
		}
	})
}

func TestHasNumber(t *testing.T) {
	ts := []TestSuite{
		{
			name:           "valid number character",
			input:          "P@ssword123",
			expectedResult: true,
		},
		{
			name:           "invalid number character",
			input:          "P@SSWORDABC",
			expectedResult: false,
		},
	}

	type TestUserHasNumber struct {
		Password string `validate:"has_number"`
	}

	runValidatorTests(t, ts, "has_number", HasNumber, func(val string) any {
		return TestUserHasNumber{
			Password: val,
		}
	})
}

func TestHasSpacial(t *testing.T) {
	ts := []TestSuite{
		{
			name:           "valid symbol character",
			input:          "P@ssword123",
			expectedResult: true,
		}, {
			name:           "invalid symbol character",
			input:          "Passw0rd123",
			expectedResult: false,
		},
	}

	type TestUserHasSpacial struct {
		Password string `validate:"has_spacial"`
	}

	runValidatorTests(t, ts, "has_spacial", HasSpacial, func(val string) any {
		return TestUserHasSpacial{
			Password: val,
		}
	})
}

func runValidatorTests(t *testing.T, ts []TestSuite, tag string, validatorFunc validator.Func, createStruct func(val string) interface{}) {
	t.Helper() // Helper Function เพื่อให้ Stack trace แสดงตำแหน่งที่ถูกต้อง

	validate := validator.New()
	err := validate.RegisterValidation(tag, validatorFunc)
	if err != nil {
		t.Errorf("Failed to register HasUpper validator: %v", err)
	}

	for _, tc := range ts {
		t.Run(tc.name, func(t *testing.T) {
			user := createStruct(tc.input)
			isValid := validate.Struct(user) == nil
			if isValid != tc.expectedResult {
				t.Errorf("testsuite name = %v isValid = %v, expectedResult = %v", tc.name, isValid, tc.expectedResult)
			}
		})
	}
}

package validation

import (
	"testing"

	"github.com/go-playground/validator/v10"
	"github.com/stretchr/testify/assert"
)

type TestStruct struct {
	Email    string `validate:"email_advanced"`
	Password string `validate:"password_strong"`
	Slug     string `validate:"slug"`
	Search   string `validate:"search"`
	MinInt   int    `validate:"min_int=10"`
	MaxInt   int    `validate:"max_int=20"`
	File     string `validate:"file_ext=jpg png"`
}

func TestValidationAndErrorHandling(t *testing.T) {
	v := validator.New()
	RegisterCustomValidation(v)
	tests := []struct {
		name          string
		input         TestStruct
		expectErr     bool
		expectedField string
		expectedMsg   string
	}{
		{
			name:      "valid",
			input:     TestStruct{Email: "test@gmail.com", Password: "PassWord@123", Slug: "ok-slug", Search: "ok", MinInt: 15, MaxInt: 15, File: "test.jpg"},
			expectErr: false,
		},
		{
			name:          "blocked",
			input:         TestStruct{Email: "test@blacklist.com", Password: "PassWord@123", Slug: "s", Search: "s", MinInt: 15, MaxInt: 15, File: "test.jpg"},
			expectErr:     true,
			expectedField: "email",
			expectedMsg:   "email is not allowed (blacklisted)",
		},
		{
			name:          "short password",
			input:         TestStruct{Email: "test@gmail.com", Password: "weak", Slug: "s", Search: "s", MinInt: 15, MaxInt: 15, File: "test.jpg"},
			expectErr:     true,
			expectedField: "password",
			expectedMsg:   "password must be at least 8 characters long and contain lowercase, uppercase, numbers, and special characters",
		},
		{
			name:          "file ext",
			input:         TestStruct{Email: "test@gmail.com", Password: "PassWord@123", Slug: "s", Search: "s", MinInt: 15, MaxInt: 15, File: "test.exe"},
			expectErr:     true,
			expectedField: "file",
			expectedMsg:   "file must be one of the following extensions: jpg, png",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			err := v.Struct(tc.input)
			if tc.expectErr {
				assert.Error(t, err)
				response := HandleValidationErrors(err)
				errMap, ok := response["error"].(map[string]string)
				assert.True(t, ok, "Result should be a map")
				actualMsg, exists := errMap[tc.expectedField]
				assert.True(t, exists, "Expected field %s not found in error map", tc.expectedField)
				assert.Equal(t, tc.expectedMsg, actualMsg)

			} else {
				assert.NoError(t, err)
			}
		})
	}
}
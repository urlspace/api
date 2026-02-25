package validator_test

import (
	"strings"
	"testing"

	"github.com/hreftools/api/internal/validator"
)

func TestEmail(t *testing.T) {
	tests := []struct {
		name       string
		input      string
		wantErr    bool
		wantErrMsg string
	}{
		{
			name:    "Email is valid",
			input:   "example@example.com",
			wantErr: false,
		},
		{
			name:       "Email is missing",
			input:      "",
			wantErr:    true,
			wantErrMsg: "email is required",
		},
		{
			name:       "Email is invalid",
			input:      "invalidemail.com",
			wantErr:    true,
			wantErrMsg: "email format is invalid",
		},
		{
			name:       "Email's length exceeds 254 characters",
			input:      strings.Repeat("a", 245) + "@email.com",
			wantErr:    true,
			wantErrMsg: "email must be at most 254 characters",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotErr := validator.Email(tt.input)
			if gotErr != nil {
				if !tt.wantErr {
					t.Errorf("Email() failed: %v", gotErr)
				}
				if gotErr.Error() != tt.wantErrMsg {
					t.Errorf("Email() error message = %v, want %v", gotErr.Error(), tt.wantErrMsg)
				}
				return
			}
			if tt.wantErr {
				t.Fatal("Email() succeeded unexpectedly")
			}
		})
	}
}

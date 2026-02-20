package validator_test

import (
	"strings"
	"testing"

	"github.com/hreftools/api/internal/validator"
)

func TestEmail(t *testing.T) {
	tests := []struct {
		name       string
		e          string
		wantErr    bool
		wantErrMsg string
	}{
		{
			name:       "Email is missing",
			e:          "",
			wantErr:    true,
			wantErrMsg: "email is required",
		},
		{
			name:       "Email is invalid",
			e:          "invalidemail.com",
			wantErr:    true,
			wantErrMsg: "email format is invalid",
		},
		{
			name:       "Email's length exceeds 254 characters",
			e:          strings.Repeat("a", 245) + "@email.com",
			wantErr:    true,
			wantErrMsg: "email must be at most 254 characters",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotErr := validator.Email(tt.e)
			if gotErr != nil {
				if !tt.wantErr {
					t.Errorf("Email() failed: %v", gotErr)
				}
				return
			}
			if tt.wantErr {
				t.Fatal("Email() succeeded unexpectedly")
			}
		})
	}
}

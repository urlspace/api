package validator_test

import (
	"testing"

	"github.com/hreftools/api/internal/validator"
)

func TestPassword(t *testing.T) {
	tests := []struct {
		name       string
		input      string
		wantErr    bool
		wantErrMsg string
	}{
		{
			name:    "Valid password",
			input:   "supersecretpassword123",
			wantErr: false,
		},
		{
			name:       "Missing password",
			input:      "",
			wantErr:    true,
			wantErrMsg: "password is required",
		},
		{
			name:       "Password too short",
			input:      "password",
			wantErr:    true,
			wantErrMsg: "password must be at least 12 characters",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotErr := validator.Password(tt.input)
			if gotErr != nil {
				if !tt.wantErr {
					t.Errorf("Password() failed: %v", gotErr)
				}
				if gotErr.Error() != tt.wantErrMsg {
					t.Errorf("Password() error message = %v, want %v", gotErr.Error(), tt.wantErrMsg)
				}
				return
			}
			if tt.wantErr {
				t.Fatal("Password() succeeded unexpectedly")
			}
		})
	}
}

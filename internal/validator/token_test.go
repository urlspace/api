package validator_test

import (
	"testing"

	"github.com/google/uuid"
	"github.com/hreftools/api/internal/validator"
)

func TestToken(t *testing.T) {
	tests := []struct {
		name       string
		input      string
		wantErr    bool
		wantErrMsg string
	}{
		{
			name:    "Valid token",
			input:   uuid.New().String(),
			wantErr: false,
		},
		{
			name:       "Empty Token",
			input:      "",
			wantErr:    true,
			wantErrMsg: "token is required",
		},
		{
			name:       "Invalid Token (not a UUID)",
			input:      "not-a-uuid",
			wantErr:    true,
			wantErrMsg: "token is invalid",
		},
		{
			name:       "Invalid Token (incorrect format, last segment is missing)",
			input:      "12345678-1234-1234-1234",
			wantErr:    true,
			wantErrMsg: "token is invalid",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotErr := validator.Token(tt.input)
			if gotErr != nil {
				if !tt.wantErr {
					t.Errorf("Token() failed: %v", gotErr)
				}
				if gotErr.Error() != tt.wantErrMsg {
					t.Errorf("Token() error message = %v, want %v", gotErr.Error(), tt.wantErrMsg)
				}
				return
			}
			if tt.wantErr {
				t.Fatal("Token() succeeded unexpectedly")
			}
		})
	}
}

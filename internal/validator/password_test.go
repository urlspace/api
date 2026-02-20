package validator_test

import (
	"github.com/hreftools/api/internal/validator"
	"testing"
)

func TestPassword(t *testing.T) {
	tests := []struct {
		name       string
		p          string
		wantErr    bool
		wantErrMsg string
	}{
		{
			name:       "Missing password",
			p:          "",
			wantErr:    true,
			wantErrMsg: "password is required",
		},
		{
			name:       "Password too short",
			p:          "password",
			wantErr:    true,
			wantErrMsg: "password must be at least 12 characters",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotErr := validator.Password(tt.p)
			if gotErr != nil {
				if !tt.wantErr {
					t.Errorf("Password() failed: %v", gotErr)
				}
				return
			}
			if tt.wantErr {
				t.Fatal("Password() succeeded unexpectedly")
			}
		})
	}
}

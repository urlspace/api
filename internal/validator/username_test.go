package validator_test

import (
	"testing"

	"github.com/hreftools/api/internal/validator"
)

func TestUsername(t *testing.T) {
	tests := []struct {
		name       string
		input      string
		wantErr    bool
		wantErrMsg string
	}{
		{
			name:    "Username is valid",
			input:   "username",
			wantErr: false,
		},
		{
			name:       "Missing username",
			input:      "",
			wantErr:    true,
			wantErrMsg: "username is required",
		},
		{
			name:       "Username too short",
			input:      "ab",
			wantErr:    true,
			wantErrMsg: "username must be min 3 characters",
		},
		{
			name:       "Username too long",
			input:      "thisusernameiswaytoolongtobevalid",
			wantErr:    true,
			wantErrMsg: "username must be max 32 characters",
		},
		{
			name:       "Username contains invalid character",
			input:      "invalid_%_user",
			wantErr:    true,
			wantErrMsg: "username can only contain lowercase characters, numbers, hyphens, and underscores",
		},
		{
			name:       "Username starts with hyphen",
			input:      "-invalid_user",
			wantErr:    true,
			wantErrMsg: "username cannot start with hyphen or underscore",
		},
		{
			name:       "Username ends with hyphen",
			input:      "invalid_user-",
			wantErr:    true,
			wantErrMsg: "username cannot end with hyphen or underscore",
		},
		{
			name:       "Username starts with underscore",
			input:      "_invalid_user",
			wantErr:    true,
			wantErrMsg: "username cannot start with hyphen or underscore",
		},
		{
			name:       "Username ends with underscore",
			input:      "invalid_user_",
			wantErr:    true,
			wantErrMsg: "username cannot end with hyphen or underscore",
		},
		{
			name:       "Username is reserved",
			input:      "admin",
			wantErr:    true,
			wantErrMsg: "username is reserved",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotErr := validator.Username(tt.input)
			if gotErr != nil {
				if !tt.wantErr {
					t.Errorf("Username() failed: %v", gotErr)
				}
				if gotErr.Error() != tt.wantErrMsg {
					t.Errorf("Username() error message = %v, want %v", gotErr.Error(), tt.wantErrMsg)
				}
				return
			}
			if tt.wantErr {
				t.Fatal("Username() succeeded unexpectedly")
			}
		})
	}
}

package validator_test

import (
	"github.com/hreftools/api/internal/validator"
	"testing"
)

func TestUsername(t *testing.T) {
	tests := []struct {
		name       string
		u          string
		wantErr    bool
		wantErrMsg string
	}{
		{
			name:       "Missing username",
			u:          "",
			wantErr:    true,
			wantErrMsg: "username is required",
		},
		{
			name:       "Username too short",
			u:          "ab",
			wantErr:    true,
			wantErrMsg: "username must be min 3 characters",
		},
		{
			name:       "Username too long",
			u:          "thisusernameiswaytoolongtobevalid",
			wantErr:    true,
			wantErrMsg: "username must be max 32 characters",
		},
		{
			name:       "Username contains invalid character",
			u:          "invalid_%_user",
			wantErr:    true,
			wantErrMsg: "username can only contain lowercase characters, numbers, hyphens, and underscores",
		},
		{
			name:       "Username starts with hyphen",
			u:          "-invalid_user",
			wantErr:    true,
			wantErrMsg: "username cannot start with hyphen or underscore",
		},
		{
			name:       "Username ends with hyphen",
			u:          "invalid_user-",
			wantErr:    true,
			wantErrMsg: "username cannot end with hyphen or underscore",
		},
		{
			name:       "Username starts with underscore",
			u:          "_invalid_user",
			wantErr:    true,
			wantErrMsg: "username cannot start with hyphen or underscore",
		},
		{
			name:       "Username ends with underscore",
			u:          "invalid_user_",
			wantErr:    true,
			wantErrMsg: "username cannot end with hyphen or underscore",
		},
		{
			name:       "Username is reserved",
			u:          "admin",
			wantErr:    true,
			wantErrMsg: "username is reserved",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotErr := validator.Username(tt.u)
			if gotErr != nil {
				if !tt.wantErr {
					t.Errorf("Username() failed: %v", gotErr)
				}
				return
			}
			if tt.wantErr {
				t.Fatal("Username() succeeded unexpectedly")
			}
		})
	}
}

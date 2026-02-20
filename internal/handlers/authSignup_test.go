package handlers_test

import (
	"testing"

	"github.com/hreftools/api/internal/handlers"
)

func TestAuthSignupBody_Normalize(t *testing.T) {
	tests := []struct {
		name     string
		input    handlers.AuthSignupBody
		expected handlers.AuthSignupBody
	}{
		{
			name: "Produces no change to already normalized input",
			input: handlers.AuthSignupBody{
				Username: "user_name",
				Email:    "user@email.com",
				Password: "  whateva  ",
			},
			expected: handlers.AuthSignupBody{
				Username: "user_name",
				Email:    "user@email.com",
				Password: "  whateva  ",
			},
		},
		{
			name: "Trim and lowercase username and email",
			input: handlers.AuthSignupBody{
				Username: "  User_Name  ",
				Email:    "  User@Email.com  ",
				Password: "  whateva  ",
			},
			expected: handlers.AuthSignupBody{
				Username: "user_name",
				Email:    "user@email.com",
				Password: "  whateva  ",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			b := tt.input
			b.Normalize()

			if b.Username != tt.expected.Username {
				t.Errorf("Normalize() Username = %v, want %v", b.Username, tt.expected.Username)
			}
			if b.Email != tt.expected.Email {
				t.Errorf("Normalize() Email = %v, want %v", b.Email, tt.expected.Email)
			}
			if b.Password != tt.expected.Password {
				t.Errorf("Normalize() Password = %v, want %v", b.Password, tt.expected.Password)
			}
		})
	}
}

func TestAuthSignupBody_Validate(t *testing.T) {
	tests := []struct {
		name       string
		input      handlers.AuthSignupBody
		wantErr    bool
		wantErrMsg string
	}{
		{
			name: "Valid input",
			input: handlers.AuthSignupBody{
				Username: "valid_user",
				Email:    "valid@email.com",
				Password: "strongpassword",
			},
			wantErr:    false,
			wantErrMsg: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			b := tt.input
			gotErr := b.Validate()

			if gotErr != nil {
				if !tt.wantErr {
					t.Errorf("Validate() failed: %v", gotErr)
				}
				if gotErr.Error() != tt.wantErrMsg {
					t.Errorf("Validate() error = %q, want %q", gotErr.Error(), tt.wantErrMsg)
				}
				return
			}
			if tt.wantErr {
				t.Fatal("Validate() succeeded unexpectedly")
			}
		})
	}
}

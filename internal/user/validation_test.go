package user

import (
	"strings"
	"testing"

	"github.com/google/uuid"
)

func Test_validateEmail(t *testing.T) {
	tests := []struct {
		name       string
		input      string
		wantResult string
		wantErr    bool
		wantErrMsg string
	}{
		{
			name:       "Email is valid",
			input:      "example@example.com",
			wantResult: "example@example.com",
			wantErr:    false,
		},
		{
			name:       "Email is normalized to lowercase and trimmed",
			input:      "  Example@Example.COM  ",
			wantResult: "example@example.com",
			wantErr:    false,
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
			name:       "Email with RFC 5322 display name is rejected",
			input:      `"Joe" <joe@evil.com>`,
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
			gotResult, gotErr := validateEmail(tt.input)
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
			if tt.wantResult != "" && gotResult != tt.wantResult {
				t.Errorf("Email() result = %v, want %v", gotResult, tt.wantResult)
			}
		})
	}
}

func Test_validatePassword(t *testing.T) {
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
		{
			name:       "Multi-byte characters are counted as characters not bytes",
			input:      "🔒🔒🔒",
			wantErr:    true,
			wantErrMsg: "password must be at least 12 characters",
		},
		{
			name:       "Whitespace-only password is rejected",
			input:      "            ",
			wantErr:    true,
			wantErrMsg: "password is required",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, gotErr := validatePassword(tt.input)
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

func Test_validateToken(t *testing.T) {
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
			_, gotErr := validateToken(tt.input)
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

func Test_validateUsername(t *testing.T) {
	tests := []struct {
		name       string
		input      string
		wantResult string
		wantErr    bool
		wantErrMsg string
	}{
		{
			name:       "Username is valid",
			input:      "username",
			wantResult: "username",
			wantErr:    false,
		},
		{
			name:       "Username is normalized to lowercase and trimmed",
			input:      "  UserName  ",
			wantResult: "username",
			wantErr:    false,
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
		{
			name:       "Unicode lookalike characters are rejected",
			input:      "аdmin",
			wantErr:    true,
			wantErrMsg: "username can only contain lowercase characters, numbers, hyphens, and underscores",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotResult, gotErr := validateUsername(tt.input)
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
			if tt.wantResult != "" && gotResult != tt.wantResult {
				t.Errorf("Username() result = %v, want %v", gotResult, tt.wantResult)
			}
		})
	}
}

func Test_validateIsAdmin(t *testing.T) {
	boolTrue := true
	boolFalse := false

	tests := []struct {
		name       string
		input      *bool
		want       bool
		wantErr    bool
		wantErrMsg string
	}{
		{
			name:    "Valid isAdmin true",
			input:   &boolTrue,
			want:    true,
			wantErr: false,
		},
		{
			name:    "Valid isAdmin false",
			input:   &boolFalse,
			want:    false,
			wantErr: false,
		},
		{
			name:       "Missing isAdmin",
			input:      nil,
			wantErr:    true,
			wantErrMsg: "isAdmin flag is required",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, gotErr := validateIsAdmin(tt.input)
			if gotErr != nil {
				if !tt.wantErr {
					t.Errorf("validateIsAdmin() failed: %v", gotErr)
				}
				if gotErr.Error() != tt.wantErrMsg {
					t.Errorf("validateIsAdmin() error message = %v, want %v", gotErr.Error(), tt.wantErrMsg)
				}
				return
			}
			if tt.wantErr {
				t.Fatal("validateIsAdmin() succeeded unexpectedly")
			}
			if got != tt.want {
				t.Errorf("validateIsAdmin() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_validateIsPro(t *testing.T) {
	boolTrue := true
	boolFalse := false

	tests := []struct {
		name       string
		input      *bool
		want       bool
		wantErr    bool
		wantErrMsg string
	}{
		{
			name:    "Valid isPro true",
			input:   &boolTrue,
			want:    true,
			wantErr: false,
		},
		{
			name:    "Valid isPro false",
			input:   &boolFalse,
			want:    false,
			wantErr: false,
		},
		{
			name:       "Missing isPro",
			input:      nil,
			wantErr:    true,
			wantErrMsg: "isPro flag is required",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, gotErr := validateIsPro(tt.input)
			if gotErr != nil {
				if !tt.wantErr {
					t.Errorf("validateIsPro() failed: %v", gotErr)
				}
				if gotErr.Error() != tt.wantErrMsg {
					t.Errorf("validateIsPro() error message = %v, want %v", gotErr.Error(), tt.wantErrMsg)
				}
				return
			}
			if tt.wantErr {
				t.Fatal("validateIsPro() succeeded unexpectedly")
			}
			if got != tt.want {
				t.Errorf("validateIsPro() = %v, want %v", got, tt.want)
			}
		})
	}
}

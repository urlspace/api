package collection

import (
	"strings"
	"testing"
)

func Test_ValidateName(t *testing.T) {
	tests := []struct {
		name       string
		input      string
		wantResult string
		wantErr    bool
		wantErrMsg string
	}{
		{
			name:       "Valid name",
			input:      "My Collection",
			wantResult: "My Collection",
			wantErr:    false,
		},
		{
			name:       "Two character name is valid",
			input:      "AI",
			wantResult: "AI",
			wantErr:    false,
		},
		{
			name:       "Name is trimmed",
			input:      "  My Collection  ",
			wantResult: "My Collection",
			wantErr:    false,
		},
		{
			name:       "Internal whitespace is collapsed",
			input:      "Reading   list",
			wantResult: "Reading list",
			wantErr:    false,
		},
		{
			name:       "Non-breaking spaces are collapsed",
			input:      "Reading\u00a0\u00a0list",
			wantResult: "Reading list",
			wantErr:    false,
		},
		{
			name:       "Name is normalized to NFC",
			input:      "Cafe\u0301",
			wantResult: "Caf\u00e9",
			wantErr:    false,
		},
		{
			name:       "Name is too short",
			input:      "a",
			wantErr:    true,
			wantErrMsg: "name must be between 2 and 128 characters",
		},
		{
			name:       "Name is too long",
			input:      strings.Repeat("a", 129),
			wantErr:    true,
			wantErrMsg: "name must be between 2 and 128 characters",
		},
		{
			name:       "Name with null byte is rejected",
			input:      "My \x00 Collection",
			wantErr:    true,
			wantErrMsg: "name must not contain control characters",
		},
		{
			name:       "Name with tab is rejected",
			input:      "My \t Collection",
			wantErr:    true,
			wantErrMsg: "name must not contain control characters",
		},
		{
			name:       "Multi-byte characters are counted as characters not bytes",
			input:      strings.Repeat("ą", 128),
			wantResult: strings.Repeat("ą", 128),
			wantErr:    false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotResult, gotErr := ValidateName(tt.input)
			if gotErr != nil {
				if !tt.wantErr {
					t.Errorf("ValidateName() failed: %v", gotErr)
				}
				if gotErr.Error() != tt.wantErrMsg {
					t.Errorf("ValidateName() error message = %v, want %v", gotErr.Error(), tt.wantErrMsg)
				}
				return
			}
			if tt.wantErr {
				t.Fatal("ValidateName() succeeded unexpectedly")
			}
			if tt.wantResult != "" && gotResult != tt.wantResult {
				t.Errorf("ValidateName() result = %v, want %v", gotResult, tt.wantResult)
			}
		})
	}
}

func Test_ValidateDescription(t *testing.T) {
	tests := []struct {
		name       string
		input      string
		wantResult string
		wantErr    bool
		wantErrMsg string
	}{
		{
			name:       "Valid description",
			input:      "A helpful collection description.",
			wantResult: "A helpful collection description.",
			wantErr:    false,
		},
		{
			name:    "Empty description",
			input:   "",
			wantErr: false,
		},
		{
			name:       "Description is trimmed",
			input:      "  A helpful collection description.  ",
			wantResult: "A helpful collection description.",
			wantErr:    false,
		},
		{
			name:       "Description is too long",
			input:      strings.Repeat("a", 1025),
			wantErr:    true,
			wantErrMsg: "description must be at most 1024 characters",
		},
		{
			name:       "Description with null byte is rejected",
			input:      "A description with \x00 null byte",
			wantErr:    true,
			wantErrMsg: "description must not contain control characters",
		},
		{
			name:       "Description with newline is rejected",
			input:      "A description with \n newline",
			wantErr:    true,
			wantErrMsg: "description must not contain control characters",
		},
		{
			name:       "Multi-byte characters are counted as characters not bytes",
			input:      strings.Repeat("ą", 513),
			wantResult: strings.Repeat("ą", 513),
			wantErr:    false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotResult, gotErr := ValidateDescription(tt.input)
			if gotErr != nil {
				if !tt.wantErr {
					t.Errorf("ValidateDescription() failed: %v", gotErr)
				}
				if gotErr.Error() != tt.wantErrMsg {
					t.Errorf("ValidateDescription() error message = %v, want %v", gotErr.Error(), tt.wantErrMsg)
				}
				return
			}
			if tt.wantErr {
				t.Fatal("ValidateDescription() succeeded unexpectedly")
			}
			if tt.wantResult != "" && gotResult != tt.wantResult {
				t.Errorf("ValidateDescription() result = %v, want %v", gotResult, tt.wantResult)
			}
		})
	}
}

func Test_ValidateSlug(t *testing.T) {
	tests := []struct {
		name       string
		input      string
		wantResult string
		wantErr    bool
		wantErrMsg string
	}{
		{
			name:       "Valid slug",
			input:      "reading-list",
			wantResult: "reading-list",
			wantErr:    false,
		},
		{
			name:       "Slug is trimmed and lowercased",
			input:      "  Reading-List  ",
			wantResult: "reading-list",
			wantErr:    false,
		},
		{
			name:       "Slug is too short",
			input:      "a",
			wantErr:    true,
			wantErrMsg: "slug must be between 2 and 128 characters",
		},
		{
			name:       "Slug is too long",
			input:      strings.Repeat("a", 129),
			wantErr:    true,
			wantErrMsg: "slug must be between 2 and 128 characters",
		},
		{
			name:       "Slug with space is rejected",
			input:      "reading list",
			wantErr:    true,
			wantErrMsg: "slug must contain only lowercase letters, digits and single hyphens",
		},
		{
			name:       "Slug with double hyphen is rejected",
			input:      "reading--list",
			wantErr:    true,
			wantErrMsg: "slug must contain only lowercase letters, digits and single hyphens",
		},
		{
			name:       "Slug with leading hyphen is rejected",
			input:      "-reading",
			wantErr:    true,
			wantErrMsg: "slug must contain only lowercase letters, digits and single hyphens",
		},
		{
			name:       "Slug with accented letter is rejected",
			input:      "café",
			wantErr:    true,
			wantErrMsg: "slug must contain only lowercase letters, digits and single hyphens",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotResult, gotErr := ValidateSlug(tt.input)
			if gotErr != nil {
				if !tt.wantErr {
					t.Errorf("ValidateSlug() failed: %v", gotErr)
				}
				if gotErr.Error() != tt.wantErrMsg {
					t.Errorf("ValidateSlug() error message = %v, want %v", gotErr.Error(), tt.wantErrMsg)
				}
				return
			}
			if tt.wantErr {
				t.Fatal("ValidateSlug() succeeded unexpectedly")
			}
			if gotResult != tt.wantResult {
				t.Errorf("ValidateSlug() result = %v, want %v", gotResult, tt.wantResult)
			}
		})
	}
}

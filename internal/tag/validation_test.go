package tag

import (
	"strings"
	"testing"
)

func Test_validateName(t *testing.T) {
	tests := []struct {
		name       string
		input      string
		wantResult string
		wantErr    bool
		wantErrMsg string
	}{
		{
			name:       "Valid tag name",
			input:      "javascript",
			wantResult: "javascript",
			wantErr:    false,
		},
		{
			name:       "Valid tag name with hyphen",
			input:      "react-native",
			wantResult: "react-native",
			wantErr:    false,
		},
		{
			name:       "Valid tag name with digits",
			input:      "es2024",
			wantResult: "es2024",
			wantErr:    false,
		},
		{
			name:       "Name is trimmed",
			input:      "  javascript  ",
			wantResult: "javascript",
			wantErr:    false,
		},
		{
			name:       "Name is lowercased",
			input:      "JavaScript",
			wantResult: "javascript",
			wantErr:    false,
		},
		{
			name:       "Mixed case is lowercased",
			input:      "TeStInG",
			wantResult: "testing",
			wantErr:    false,
		},
		{
			name:       "Minimum length (2 chars)",
			input:      "go",
			wantResult: "go",
			wantErr:    false,
		},
		{
			name:       "Maximum length (50 chars)",
			input:      strings.Repeat("a", 50),
			wantResult: strings.Repeat("a", 50),
			wantErr:    false,
		},
		{
			name:       "Too short (1 char)",
			input:      "a",
			wantErr:    true,
			wantErrMsg: "tag name must be between 2 and 50 characters",
		},
		{
			name:       "Empty string",
			input:      "",
			wantErr:    true,
			wantErrMsg: "tag name must be between 2 and 50 characters",
		},
		{
			name:       "Too long (51 chars)",
			input:      strings.Repeat("a", 51),
			wantErr:    true,
			wantErrMsg: "tag name must be between 2 and 50 characters",
		},
		{
			name:       "Spaces are rejected",
			input:      "react native",
			wantErr:    true,
			wantErrMsg: "tag name must contain only lowercase letters, digits, and hyphens",
		},
		{
			name:       "Underscores are rejected",
			input:      "react_native",
			wantErr:    true,
			wantErrMsg: "tag name must contain only lowercase letters, digits, and hyphens",
		},
		{
			name:       "Special characters are rejected",
			input:      "c++",
			wantErr:    true,
			wantErrMsg: "tag name must contain only lowercase letters, digits, and hyphens",
		},
		{
			name:       "Leading hyphen is rejected",
			input:      "-javascript",
			wantErr:    true,
			wantErrMsg: "tag name must not start, end with, or contain consecutive hyphens",
		},
		{
			name:       "Trailing hyphen is rejected",
			input:      "javascript-",
			wantErr:    true,
			wantErrMsg: "tag name must not start, end with, or contain consecutive hyphens",
		},
		{
			name:       "Consecutive hyphens are rejected",
			input:      "react--native",
			wantErr:    true,
			wantErrMsg: "tag name must not start, end with, or contain consecutive hyphens",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotResult, gotErr := validateName(tt.input)
			if gotErr != nil {
				if !tt.wantErr {
					t.Errorf("validateName() failed: %v", gotErr)
				}
				if gotErr.Error() != tt.wantErrMsg {
					t.Errorf("validateName() error message = %v, want %v", gotErr.Error(), tt.wantErrMsg)
				}
				return
			}
			if tt.wantErr {
				t.Fatal("validateName() succeeded unexpectedly")
			}
			if gotResult != tt.wantResult {
				t.Errorf("validateName() result = %v, want %v", gotResult, tt.wantResult)
			}
		})
	}
}

func Test_ValidateTagNames(t *testing.T) {
	tests := []struct {
		name       string
		input      []string
		wantResult []string
		wantErr    bool
		wantErrMsg string
	}{
		{
			name:       "Valid list",
			input:      []string{"javascript", "frontend"},
			wantResult: []string{"javascript", "frontend"},
			wantErr:    false,
		},
		{
			name:       "Empty list",
			input:      []string{},
			wantResult: []string{},
			wantErr:    false,
		},
		{
			name:       "Duplicates are removed",
			input:      []string{"javascript", "javascript", "frontend"},
			wantResult: []string{"javascript", "frontend"},
			wantErr:    false,
		},
		{
			name:       "Case duplicates are removed after normalization",
			input:      []string{"JavaScript", "javascript"},
			wantResult: []string{"javascript"},
			wantErr:    false,
		},
		{
			name:       "Exactly 10 tags is allowed",
			input:      []string{"a1", "a2", "a3", "a4", "a5", "a6", "a7", "a8", "a9", "a0"},
			wantResult: []string{"a1", "a2", "a3", "a4", "a5", "a6", "a7", "a8", "a9", "a0"},
			wantErr:    false,
		},
		{
			name:       "11 unique tags is rejected",
			input:      []string{"a1", "a2", "a3", "a4", "a5", "a6", "a7", "a8", "a9", "a0", "a11"},
			wantErr:    true,
			wantErrMsg: "a link can have at most 10 tags",
		},
		{
			name:       "12 tags with duplicates reducing to 10 is allowed",
			input:      []string{"a1", "a2", "a3", "a4", "a5", "a6", "a7", "a8", "a9", "a0", "a1", "a2"},
			wantResult: []string{"a1", "a2", "a3", "a4", "a5", "a6", "a7", "a8", "a9", "a0"},
			wantErr:    false,
		},
		{
			name:       "Invalid tag in list fails validation",
			input:      []string{"javascript", "a"},
			wantErr:    true,
			wantErrMsg: "tag name must be between 2 and 50 characters",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotResult, gotErr := ValidateTagNames(tt.input)
			if gotErr != nil {
				if !tt.wantErr {
					t.Errorf("ValidateTagNames() failed: %v", gotErr)
				}
				if gotErr.Error() != tt.wantErrMsg {
					t.Errorf("ValidateTagNames() error message = %v, want %v", gotErr.Error(), tt.wantErrMsg)
				}
				return
			}
			if tt.wantErr {
				t.Fatal("ValidateTagNames() succeeded unexpectedly")
			}
			if len(gotResult) != len(tt.wantResult) {
				t.Fatalf("ValidateTagNames() result length = %v, want %v", len(gotResult), len(tt.wantResult))
			}
			for i := range gotResult {
				if gotResult[i] != tt.wantResult[i] {
					t.Errorf("ValidateTagNames() result[%d] = %v, want %v", i, gotResult[i], tt.wantResult[i])
				}
			}
		})
	}
}

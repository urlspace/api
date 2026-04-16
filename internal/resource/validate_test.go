package resource

import (
	"strings"
	"testing"
)

func Test_validateTitle(t *testing.T) {

	tests := []struct {
		name       string
		input      string
		wantResult string
		wantErr    bool
		wantErrMsg string
	}{
		{
			name:       "Valid title",
			input:      "My Resource",
			wantResult: "My Resource",
			wantErr:    false,
		},
		{
			name:       "Title is trimmed",
			input:      "  My Resource  ",
			wantResult: "My Resource",
			wantErr:    false,
		},
		{
			name:       "Title is too short",
			input:      "ab",
			wantErr:    true,
			wantErrMsg: "title must be between 3 and 255 characters",
		},
		{
			name:       "Title is too long",
			input:      strings.Repeat("a", 256),
			wantErr:    true,
			wantErrMsg: "title must be between 3 and 255 characters",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotResult, gotErr := validateTitle(tt.input)
			if gotErr != nil {
				if !tt.wantErr {
					t.Errorf("ResourceTitle() failed: %v", gotErr)
				}
				if gotErr.Error() != tt.wantErrMsg {
					t.Errorf("ResourceTitle() error message = %v, want %v", gotErr.Error(), tt.wantErrMsg)
				}
				return
			}
			if tt.wantErr {
				t.Fatal("ResourceTitle() succeeded unexpectedly")
			}
			if tt.wantResult != "" && gotResult != tt.wantResult {
				t.Errorf("ResourceTitle() result = %v, want %v", gotResult, tt.wantResult)
			}
		})
	}
}

func Test_validateDescription(t *testing.T) {
	tests := []struct {
		name       string
		input      string
		wantResult string
		wantErr    bool
		wantErrMsg string
	}{
		{
			name:       "Valid description",
			input:      "A helpful resource description.",
			wantResult: "A helpful resource description.",
			wantErr:    false,
		},
		{
			name:    "Empty description",
			input:   "",
			wantErr: false,
		},
		{
			name:       "Description is trimmed",
			input:      "  A helpful resource description.  ",
			wantResult: "A helpful resource description.",
			wantErr:    false,
		},
		{
			name:       "Description is too long",
			input:      strings.Repeat("a", 513),
			wantErr:    true,
			wantErrMsg: "description must be less than 512 characters",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotResult, gotErr := validateDescription(tt.input)
			if gotErr != nil {
				if !tt.wantErr {
					t.Errorf("ResourceDescription() failed: %v", gotErr)
				}
				if gotErr.Error() != tt.wantErrMsg {
					t.Errorf("ResourceDescription() error message = %v, want %v", gotErr.Error(), tt.wantErrMsg)
				}
				return
			}
			if tt.wantErr {
				t.Fatal("ResourceDescription() succeeded unexpectedly")
			}
			if tt.wantResult != "" && gotResult != tt.wantResult {
				t.Errorf("ResourceDescription() result = %v, want %v", gotResult, tt.wantResult)
			}
		})
	}
}

func Test_validateURL(t *testing.T) {
	tests := []struct {
		name       string
		input      string
		wantErr    bool
		wantErrMsg string
	}{
		{
			name:    "Valid URL",
			input:   "https://example.com",
			wantErr: false,
		},
		{
			name:       "Empty URL",
			input:      "",
			wantErr:    true,
			wantErrMsg: "url is invalid",
		},
		{
			name:       "URL without scheme",
			input:      "example.com/path",
			wantErr:    true,
			wantErrMsg: "url is invalid",
		},
		{
			name:       "URL without host",
			input:      "http://",
			wantErr:    true,
			wantErrMsg: "url is invalid",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, gotErr := validateURL(tt.input)
			if gotErr != nil {
				if !tt.wantErr {
					t.Errorf("Url() failed: %v", gotErr)
				}
				if gotErr.Error() != tt.wantErrMsg {
					t.Errorf("Url() error message = %v, want %v", gotErr.Error(), tt.wantErrMsg)
				}
				return
			}
			if tt.wantErr {
				t.Fatal("Url() succeeded unexpectedly")
			}
		})
	}
}

func Test_validateFavourite(t *testing.T) {
	tests := []struct {
		name       string
		input      *bool
		wantErr    bool
		wantErrMsg string
	}{
		{
			name:    "Valid favourite (true)",
			input:   new(true),
			wantErr: false,
		},
		{
			name:    "Valid favourite (false)",
			input:   new(false),
			wantErr: false,
		},
		{
			name:       "Nil favourite",
			input:      nil,
			wantErr:    true,
			wantErrMsg: "favourite field is required",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, gotErr := validateFavourite(tt.input)
			if gotErr != nil {
				if !tt.wantErr {
					t.Errorf("ResourceFavourite() failed: %v", gotErr)
				}
				if gotErr.Error() != tt.wantErrMsg {
					t.Errorf("ResourceFavourite() error message = %v, want %v", gotErr.Error(), tt.wantErrMsg)
				}
				return
			}
			if tt.wantErr {
				t.Fatal("ResourceFavourite() succeeded unexpectedly")
			}
		})
	}
}

func Test_validateReadLater(t *testing.T) {
	tests := []struct {
		name       string
		input      *bool
		wantErr    bool
		wantErrMsg string
	}{
		{
			name:    "Valid readLater (true)",
			input:   new(true),
			wantErr: false,
		},
		{
			name:    "Valid readLater (false)",
			input:   new(false),
			wantErr: false,
		},
		{
			name:       "Nil readLater",
			input:      nil,
			wantErr:    true,
			wantErrMsg: "readLater field is required",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, gotErr := validateReadLater(tt.input)
			if gotErr != nil {
				if !tt.wantErr {
					t.Errorf("ResourceReadLater() failed: %v", gotErr)
				}
				if gotErr.Error() != tt.wantErrMsg {
					t.Errorf("ResourceReadLater() error message = %v, want %v", gotErr.Error(), tt.wantErrMsg)
				}
				return
			}
			if tt.wantErr {
				t.Fatal("ResourceReadLater() succeeded unexpectedly")
			}
		})
	}
}

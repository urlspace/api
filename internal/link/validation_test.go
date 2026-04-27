package link

import (
	"strings"
	"testing"
)

func Test_ValidateTitle(t *testing.T) {

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
		{
			name:       "Title with null byte is rejected",
			input:      "My \x00 Resource",
			wantErr:    true,
			wantErrMsg: "title must not contain control characters",
		},
		{
			name:       "Title with tab is rejected",
			input:      "My \t Resource",
			wantErr:    true,
			wantErrMsg: "title must not contain control characters",
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
			gotResult, gotErr := ValidateTitle(tt.input)
			if gotErr != nil {
				if !tt.wantErr {
					t.Errorf("ValidateTitle() failed: %v", gotErr)
				}
				if gotErr.Error() != tt.wantErrMsg {
					t.Errorf("ValidateTitle() error message = %v, want %v", gotErr.Error(), tt.wantErrMsg)
				}
				return
			}
			if tt.wantErr {
				t.Fatal("ValidateTitle() succeeded unexpectedly")
			}
			if tt.wantResult != "" && gotResult != tt.wantResult {
				t.Errorf("ValidateTitle() result = %v, want %v", gotResult, tt.wantResult)
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
			input:      strings.Repeat("ą", 257),
			wantResult: strings.Repeat("ą", 257),
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

func Test_ValidateURL(t *testing.T) {
	tests := []struct {
		name       string
		input      string
		wantResult string
		wantErr    bool
		wantErrMsg string
	}{
		{
			name:       "Valid URL",
			input:      "https://example.com",
			wantResult: "https://example.com",
			wantErr:    false,
		},
		{
			name:       "URL is trimmed",
			input:      "  https://example.com  ",
			wantResult: "https://example.com",
			wantErr:    false,
		},
		{
			name:       "URL scheme and host are normalized to lowercase",
			input:      "HTTP://Example.COM/path",
			wantResult: "http://example.com/path",
			wantErr:    false,
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
		{
			name:       "URL with credentials is rejected",
			input:      "http://user:secret@example.com",
			wantErr:    true,
			wantErrMsg: "url is invalid",
		},
		{
			name:       "javascript: scheme is rejected",
			input:      "javascript:alert(1)",
			wantErr:    true,
			wantErrMsg: "url is invalid",
		},
		{
			name:       "data: scheme is rejected",
			input:      "data:text/html,<script>alert(1)</script>",
			wantErr:    true,
			wantErrMsg: "url is invalid",
		},
		{
			name:       "file: scheme is rejected",
			input:      "file:///etc/passwd",
			wantErr:    true,
			wantErrMsg: "url is invalid",
		},
		{
			name:       "ftp: scheme is rejected",
			input:      "ftp://example.com",
			wantErr:    true,
			wantErrMsg: "url is invalid",
		},
		{
			name:       "URL exceeding 2048 characters is rejected",
			input:      "https://example.com/" + strings.Repeat("a", 2029),
			wantErr:    true,
			wantErrMsg: "url must be at most 2048 characters",
		},
		{
			name:       "localhost URL is rejected",
			input:      "http://localhost",
			wantErr:    true,
			wantErrMsg: "url must not point to a private or local address",
		},
		{
			name:       "Loopback IP is rejected",
			input:      "http://127.0.0.1:8080",
			wantErr:    true,
			wantErrMsg: "url must not point to a private or local address",
		},
		{
			name:       "Private network IP (10.x) is rejected",
			input:      "http://10.0.0.1",
			wantErr:    true,
			wantErrMsg: "url must not point to a private or local address",
		},
		{
			name:       "Private network IP (192.168.x) is rejected",
			input:      "http://192.168.1.1",
			wantErr:    true,
			wantErrMsg: "url must not point to a private or local address",
		},
		{
			name:       "AWS metadata endpoint is rejected",
			input:      "http://169.254.169.254/latest/meta-data/",
			wantErr:    true,
			wantErrMsg: "url must not point to a private or local address",
		},
		{
			name:       "IPv6 loopback is rejected",
			input:      "http://[::1]",
			wantErr:    true,
			wantErrMsg: "url must not point to a private or local address",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotResult, gotErr := ValidateURL(tt.input)
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
			if tt.wantResult != "" && gotResult != tt.wantResult {
				t.Errorf("Url() result = %v, want %v", gotResult, tt.wantResult)
			}
		})
	}
}

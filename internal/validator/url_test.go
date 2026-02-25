package validator_test

import (
	"testing"

	"github.com/hreftools/api/internal/validator"
)

func TestUrl(t *testing.T) {
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
			gotErr := validator.Url(tt.input)
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

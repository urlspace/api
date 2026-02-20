package validator_test

import (
	"github.com/hreftools/api/internal/validator"
	"testing"
)

func TestUrl(t *testing.T) {
	tests := []struct {
		name       string
		uri        string
		wantErr    bool
		wantErrMsg string
	}{
		{
			name:    "Valid URL",
			uri:     "https://example.com",
			wantErr: false,
		},
		{
			name:       "Empty URL",
			uri:        "",
			wantErr:    true,
			wantErrMsg: "url is invalid",
		},
		{
			name:       "URL without scheme",
			uri:        "example.com/path",
			wantErr:    true,
			wantErrMsg: "url is invalid",
		},
		{
			name:       "URL without host",
			uri:        "http://",
			wantErr:    true,
			wantErrMsg: "url is invalid",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotErr := validator.Url(tt.uri)
			if gotErr != nil {
				if !tt.wantErr {
					t.Errorf("Url() failed: %v", gotErr)
				}
				return
			}
			if tt.wantErr {
				t.Fatal("Url() succeeded unexpectedly")
			}
		})
	}
}

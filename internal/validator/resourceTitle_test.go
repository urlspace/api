package validator_test

import (
	"strings"
	"testing"

	"github.com/hreftools/api/internal/validator"
)

func TestResourceTitle(t *testing.T) {
	tests := []struct {
		name       string
		rt         string
		wantErr    bool
		wantErrMsg string
	}{
		{
			name:    "Valid title",
			rt:      "My Resource",
			wantErr: false,
		},
		{
			name:       "Title is too short",
			rt:         "ab",
			wantErr:    true,
			wantErrMsg: "title must be between 3 and 255 characters",
		},
		{
			name:       "Title is too long",
			rt:         strings.Repeat("a", 256),
			wantErr:    true,
			wantErrMsg: "title must be between 3 and 255 characters",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotErr := validator.ResourceTitle(tt.rt)
			if gotErr != nil {
				if !tt.wantErr {
					t.Errorf("ResourceTitle() failed: %v", gotErr)
				}
				return
			}
			if tt.wantErr {
				t.Fatal("ResourceTitle() succeeded unexpectedly")
			}
		})
	}
}

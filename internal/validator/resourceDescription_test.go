package validator_test

import (
	"strings"
	"testing"

	"github.com/hreftools/api/internal/validator"
)

func TestResourceDescription(t *testing.T) {
	tests := []struct {
		name       string
		input      string
		wantErr    bool
		wantErrMsg string
	}{
		{
			name:    "Valid description",
			input:   "A helpful resource description.",
			wantErr: false,
		},
		{
			name:    "Empty description",
			input:   "",
			wantErr: false,
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
			gotErr := validator.ResourceDescription(tt.input)
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
		})
	}
}

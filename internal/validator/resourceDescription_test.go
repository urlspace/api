package validator_test

import (
	"strings"
	"testing"

	"github.com/hreftools/api/internal/validator"
)

func TestResourceDescription(t *testing.T) {
	tests := []struct {
		name       string
		rd         string
		wantErr    bool
		wantErrMsg string
	}{
		{
			name:    "Valid description",
			rd:      "A helpful resource description.",
			wantErr: false,
		},
		{
			name:    "Empty description",
			rd:      "",
			wantErr: false,
		},
		{
			name:       "Description is too long",
			rd:         strings.Repeat("a", 513),
			wantErr:    true,
			wantErrMsg: "description must be less than 512 characters",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotErr := validator.ResourceDescription(tt.rd)
			if gotErr != nil {
				if !tt.wantErr {
					t.Errorf("ResourceDescription() failed: %v", gotErr)
				}
				return
			}
			if tt.wantErr {
				t.Fatal("ResourceDescription() succeeded unexpectedly")
			}
		})
	}
}

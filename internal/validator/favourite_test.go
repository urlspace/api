package validator_test

import (
	"testing"

	"github.com/hreftools/api/internal/validator"
)

func TestResourceFavourite(t *testing.T) {
	tests := []struct {
		name       string
		input      *bool
		wantErr    bool
		wantErrMsg string
	}{
		{
			name:    "Valid favourite (true)",
			input:   boolPtr(true),
			wantErr: false,
		},
		{
			name:    "Valid favourite (false)",
			input:   boolPtr(false),
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
			gotErr := validator.ResourceFavourite(tt.input)
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

func boolPtr(b bool) *bool {
	return &b
}

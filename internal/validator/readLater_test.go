package validator_test

import (
	"testing"

	"github.com/hreftools/api/internal/validator"
)

func TestResourceReadLater(t *testing.T) {
	tests := []struct {
		name       string
		input      *bool
		wantErr    bool
		wantErrMsg string
	}{
		{
			name:    "Valid readLater (true)",
			input:   boolPtr(true),
			wantErr: false,
		},
		{
			name:    "Valid readLater (false)",
			input:   boolPtr(false),
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
			gotErr := validator.ResourceReadLater(tt.input)
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

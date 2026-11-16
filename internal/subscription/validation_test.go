package subscription

import "testing"

func Test_validateReturnTo(t *testing.T) {
	tests := []struct {
		name       string
		input      string
		want       string
		wantErr    bool
		wantErrMsg string
	}{
		{
			name:  "Home returns to pricing",
			input: "home",
			want:  "/#pricing",
		},
		{
			name:  "Dashboard",
			input: "dashboard",
			want:  "/dashboard",
		},
		{
			name:  "Settings",
			input: "settings",
			want:  "/settings",
		},
		{
			name:       "Missing",
			input:      "",
			wantErr:    true,
			wantErrMsg: "returnTo is required",
		},
		{
			name:       "Unknown key",
			input:      "pricing",
			wantErr:    true,
			wantErrMsg: "returnTo must be one of home, dashboard or settings",
		},
		{
			name:       "URL instead of key",
			input:      "https://example.com",
			wantErr:    true,
			wantErrMsg: "returnTo must be one of home, dashboard or settings",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, gotErr := validateReturnTo(tt.input)
			if gotErr != nil {
				if !tt.wantErr {
					t.Errorf("validateReturnTo() failed: %v", gotErr)
				}
				if gotErr.Error() != tt.wantErrMsg {
					t.Errorf("validateReturnTo() error message = %v, want %v", gotErr.Error(), tt.wantErrMsg)
				}
				return
			}
			if tt.wantErr {
				t.Fatal("validateReturnTo() succeeded unexpectedly")
			}
			if got != tt.want {
				t.Errorf("validateReturnTo() = %v, want %v", got, tt.want)
			}
		})
	}
}

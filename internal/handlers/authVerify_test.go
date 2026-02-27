package handlers_test

import (
	"net/http"
	"testing"

	"github.com/hreftools/api/internal/handlers"
	"github.com/hreftools/api/internal/store"
)

func TestAuthVerifyBody_Normalize(t *testing.T) {
	tests := []struct {
		name     string
		input    handlers.AuthVerifyBody
		expected handlers.AuthVerifyBody
	}{
		{
			name: "Produces no change to already normalized input",
			input: handlers.AuthVerifyBody{
				Token: "12345678-1234-1234-1234-123456789abc",
			},
			expected: handlers.AuthVerifyBody{
				Token: "12345678-1234-1234-1234-123456789abc",
			},
		},
		{
			name: "Trim token",
			input: handlers.AuthVerifyBody{
				Token: " 12345678-1234-1234-1234-123456789abc ",
			},
			expected: handlers.AuthVerifyBody{
				Token: "12345678-1234-1234-1234-123456789abc",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			b := tt.input
			b.Normalize()

			if b.Token != tt.expected.Token {
				t.Errorf("Normalize() Email = %v, want %v", b.Token, tt.expected.Token)
			}
		})
	}
}

func TestAuthVerifyBody_Validate(t *testing.T) {
	tests := []struct {
		name    string
		input   handlers.AuthVerifyBody
		wantErr bool
	}{
		{
			name: "Valid input",
			input: handlers.AuthVerifyBody{
				Token: "12345678-1234-1234-1234-123456789abc",
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			b := tt.input
			gotErr := b.Validate()

			if gotErr != nil {
				if !tt.wantErr {
					t.Errorf("Validate() failed: %v", gotErr)
				}
				return
			}
			if tt.wantErr {
				t.Fatal("Validate() succeeded unexpectedly")
			}
		})
	}
}

func TestAuthVerify(t *testing.T) {
	tests := []struct {
		name string // description of this test case
		// Named input parameters for target function.
		s    *store.Store
		want http.HandlerFunc
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := handlers.AuthVerify(tt.s)
			// TODO: update the condition below to compare got with tt.want.
			if true {
				t.Errorf("AuthVerify() = %v, want %v", got, tt.want)
			}
		})
	}
}

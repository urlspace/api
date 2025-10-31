package server_test

import (
	"github.com/zapi-sh/api/internal/server"
	"net/http"
	"testing"
)

func TestNew(t *testing.T) {
	tests := []struct {
		name string // description of this test case
		want *http.Server
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := server.New()
			// TODO: update the condition below to compare got with tt.want.
			if true {
				t.Errorf("New() = %v, want %v", got, tt.want)
			}
		})
	}
}

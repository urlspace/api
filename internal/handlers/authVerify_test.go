package handlers_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/hreftools/api/internal/handlers"
	"github.com/hreftools/api/internal/response"
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
	t.Run("fails on incorrect body", func(t *testing.T) {
		s := setupTestStore(t)

		handler := handlers.AuthVerify(s)

		body := `this is not a json body`
		req := httptest.NewRequest("POST", "/auth/verify", strings.NewReader(body))
		rec := httptest.NewRecorder()

		handler.ServeHTTP(rec, req)

		if rec.Code != http.StatusBadRequest {
			t.Errorf("expected 400, got %d", rec.Code)
		}

		var res response.ErrorResponse
		if err := json.NewDecoder(rec.Body).Decode(&res); err != nil {
			t.Fatalf("failed to decode response: %v", err)
		}

		if expected := "error"; res.Status != expected {
			t.Errorf("expected status: %s, got %s", expected, res.Status)
		}

		if expected := "invalid request body"; res.Data != expected {
			t.Errorf("expected data: %s, got %s", expected, res.Data)
		}
	})

	t.Run("fails on unexpected field in body", func(t *testing.T) {
		s := setupTestStore(t)

		handler := handlers.AuthVerify(s)

		body := `{"token":"12345678-1234-1234-1234-123456789abc","unexpected":"field"}`
		req := httptest.NewRequest("POST", "/auth/verify", strings.NewReader(body))
		rec := httptest.NewRecorder()

		handler.ServeHTTP(rec, req)

		if rec.Code != http.StatusBadRequest {
			t.Errorf("expected 400, got %d", rec.Code)
		}

		var res response.ErrorResponse
		if err := json.NewDecoder(rec.Body).Decode(&res); err != nil {
			t.Fatalf("failed to decode response: %v", err)
		}

		if expected := "error"; res.Status != expected {
			t.Errorf("expected status: %s, got %s", expected, res.Status)
		}

		if expected := "invalid request body"; res.Data != expected {
			t.Errorf("expected data: %s, got %s", expected, res.Data)
		}
	})

	t.Run("fails on invalid request body", func(t *testing.T) {
		s := setupTestStore(t)

		handler := handlers.AuthVerify(s)

		body := `{"token":""}`
		req := httptest.NewRequest("POST", "/auth/verify", strings.NewReader(body))
		rec := httptest.NewRecorder()

		handler.ServeHTTP(rec, req)

		if rec.Code != http.StatusBadRequest {
			t.Errorf("expected 400, got %d", rec.Code)
		}

		var res response.ErrorResponse
		if err := json.NewDecoder(rec.Body).Decode(&res); err != nil {
			t.Fatalf("failed to decode response: %v", err)
		}

		if expected := "error"; res.Status != expected {
			t.Errorf("expected status: %s, got %s", expected, res.Status)
		}

		if expected := "token is required"; res.Data != expected {
			t.Errorf("expected data: %s, got %s", expected, res.Data)
		}
	})

	t.Run("fails on non-existing token", func(t *testing.T) {
		s := setupTestStore(t)

		handler := handlers.AuthVerify(s)

		token := uuid.New().String()
		body := `{"token":"` + token + `"}`
		req := httptest.NewRequest("POST", "/auth/verify", strings.NewReader(body))
		rec := httptest.NewRecorder()

		handler.ServeHTTP(rec, req)

		if rec.Code != http.StatusNotFound {
			t.Errorf("expected 404, got %d", rec.Code)
		}

		var res response.ErrorResponse
		if err := json.NewDecoder(rec.Body).Decode(&res); err != nil {
			t.Fatalf("failed to decode response: %v", err)
		}

		if expected := "error"; res.Status != expected {
			t.Errorf("expected status: %s, got %s", expected, res.Status)
		}

		if expected := "entry not found"; res.Data != expected {
			t.Errorf("expected data: %s, got %s", expected, res.Data)
		}
	})

	t.Run("fails on expired token", func(t *testing.T) {
		s := setupTestStore(t)

		token := uuid.New()

		s.Users.Create(context.Background(), store.UserCreateParams{
			Email:                           "test@example.com",
			EmailVerified:                   false,
			EmailVerificationToken:          uuid.NullUUID{Valid: true, UUID: token},
			EmailVerificationTokenExpiresAt: new(time.Now().Add(-time.Hour)),
			Password:                        "strongpassword",
			Username:                        "testuser",
			IsAdmin:                         false,
			IsPro:                           false,
		})

		handler := handlers.AuthVerify(s)

		body := `{"token":"` + token.String() + `"}`
		req := httptest.NewRequest("POST", "/auth/verify", strings.NewReader(body))
		rec := httptest.NewRecorder()

		handler.ServeHTTP(rec, req)

		if rec.Code != http.StatusBadRequest {
			t.Errorf("expected 400, got %d", rec.Code)
		}

		var res response.ErrorResponse
		if err := json.NewDecoder(rec.Body).Decode(&res); err != nil {
			t.Fatalf("failed to decode response: %v", err)
		}

		if expected := "error"; res.Status != expected {
			t.Errorf("expected status: %s, got %s", expected, res.Status)
		}

		if expected := "token has expired"; res.Data != expected {
			t.Errorf("expected data: %s, got %s", expected, res.Data)
		}
	})

	t.Run("success", func(t *testing.T) {
		s := setupTestStore(t)

		token := uuid.New()

		s.Users.Create(context.Background(), store.UserCreateParams{
			Email:                           "test@example.com",
			EmailVerified:                   false,
			EmailVerificationToken:          uuid.NullUUID{Valid: true, UUID: token},
			EmailVerificationTokenExpiresAt: new(time.Now().Add(handlers.TokenExpiryDuration)),
			Password:                        "strongpassword",
			Username:                        "testuser",
			IsAdmin:                         false,
			IsPro:                           false,
		})

		handler := handlers.AuthVerify(s)

		body := `{"token":"` + token.String() + `"}`
		req := httptest.NewRequest("POST", "/auth/verify", strings.NewReader(body))
		rec := httptest.NewRecorder()

		handler.ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Errorf("expected 200, got %d", rec.Code)
		}

		var res handlers.AuthVerifyResponse
		if err := json.NewDecoder(rec.Body).Decode(&res); err != nil {
			t.Fatalf("failed to decode response: %v", err)
		}

		if expected := "ok"; res.Status != expected {
			t.Errorf("expected status: %s, got %s", expected, res.Status)
		}

		if expected := "testuser"; res.Data.Username != expected {
			t.Errorf("expected username %s, got %s", expected, res.Data.Username)
		}

		if expected := "test@example.com"; res.Data.Email != expected {
			t.Errorf("expected email %s, got %s", expected, res.Data.Email)
		}

		if expected := true; res.Data.EmailVerified != expected {
			t.Errorf("expected emailVerified %v, got %v", expected, res.Data.EmailVerified)
		}

		if expected := false; res.Data.IsAdmin != expected {
			t.Errorf("expected isAdmin %v, got %v", expected, res.Data.IsAdmin)
		}

		if expected := false; res.Data.IsPro != expected {
			t.Errorf("expected isPro %v, got %v", expected, res.Data.IsPro)
		}
	})
}

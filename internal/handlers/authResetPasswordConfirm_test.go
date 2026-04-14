package handlers_test

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/hreftools/api/internal/config"
	"github.com/hreftools/api/internal/handlers"
	"github.com/hreftools/api/internal/response"
	"github.com/hreftools/api/internal/user"
)

func TestAuthResetPasswordConfirmBody_Normalize(t *testing.T) {
	tests := []struct {
		name     string
		input    handlers.AuthResetPasswordConfirmBody
		expected handlers.AuthResetPasswordConfirmBody
	}{
		{
			name: "Produces no change to already normalized input",
			input: handlers.AuthResetPasswordConfirmBody{
				Token:    "12345678-1234-1234-1234-123456789abc",
				Password: "newpassword1234",
			},
			expected: handlers.AuthResetPasswordConfirmBody{
				Token:    "12345678-1234-1234-1234-123456789abc",
				Password: "newpassword1234",
			},
		},
		{
			name: "Trim token",
			input: handlers.AuthResetPasswordConfirmBody{
				Token:    " 12345678-1234-1234-1234-123456789abc ",
				Password: "newpassword1234",
			},
			expected: handlers.AuthResetPasswordConfirmBody{
				Token:    "12345678-1234-1234-1234-123456789abc",
				Password: "newpassword1234",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			b := tt.input
			b.Normalize()

			if b.Token != tt.expected.Token {
				t.Errorf("Normalize() Token = %v, want %v", b.Token, tt.expected.Token)
			}

			if b.Password != tt.expected.Password {
				t.Errorf("Normalize() Password = %v, want %v", b.Password, tt.expected.Password)
			}
		})
	}
}

func TestAuthResetPasswordConfirmBody_Validate(t *testing.T) {
	tests := []struct {
		name    string
		input   handlers.AuthResetPasswordConfirmBody
		wantErr bool
	}{
		{
			name: "Valid input",
			input: handlers.AuthResetPasswordConfirmBody{
				Token:    "12345678-1234-1234-1234-123456789abc",
				Password: "newstrongpassword",
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

func TestAuthResetPasswordConfirm(t *testing.T) {
	t.Run("fails on incorrect body", func(t *testing.T) {
		svc, _, _ := setupServices(t)

		handler := handlers.AuthResetPasswordConfirm(svc)

		body := `this is not a json body`
		req := httptest.NewRequest("POST", "/auth/reset-password-confirm", strings.NewReader(body))
		rec := httptest.NewRecorder()

		handler.ServeHTTP(rec, req)

		if expected := 400; rec.Code != http.StatusBadRequest {
			t.Errorf("expected %d, got %d", expected, rec.Code)
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
		svc, _, _ := setupServices(t)

		handler := handlers.AuthResetPasswordConfirm(svc)

		body := `{"token":"12345678-1234-1234-1234-123456789abc","password":"newpassword1234","unexpected":"field"}`
		req := httptest.NewRequest("POST", "/auth/reset-password-confirm", strings.NewReader(body))
		rec := httptest.NewRecorder()

		handler.ServeHTTP(rec, req)

		if expected := 400; rec.Code != http.StatusBadRequest {
			t.Errorf("expected %d, got %d", expected, rec.Code)
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

	t.Run("fails on invalid request body - empty token", func(t *testing.T) {
		svc, _, _ := setupServices(t)

		handler := handlers.AuthResetPasswordConfirm(svc)

		body := `{"token":"","password":"newpassword1234"}`
		req := httptest.NewRequest("POST", "/auth/reset-password-confirm", strings.NewReader(body))
		rec := httptest.NewRecorder()

		handler.ServeHTTP(rec, req)

		if expected := 400; rec.Code != http.StatusBadRequest {
			t.Errorf("expected %d, got %d", expected, rec.Code)
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

	t.Run("fails on invalid request body - empty password", func(t *testing.T) {
		svc, _, _ := setupServices(t)

		handler := handlers.AuthResetPasswordConfirm(svc)

		body := `{"token":"12345678-1234-1234-1234-123456789abc","password":""}`
		req := httptest.NewRequest("POST", "/auth/reset-password-confirm", strings.NewReader(body))
		rec := httptest.NewRecorder()

		handler.ServeHTTP(rec, req)

		if expected := 400; rec.Code != http.StatusBadRequest {
			t.Errorf("expected %d, got %d", expected, rec.Code)
		}

		var res response.ErrorResponse
		if err := json.NewDecoder(rec.Body).Decode(&res); err != nil {
			t.Fatalf("failed to decode response: %v", err)
		}

		if expected := "error"; res.Status != expected {
			t.Errorf("expected status: %s, got %s", expected, res.Status)
		}
	})

	t.Run("fails on non-existing token", func(t *testing.T) {
		svc, _, _ := setupServices(t)

		handler := handlers.AuthResetPasswordConfirm(svc)

		token := uuid.New().String()
		body := fmt.Sprintf(`{"token":"%s","password":"newpassword1234"}`, token)
		req := httptest.NewRequest("POST", "/auth/reset-password-confirm", strings.NewReader(body))
		rec := httptest.NewRecorder()

		handler.ServeHTTP(rec, req)

		if expected := 404; rec.Code != http.StatusNotFound {
			t.Errorf("expected %d, got %d", expected, rec.Code)
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
		svc, _, _ := setupServices(t)

		token := uuid.New()

		u, _ := svc.Repo.Create(context.Background(), user.CreateParams{
			Email:                           "test@example.com",
			EmailVerified:                   true,
			EmailVerificationToken:          uuid.NullUUID{},
			EmailVerificationTokenExpiresAt: nil,
			Password:                        "strongpassword",
			Username:                        "testuser",
			IsAdmin:                         false,
			IsPro:                           false,
		})

		exp := time.Now().Add(-time.Hour)
		svc.Repo.UpdatePasswordResetToken(context.Background(), user.UpdatePasswordResetTokenParams{
			ID:                          u.ID,
			PasswordResetToken:          uuid.NullUUID{Valid: true, UUID: token},
			PasswordResetTokenExpiresAt: &exp,
		})

		handler := handlers.AuthResetPasswordConfirm(svc)

		body := fmt.Sprintf(`{"token":"%s","password":"newpassword1234"}`, token.String())
		req := httptest.NewRequest("POST", "/auth/reset-password-confirm", strings.NewReader(body))
		rec := httptest.NewRecorder()

		handler.ServeHTTP(rec, req)

		if expected := 400; rec.Code != http.StatusBadRequest {
			t.Errorf("expected %d, got %d", expected, rec.Code)
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
		svc, _, _ := setupServices(t)

		token := uuid.New()

		u, _ := svc.Repo.Create(context.Background(), user.CreateParams{
			Email:                           "test@example.com",
			EmailVerified:                   true,
			EmailVerificationToken:          uuid.NullUUID{},
			EmailVerificationTokenExpiresAt: nil,
			Password:                        "strongpassword",
			Username:                        "testuser",
			IsAdmin:                         false,
			IsPro:                           false,
		})

		exp := time.Now().Add(config.PasswordResetTokenExpiryDuration)
		svc.Repo.UpdatePasswordResetToken(context.Background(), user.UpdatePasswordResetTokenParams{
			ID:                          u.ID,
			PasswordResetToken:          uuid.NullUUID{Valid: true, UUID: token},
			PasswordResetTokenExpiresAt: &exp,
		})

		handler := handlers.AuthResetPasswordConfirm(svc)

		body := `{"token":"` + token.String() + `","password":"newstrongpassword"}`
		req := httptest.NewRequest("POST", "/auth/reset-password-confirm", strings.NewReader(body))
		rec := httptest.NewRecorder()

		handler.ServeHTTP(rec, req)

		if expected := 200; rec.Code != http.StatusOK {
			t.Errorf("expected %d, got %d", expected, rec.Code)
		}

		var res handlers.AuthResetPasswordConfirmResponse
		if err := json.NewDecoder(rec.Body).Decode(&res); err != nil {
			t.Fatalf("failed to decode response: %v", err)
		}

		if expected := "ok"; res.Status != expected {
			t.Errorf("expected status: %s, got %s", expected, res.Status)
		}

		if expected := "ok"; res.Data != expected {
			t.Errorf("expected data %s, got %s", expected, res.Data)
		}
	})
}

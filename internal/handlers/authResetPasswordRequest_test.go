package handlers_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"slices"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/hreftools/api/internal/config"
	"github.com/hreftools/api/internal/handlers"
	"github.com/hreftools/api/internal/response"
	"github.com/hreftools/api/internal/user"
)

func TestAuthResetPasswordRequestBody_Normalize(t *testing.T) {
	tests := []struct {
		name     string
		input    handlers.AuthResetPasswordRequestBody
		expected handlers.AuthResetPasswordRequestBody
	}{
		{
			name: "Produces no change to already normalized input",
			input: handlers.AuthResetPasswordRequestBody{
				Email: "user@email.com",
			},
			expected: handlers.AuthResetPasswordRequestBody{
				Email: "user@email.com",
			},
		},
		{
			name: "Trim and lowercase email",
			input: handlers.AuthResetPasswordRequestBody{
				Email: "  User@Email.com  ",
			},
			expected: handlers.AuthResetPasswordRequestBody{
				Email: "user@email.com",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			b := tt.input
			b.Normalize()

			if b.Email != tt.expected.Email {
				t.Errorf("Normalize() Email = %v, want %v", b.Email, tt.expected.Email)
			}
		})
	}
}

func TestAuthResetPasswordRequestBody_Validate(t *testing.T) {
	tests := []struct {
		name    string
		input   handlers.AuthResetPasswordRequestBody
		wantErr bool
	}{
		{
			name: "Valid input",
			input: handlers.AuthResetPasswordRequestBody{
				Email: "valid@email.com",
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

func TestAuthResetPasswordRequest(t *testing.T) {
	t.Run("fails on incorrect body", func(t *testing.T) {
		svc, _, emailSenderMock := setupServices(t)

		handler := handlers.AuthResetPasswordRequest(svc)

		body := `this is not a json body`
		req := httptest.NewRequest("POST", "/auth/reset-password-request", strings.NewReader(body))
		rec := httptest.NewRecorder()

		handler.ServeHTTP(rec, req)

		if expected := 400; rec.Code != http.StatusBadRequest {
			t.Errorf("expected %d, got %d", expected, rec.Code)
		}

		if emailSenderMock.called {
			t.Error("expected email not to be sent")
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
		svc, _, emailSenderMock := setupServices(t)

		handler := handlers.AuthResetPasswordRequest(svc)

		body := `{"email":"test@example.com","unexpected":"field"}`
		req := httptest.NewRequest("POST", "/auth/reset-password-request", strings.NewReader(body))
		rec := httptest.NewRecorder()

		handler.ServeHTTP(rec, req)

		if expected := 400; rec.Code != http.StatusBadRequest {
			t.Errorf("expected %d, got %d", expected, rec.Code)
		}

		if emailSenderMock.called {
			t.Error("expected email not to be sent")
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

	t.Run("invalid request body", func(t *testing.T) {
		svc, _, emailSenderMock := setupServices(t)

		handler := handlers.AuthResetPasswordRequest(svc)

		body := `{"email":""}`

		req := httptest.NewRequest("POST", "/auth/reset-password-request", strings.NewReader(body))
		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, req)

		if expected := 400; rec.Code != http.StatusBadRequest {
			t.Errorf("expected %d, got %d", expected, rec.Code)
		}

		if emailSenderMock.called {
			t.Error("expected email not to be sent")
		}

		var res response.ErrorResponse
		if err := json.NewDecoder(rec.Body).Decode(&res); err != nil {
			t.Fatalf("failed to decode response: %v", err)
		}

		if expected := "error"; res.Status != expected {
			t.Errorf("expected status: %s, got %s", expected, res.Status)
		}

		if expected := "email is required"; res.Data != expected {
			t.Errorf("expected data: %s, got %s", expected, res.Data)
		}
	})

	t.Run("returns 200 for non existing users and sends no email", func(t *testing.T) {
		svc, _, emailSenderMock := setupServices(t)

		handler := handlers.AuthResetPasswordRequest(svc)

		body := `{"email":"test@example.com"}`
		req := httptest.NewRequest("POST", "/auth/reset-password-request", strings.NewReader(body))
		rec := httptest.NewRecorder()

		handler.ServeHTTP(rec, req)

		if expected := 200; rec.Code != http.StatusOK {
			t.Errorf("expected %d, got %d", expected, rec.Code)
		}

		if emailSenderMock.called {
			t.Error("expected email not to be sent")
		}

		var res handlers.AuthResetPasswordRequestResponse
		if err := json.NewDecoder(rec.Body).Decode(&res); err != nil {
			t.Fatalf("failed to decode response: %v", err)
		}

		if expected := "ok"; res.Status != expected {
			t.Errorf("expected status: %s, got %s", expected, res.Status)
		}

		if expected := "ok"; res.Data != expected {
			t.Errorf("expected data: %s, got %s", expected, res.Data)
		}
	})

	t.Run("rate limiter hit for fresh token", func(t *testing.T) {
		svc, _, _ := setupServices(t)

		// create user, then set a fresh password reset token
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
			PasswordResetToken:          uuid.NullUUID{Valid: true, UUID: uuid.New()},
			PasswordResetTokenExpiresAt: &exp,
		})

		handler := handlers.AuthResetPasswordRequest(svc)

		body := `{"email":"test@example.com"}`
		req := httptest.NewRequest("POST", "/auth/reset-password-request", strings.NewReader(body))
		rec := httptest.NewRecorder()

		handler.ServeHTTP(rec, req)

		if expected := 429; rec.Code != http.StatusTooManyRequests {
			t.Errorf("expected %d, got %d", expected, rec.Code)
		}

		var res response.ErrorResponse
		if err := json.NewDecoder(rec.Body).Decode(&res); err != nil {
			t.Fatalf("failed to decode response: %v", err)
		}

		if expected := "error"; res.Status != expected {
			t.Errorf("expected status: %s, got %s", expected, res.Status)
		}

		if expected := "password reset email already sent, please wait before requesting a new one"; res.Data != expected {
			t.Errorf("expected data: %s, got %s", expected, res.Data)
		}
	})

	t.Run("success", func(t *testing.T) {
		svc, _, emailSenderMock := setupServices(t)

		// create a user with a password reset token that expires in 50 minutes
		// (in contrast to 1 hr), so there is more than 5 minutes since the token
		// was generated, so the rate limiter won't block us
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

		exp := time.Now().Add(time.Minute * 50)
		svc.Repo.UpdatePasswordResetToken(context.Background(), user.UpdatePasswordResetTokenParams{
			ID:                          u.ID,
			PasswordResetToken:          uuid.NullUUID{Valid: true, UUID: uuid.New()},
			PasswordResetTokenExpiresAt: &exp,
		})

		handler := handlers.AuthResetPasswordRequest(svc)

		body := `{"email":"test@example.com"}`
		req := httptest.NewRequest("POST", "/auth/reset-password-request", strings.NewReader(body))
		rec := httptest.NewRecorder()

		handler.ServeHTTP(rec, req)

		if expected := 200; rec.Code != http.StatusOK {
			t.Errorf("expected %d, got %d", expected, rec.Code)
		}

		if !emailSenderMock.called {
			t.Error("expected email to be sent")
		}

		if !slices.Contains(emailSenderMock.params.To, "test@example.com") {
			t.Error("email sent to wrong recipient")
		}

		var res handlers.AuthResetPasswordRequestResponse
		if err := json.NewDecoder(rec.Body).Decode(&res); err != nil {
			t.Fatalf("failed to decode response: %v", err)
		}

		if expected := "ok"; res.Status != expected {
			t.Errorf("expected status: %s, got %s", expected, res.Status)
		}

		if expected := "ok"; res.Data != expected {
			t.Errorf("expected data: %s, got %s", expected, res.Data)
		}
	})
}

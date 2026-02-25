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
	"github.com/hreftools/api/internal/handlers"
	"github.com/hreftools/api/internal/response"
	"github.com/hreftools/api/internal/store"
)

func TestAuthResendVerificationBody_Normalize(t *testing.T) {
	tests := []struct {
		name     string
		input    handlers.AuthResendVerificationBody
		expected handlers.AuthResendVerificationBody
	}{
		{
			name: "Produces no change to already normalized input",
			input: handlers.AuthResendVerificationBody{
				Email: "user@email.com",
			},
			expected: handlers.AuthResendVerificationBody{
				Email: "user@email.com",
			},
		},
		{
			name: "Trim and lowercase email",
			input: handlers.AuthResendVerificationBody{
				Email: "  User@Email.com  ",
			},
			expected: handlers.AuthResendVerificationBody{
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

func TestAuthResendVerificationBody_Validate(t *testing.T) {
	tests := []struct {
		name       string
		input      handlers.AuthResendVerificationBody
		wantErr    bool
		wantErrMsg string
	}{
		{
			name: "Valid input",
			input: handlers.AuthResendVerificationBody{
				Email: "valid@email.com",
			},
			wantErr:    false,
			wantErrMsg: "",
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
				if gotErr.Error() != tt.wantErrMsg {
					t.Errorf("Validate() error = %q, want %q", gotErr.Error(), tt.wantErrMsg)
				}
				return
			}
			if tt.wantErr {
				t.Fatal("Validate() succeeded unexpectedly")
			}
		})
	}
}

func TestAuthResendVerification(t *testing.T) {
	t.Run("fails on incorrect body", func(t *testing.T) {
		s := setupTestStore(t)
		emailSenderMock := &mockEmailSender{}

		handler := handlers.AuthResendVerification(s, emailSenderMock)

		body := `this is not a json body`
		req := httptest.NewRequest("POST", "/auth/resend-verification", strings.NewReader(body))
		rec := httptest.NewRecorder()

		handler.ServeHTTP(rec, req)

		if rec.Code != http.StatusBadRequest {
			t.Errorf("expected 400, got %d", rec.Code)
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
		s := setupTestStore(t)
		emailSenderMock := &mockEmailSender{}

		handler := handlers.AuthResendVerification(s, emailSenderMock)

		body := `{"email":"test@example.com","unexpected":"field"}`
		req := httptest.NewRequest("POST", "/auth/resend-verification", strings.NewReader(body))
		rec := httptest.NewRecorder()

		handler.ServeHTTP(rec, req)

		if rec.Code != http.StatusBadRequest {
			t.Errorf("expected 400, got %d", rec.Code)
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
		s := setupTestStore(t)
		emailSenderMock := &mockEmailSender{}

		handler := handlers.AuthResendVerification(s, emailSenderMock)

		body := `{"email":""}`

		req := httptest.NewRequest("POST", "/auth/resend-verification", strings.NewReader(body))
		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, req)

		if rec.Code != http.StatusBadRequest {
			t.Errorf("expected 400, got %d", rec.Code)
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
		s := setupTestStore(t)
		emailSenderMock := &mockEmailSender{}

		handler := handlers.AuthResendVerification(s, emailSenderMock)

		body := `{"email":"test@example.com"}`
		req := httptest.NewRequest("POST", "/auth/resend-verification", strings.NewReader(body))
		rec := httptest.NewRecorder()

		handler.ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Errorf("expected 200, got %d", rec.Code)
		}

		if emailSenderMock.called {
			t.Error("expected email not to be sent")
		}

		var res handlers.AuthResendVerificationResponse
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

	t.Run("returns 200 for already verified users and sends no email", func(t *testing.T) {
		s := setupTestStore(t)
		emailSenderMock := &mockEmailSender{}

		s.Users.Create(context.Background(), store.UserCreateParams{
			Email:                           "test@example.com",
			EmailVerified:                   true,
			EmailVerificationToken:          uuid.NullUUID{},
			EmailVerificationTokenExpiresAt: nil,
			Password:                        "strongpassword",
			Username:                        "testuser",
			IsAdmin:                         false,
			IsPro:                           false,
		})

		handler := handlers.AuthResendVerification(s, emailSenderMock)

		body := `{"email":"test@example.com"}`
		req := httptest.NewRequest("POST", "/auth/resend-verification", strings.NewReader(body))
		rec := httptest.NewRecorder()

		handler.ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Errorf("expected 200, got %d", rec.Code)
		}

		if emailSenderMock.called {
			t.Error("expected email not to be sent")
		}

		var res handlers.AuthResendVerificationResponse
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

	t.Run("rate limier hit for fresh token", func(t *testing.T) {
		s := setupTestStore(t)
		emailSenderMock := &mockEmailSender{}

		s.Users.Create(context.Background(), store.UserCreateParams{
			Email:                           "test@example.com",
			EmailVerified:                   false,
			EmailVerificationToken:          uuid.NullUUID{Valid: true, UUID: uuid.New()},
			EmailVerificationTokenExpiresAt: new(time.Now().Add(handlers.TokenExpiryDuration)),
			Password:                        "strongpassword",
			Username:                        "testuser",
			IsAdmin:                         false,
			IsPro:                           false,
		})

		handler := handlers.AuthResendVerification(s, emailSenderMock)

		body := `{"email":"test@example.com"}`
		req := httptest.NewRequest("POST", "/auth/resend-verification", strings.NewReader(body))
		rec := httptest.NewRecorder()

		handler.ServeHTTP(rec, req)

		if rec.Code != http.StatusTooManyRequests {
			t.Errorf("expected 429, got %d", rec.Code)
		}

		var res response.ErrorResponse
		if err := json.NewDecoder(rec.Body).Decode(&res); err != nil {
			t.Fatalf("failed to decode response: %v", err)
		}

		if expected := "error"; res.Status != expected {
			t.Errorf("expected status: %s, got %s", expected, res.Status)
		}

		if expected := "verification email already sent, please wait before requesting a new one"; res.Data != expected {
			t.Errorf("expected data: %s, got %s", expected, res.Data)
		}

	})

	t.Run("success", func(t *testing.T) {
		s := setupTestStore(t)
		emailSenderMock := &mockEmailSender{}

		// create a user with token that expires in 23 hrs (in contrast to 24 hr),
		// so there is more than 5 minutes since the token was generated,
		// so we will be able to test the resend verification flow
		// as the token is not longer super fresh
		s.Users.Create(context.Background(), store.UserCreateParams{
			Email:                           "test@example.com",
			EmailVerified:                   false,
			EmailVerificationToken:          uuid.NullUUID{Valid: true, UUID: uuid.New()},
			EmailVerificationTokenExpiresAt: new(time.Now().Add(time.Hour * 23)),
			Password:                        "strongpassword",
			Username:                        "testuser",
			IsAdmin:                         false,
			IsPro:                           false,
		})

		handler := handlers.AuthResendVerification(s, emailSenderMock)

		body := `{"email":"test@example.com"}`
		req := httptest.NewRequest("POST", "/auth/resend-verification", strings.NewReader(body))
		rec := httptest.NewRecorder()

		handler.ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Errorf("expected 200, got %d", rec.Code)
		}

		if !emailSenderMock.called {
			t.Error("expected email to be sent")
		}

		if !slices.Contains(emailSenderMock.params.To, "test@example.com") {
			t.Error("email sent to wrong recipient")
		}

		var res handlers.AuthResendVerificationResponse
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

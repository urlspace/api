package handlers_test

import (
	"encoding/json"
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
	"github.com/hreftools/api/internal/utils"
)

func seedSigninUser(t *testing.T, svc *user.Service, email, username, password string, verified bool) {
	t.Helper()

	hash, err := utils.PasswordHash(password)
	if err != nil {
		t.Fatalf("failed to hash password: %v", err)
	}

	var token uuid.NullUUID
	var expiresAt *time.Time
	if !verified {
		token = uuid.NullUUID{Valid: true, UUID: uuid.New()}
		exp := time.Now().Add(24 * time.Hour)
		expiresAt = &exp
	}

	_, err = svc.Repo.Create(t.Context(), user.CreateParams{
		Email:                           email,
		EmailVerified:                   verified,
		EmailVerificationToken:          token,
		EmailVerificationTokenExpiresAt: expiresAt,
		Password:                        hash,
		Username:                        username,
		IsAdmin:                         false,
		IsPro:                           false,
	})
	if err != nil {
		t.Fatalf("failed to seed user: %v", err)
	}
}

func TestAuthSigninBody_Normalize(t *testing.T) {
	tests := []struct {
		name     string
		input    handlers.AuthSigninBody
		expected handlers.AuthSigninBody
	}{
		{
			name: "Produces no change to already normalized input",
			input: handlers.AuthSigninBody{
				Email:    "user@email.com",
				Password: "  whateva  ",
			},
			expected: handlers.AuthSigninBody{
				Email:    "user@email.com",
				Password: "  whateva  ",
			},
		},
		{
			name: "Trim and lowercase email",
			input: handlers.AuthSigninBody{
				Email:    "  User@Email.com  ",
				Password: "  whateva  ",
			},
			expected: handlers.AuthSigninBody{
				Email:    "user@email.com",
				Password: "  whateva  ",
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
			if b.Password != tt.expected.Password {
				t.Errorf("Normalize() Password = %v, want %v", b.Password, tt.expected.Password)
			}
		})
	}
}

func TestAuthSigninBody_Validate(t *testing.T) {
	tests := []struct {
		name    string
		input   handlers.AuthSigninBody
		wantErr bool
	}{
		{
			name: "Valid input",
			input: handlers.AuthSigninBody{
				Email:    "valid@email.com",
				Password: "strongpassword",
			},
			wantErr: false,
		},
		{
			name: "Invalid email",
			input: handlers.AuthSigninBody{
				Email:    "notanemail",
				Password: "strongpassword",
			},
			wantErr: true,
		},
		{
			name: "Empty password",
			input: handlers.AuthSigninBody{
				Email:    "valid@email.com",
				Password: "",
			},
			wantErr: true,
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

func TestAuthSignin(t *testing.T) {
	t.Run("fails on incorrect body", func(t *testing.T) {
		svc, _, _ := setupServices(t)
		handler := handlers.AuthSignin(svc)

		req := httptest.NewRequest("POST", "/auth/signin", strings.NewReader("not json"))
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
			t.Errorf("expected status %s, got %s", expected, res.Status)
		}
		if expected := "invalid request body"; res.Data != expected {
			t.Errorf("expected data %s, got %s", expected, res.Data)
		}
	})

	t.Run("fails on unexpected field in body", func(t *testing.T) {
		svc, _, _ := setupServices(t)
		handler := handlers.AuthSignin(svc)

		body := `{"email":"test@example.com","password":"SecurePass123!","unexpected":"field"}`
		req := httptest.NewRequest("POST", "/auth/signin", strings.NewReader(body))
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
			t.Errorf("expected status %s, got %s", expected, res.Status)
		}
		if expected := "invalid request body"; res.Data != expected {
			t.Errorf("expected data %s, got %s", expected, res.Data)
		}
	})

	t.Run("fails on invalid request body", func(t *testing.T) {
		svc, _, _ := setupServices(t)
		handler := handlers.AuthSignin(svc)

		body := `{"email":"","password":""}`
		req := httptest.NewRequest("POST", "/auth/signin", strings.NewReader(body))
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
			t.Errorf("expected status %s, got %s", expected, res.Status)
		}
		if expected := "email is required"; res.Data != expected {
			t.Errorf("expected data %s, got %s", expected, res.Data)
		}
	})

	t.Run("fails on non-existent email", func(t *testing.T) {
		svc, _, _ := setupServices(t)
		handler := handlers.AuthSignin(svc)

		body := `{"email":"nobody@example.com","password":"SecurePass123!"}`
		req := httptest.NewRequest("POST", "/auth/signin", strings.NewReader(body))
		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, req)

		if expected := 401; rec.Code != http.StatusUnauthorized {
			t.Errorf("expected %d, got %d", expected, rec.Code)
		}

		var res response.ErrorResponse
		if err := json.NewDecoder(rec.Body).Decode(&res); err != nil {
			t.Fatalf("failed to decode response: %v", err)
		}
		if expected := "error"; res.Status != expected {
			t.Errorf("expected status %s, got %s", expected, res.Status)
		}
		if expected := "invalid email or password"; res.Data != expected {
			t.Errorf("expected data %s, got %s", expected, res.Data)
		}
	})

	t.Run("fails on wrong password", func(t *testing.T) {
		svc, _, _ := setupServices(t)
		seedSigninUser(t, svc, "user@example.com", "user", "SecurePass123!", true)
		handler := handlers.AuthSignin(svc)

		body := `{"email":"user@example.com","password":"WrongPassword!"}`
		req := httptest.NewRequest("POST", "/auth/signin", strings.NewReader(body))
		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, req)

		if expected := 401; rec.Code != http.StatusUnauthorized {
			t.Errorf("expected %d, got %d", expected, rec.Code)
		}

		var res response.ErrorResponse
		if err := json.NewDecoder(rec.Body).Decode(&res); err != nil {
			t.Fatalf("failed to decode response: %v", err)
		}
		if expected := "error"; res.Status != expected {
			t.Errorf("expected status %s, got %s", expected, res.Status)
		}
		if expected := "invalid email or password"; res.Data != expected {
			t.Errorf("expected data %s, got %s", expected, res.Data)
		}
	})

	t.Run("fails on unverified email", func(t *testing.T) {
		svc, _, _ := setupServices(t)
		seedSigninUser(t, svc, "unverified@example.com", "unverified", "SecurePass123!", false)
		handler := handlers.AuthSignin(svc)

		body := `{"email":"unverified@example.com","password":"SecurePass123!"}`
		req := httptest.NewRequest("POST", "/auth/signin", strings.NewReader(body))
		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, req)

		if expected := 401; rec.Code != http.StatusUnauthorized {
			t.Errorf("expected %d, got %d", expected, rec.Code)
		}

		var res response.ErrorResponse
		if err := json.NewDecoder(rec.Body).Decode(&res); err != nil {
			t.Fatalf("failed to decode response: %v", err)
		}
		if expected := "error"; res.Status != expected {
			t.Errorf("expected status %s, got %s", expected, res.Status)
		}
		if expected := "invalid email or password"; res.Data != expected {
			t.Errorf("expected data %s, got %s", expected, res.Data)
		}
	})

	t.Run("success", func(t *testing.T) {
		svc, _, _ := setupServices(t)
		seedSigninUser(t, svc, "verified@example.com", "verified", "SecurePass123!", true)
		handler := handlers.AuthSignin(svc)

		body := `{"email":"verified@example.com","password":"SecurePass123!"}`
		req := httptest.NewRequest("POST", "/auth/signin", strings.NewReader(body))
		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, req)

		if expected := 200; rec.Code != http.StatusOK {
			t.Errorf("expected %d, got %d", expected, rec.Code)
		}

		var res handlers.AuthSigninResponse
		if err := json.NewDecoder(rec.Body).Decode(&res); err != nil {
			t.Fatalf("failed to decode response: %v", err)
		}
		if expected := "ok"; res.Status != expected {
			t.Errorf("expected status %s, got %s", expected, res.Status)
		}
		if expected := "ok"; res.Data != expected {
			t.Errorf("expected data %s, got %s", expected, res.Data)
		}

		cookies := rec.Result().Cookies()
		var sessionCookie *http.Cookie
		for _, c := range cookies {
			if c.Name == config.SessionCookieName {
				sessionCookie = c
				break
			}
		}

		if sessionCookie == nil {
			t.Fatal("expected session_id cookie to be set")
		}
		if _, err := uuid.Parse(sessionCookie.Value); err != nil {
			t.Errorf("expected session_id to be a UUID, got %s", sessionCookie.Value)
		}
		if !sessionCookie.HttpOnly {
			t.Error("expected session_id cookie to be HttpOnly")
		}
		if !sessionCookie.Secure {
			t.Error("expected session_id cookie to be Secure")
		}
		if expected := time.Now().Add(29 * 24 * time.Hour); sessionCookie.Expires.Before(expected) {
			t.Errorf("expected session_id cookie to expire in ~30 days, got %v", sessionCookie.Expires)
		}
	})
}

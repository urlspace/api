package handlers_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"slices"
	"strings"
	"testing"

	"github.com/hreftools/api/internal/handlers"
	"github.com/hreftools/api/internal/response"
)

func TestAuthSignupBody_Normalize(t *testing.T) {
	tests := []struct {
		name     string
		input    handlers.AuthSignupBody
		expected handlers.AuthSignupBody
	}{
		{
			name: "Produces no change to already normalized input",
			input: handlers.AuthSignupBody{
				Username: "user_name",
				Email:    "user@email.com",
				Password: "  whateva  ",
			},
			expected: handlers.AuthSignupBody{
				Username: "user_name",
				Email:    "user@email.com",
				Password: "  whateva  ",
			},
		},
		{
			name: "Trim and lowercase username and email",
			input: handlers.AuthSignupBody{
				Username: "  User_Name  ",
				Email:    "  User@Email.com  ",
				Password: "  whateva  ",
			},
			expected: handlers.AuthSignupBody{
				Username: "user_name",
				Email:    "user@email.com",
				Password: "  whateva  ",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			b := tt.input
			b.Normalize()

			if b.Username != tt.expected.Username {
				t.Errorf("Normalize() Username = %v, want %v", b.Username, tt.expected.Username)
			}
			if b.Email != tt.expected.Email {
				t.Errorf("Normalize() Email = %v, want %v", b.Email, tt.expected.Email)
			}
			if b.Password != tt.expected.Password {
				t.Errorf("Normalize() Password = %v, want %v", b.Password, tt.expected.Password)
			}
		})
	}
}

func TestAuthSignupBody_Validate(t *testing.T) {
	tests := []struct {
		name    string
		input   handlers.AuthSignupBody
		wantErr bool
	}{
		{
			name: "Valid input",
			input: handlers.AuthSignupBody{
				Username: "valid_user",
				Email:    "valid@email.com",
				Password: "strongpassword",
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

func TestAuthSignup(t *testing.T) {
	t.Run("fails on unexpected field in body", func(t *testing.T) {
		s := setupTestStore(t)
		emailSenderMock := &mockEmailSender{}

		handler := handlers.AuthSignup(s, emailSenderMock)

		body := `{"username":"testuser","email":"test@example.com","password":"SecurePass123!","unexpected":"field"}`
		req := httptest.NewRequest("POST", "/auth/signup", strings.NewReader(body))
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

	t.Run("fails on incorrect body", func(t *testing.T) {
		s := setupTestStore(t)
		emailSenderMock := &mockEmailSender{}

		handler := handlers.AuthSignup(s, emailSenderMock)

		body := `this is not a json body`
		req := httptest.NewRequest("POST", "/auth/signup", strings.NewReader(body))
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

		handler := handlers.AuthSignup(s, emailSenderMock)

		body := `{"username":"","email":"","password":""}`

		req := httptest.NewRequest("POST", "/auth/signup", strings.NewReader(body))
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

		if expected := "username is required"; res.Data != expected {
			t.Errorf("expected data: %s, got %s", expected, res.Data)
		}
	})

	t.Run("returns conflict status on duplicated emails ", func(t *testing.T) {
		s := setupTestStore(t)
		emailSenderMock := &mockEmailSender{}

		handler := handlers.AuthSignup(s, emailSenderMock)

		body1 := `{"username":"one","email":"test@example.com","password":"SecurePass123!"}`
		req1 := httptest.NewRequest("POST", "/auth/signup", strings.NewReader(body1))
		rec1 := httptest.NewRecorder()
		handler.ServeHTTP(rec1, req1)

		body2 := `{"username":"two","email":"test@example.com","password":"SecurePass123!"}`
		req2 := httptest.NewRequest("POST", "/auth/signup", strings.NewReader(body2))
		rec2 := httptest.NewRecorder()
		handler.ServeHTTP(rec2, req2)

		if rec1.Code != http.StatusCreated {
			t.Errorf("expected 201, got %d", rec1.Code)
		}

		if rec2.Code != http.StatusConflict {
			t.Errorf("expected 409, got %d", rec2.Code)
		}

		var res2 response.ErrorResponse
		if err := json.NewDecoder(rec2.Body).Decode(&res2); err != nil {
			t.Fatalf("failed to decode response: %v", err)
		}

		if expected := "error"; res2.Status != expected {
			t.Errorf("expected status: %s, got %s", expected, res2.Status)
		}

		if expected := "request conflict"; res2.Data != expected {
			t.Errorf("expected data: %s, got %s", expected, res2.Data)
		}
	})

	t.Run("returns conflict status on duplicated usernames ", func(t *testing.T) {
		s := setupTestStore(t)
		emailSenderMock := &mockEmailSender{}

		handler := handlers.AuthSignup(s, emailSenderMock)

		body1 := `{"username":"test","email":"one@example.com","password":"SecurePass123!"}`
		req1 := httptest.NewRequest("POST", "/auth/signup", strings.NewReader(body1))
		rec1 := httptest.NewRecorder()
		handler.ServeHTTP(rec1, req1)

		body2 := `{"username":"test","email":"two@example.com","password":"SecurePass123!"}`
		req2 := httptest.NewRequest("POST", "/auth/signup", strings.NewReader(body2))
		rec2 := httptest.NewRecorder()
		handler.ServeHTTP(rec2, req2)

		if rec1.Code != http.StatusCreated {
			t.Errorf("expected 201, got %d", rec1.Code)
		}

		if rec2.Code != http.StatusConflict {
			t.Errorf("expected 409, got %d", rec2.Code)
		}

		var res1 handlers.AuthSignupResponse
		if err := json.NewDecoder(rec1.Body).Decode(&res1); err != nil {
			t.Fatalf("failed to decode response: %v", err)
		}

		if expected := "ok"; res1.Status != expected {
			t.Errorf("expected status: %s, got %s", expected, res1.Status)
		}

		if expected := "test"; res1.Data.Username != expected {
			t.Errorf("expected data: %s, got %s", expected, res1.Data.Username)
		}

		if expected := "one@example.com"; res1.Data.Email != expected {
			t.Errorf("expected email %s got %s", expected, res1.Data.Email)
		}

		if expected := false; res1.Data.EmailVerified != expected {
			t.Errorf("expected emailVerified %v, got %v", expected, res1.Data.EmailVerified)
		}

		if expected := false; res1.Data.IsAdmin != expected {
			t.Errorf("expected isAdmin %v, got %v", expected, res1.Data.IsAdmin)
		}

		if expected := false; res1.Data.IsPro != expected {
			t.Errorf("expected isPro %v, got %v", expected, res1.Data.IsPro)
		}

		var res2 response.ErrorResponse
		if err := json.NewDecoder(rec2.Body).Decode(&res2); err != nil {
			t.Fatalf("failed to decode response: %v", err)
		}

		if expected := "error"; res2.Status != expected {
			t.Errorf("expected status: %s, got %s", expected, res2.Status)
		}

		if expected := "request conflict"; res2.Data != expected {
			t.Errorf("expected data: %s, got %s", expected, res2.Data)
		}
	})

	t.Run("success", func(t *testing.T) {
		s := setupTestStore(t)
		emailSenderMock := &mockEmailSender{}

		handler := handlers.AuthSignup(s, emailSenderMock)

		body := `{"username":"testuser","email":"test@example.com","password":"SecurePass123!"}`
		req := httptest.NewRequest("POST", "/auth/signup", strings.NewReader(body))
		rec := httptest.NewRecorder()

		handler.ServeHTTP(rec, req)

		if rec.Code != http.StatusCreated {
			t.Errorf("expected 201, got %d", rec.Code)
		}

		if !emailSenderMock.called {
			t.Error("expected email to be sent")
		}

		if !slices.Contains(emailSenderMock.params.To, "test@example.com") {
			t.Error("email sent to wrong recipient")
		}

		var res handlers.AuthSignupResponse
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
			t.Errorf("expected email %s got %s", expected, res.Data.Email)
		}

		if expected := false; res.Data.EmailVerified != expected {
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

package handlers_test

import (
	"database/sql"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/hreftools/api/internal/config"
	"github.com/hreftools/api/internal/handlers"
	"github.com/hreftools/api/internal/store"
	"github.com/hreftools/api/internal/utils"
)

func seedSignoutUser(t *testing.T, s *store.Store) uuid.UUID {
	t.Helper()

	hash, err := utils.PasswordHash("SecurePass123!")
	if err != nil {
		t.Fatalf("failed to hash password: %v", err)
	}

	user, err := s.Users.Create(t.Context(), store.UserCreateParams{
		Email:         "signout@example.com",
		EmailVerified: true,
		Password:      hash,
		Username:      "signoutuser",
		IsAdmin:       false,
		IsPro:         false,
	})
	if err != nil {
		t.Fatalf("failed to seed user: %v", err)
	}

	return user.ID
}

func TestAuthSignout(t *testing.T) {
	t.Run("success via cookie", func(t *testing.T) {
		s := setupTestStore(t)
		userID := seedSignoutUser(t, s)

		token, err := s.Tokens.Create(t.Context(), store.TokenCreateParams{
			UserID:    userID,
			Type:      config.TokenTypeSession,
			ExpiresAt: time.Now().Add(config.SessionExpiryDuration),
		})
		if err != nil {
			t.Fatalf("failed to create token: %v", err)
		}

		handler := handlers.AuthSignout(s)
		req := httptest.NewRequest("POST", "/auth/signout", nil)
		req.AddCookie(&http.Cookie{Name: config.SessionCookieName, Value: token.ID.String()})
		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, req)

		if expected := http.StatusOK; rec.Code != expected {
			t.Errorf("expected %d, got %d", expected, rec.Code)
		}

		var res handlers.AuthSignoutResponse
		if err := json.NewDecoder(rec.Body).Decode(&res); err != nil {
			t.Fatalf("failed to decode response: %v", err)
		}
		if expected := "ok"; res.Status != expected {
			t.Errorf("expected status: %s, got %s", expected, res.Status)
		}
		if expected := "signed out"; res.Data != expected {
			t.Errorf("expected data: %s, got %s", expected, res.Data)
		}

		var sessionCookie *http.Cookie
		for _, c := range rec.Result().Cookies() {
			if c.Name == config.SessionCookieName {
				sessionCookie = c
				break
			}
		}
		if sessionCookie == nil {
			t.Fatal("expected session_id cookie in response")
		}
		if expected := -1; sessionCookie.MaxAge != expected {
			t.Errorf("expected MaxAge: %d, got %d", expected, sessionCookie.MaxAge)
		}

		_, err = s.Tokens.GetByID(t.Context(), token.ID)
		if !errors.Is(err, sql.ErrNoRows) {
			t.Errorf("expected token to be deleted, got err: %v", err)
		}
	})

	t.Run("success via Bearer header", func(t *testing.T) {
		s := setupTestStore(t)
		userID := seedSignoutUser(t, s)

		token, err := s.Tokens.Create(t.Context(), store.TokenCreateParams{
			UserID:    userID,
			Type:      config.TokenTypeSession,
			ExpiresAt: time.Now().Add(config.SessionExpiryDuration),
		})
		if err != nil {
			t.Fatalf("failed to create token: %v", err)
		}

		handler := handlers.AuthSignout(s)
		req := httptest.NewRequest("POST", "/auth/signout", nil)
		req.Header.Set("Authorization", "Bearer "+token.ID.String())
		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, req)

		if expected := http.StatusOK; rec.Code != expected {
			t.Errorf("expected %d, got %d", expected, rec.Code)
		}

		var res handlers.AuthSignoutResponse
		if err := json.NewDecoder(rec.Body).Decode(&res); err != nil {
			t.Fatalf("failed to decode response: %v", err)
		}
		if expected := "ok"; res.Status != expected {
			t.Errorf("expected status: %s, got %s", expected, res.Status)
		}
		if expected := "signed out"; res.Data != expected {
			t.Errorf("expected data: %s, got %s", expected, res.Data)
		}

		var sessionCookie *http.Cookie
		for _, c := range rec.Result().Cookies() {
			if c.Name == config.SessionCookieName {
				sessionCookie = c
				break
			}
		}
		if sessionCookie == nil {
			t.Fatal("expected session_id cookie in response")
		}
		if expected := -1; sessionCookie.MaxAge != expected {
			t.Errorf("expected MaxAge: %d, got %d", expected, sessionCookie.MaxAge)
		}

		_, err = s.Tokens.GetByID(t.Context(), token.ID)
		if !errors.Is(err, sql.ErrNoRows) {
			t.Errorf("expected token to be deleted, got err: %v", err)
		}
	})
}

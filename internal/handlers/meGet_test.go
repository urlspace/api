package handlers_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/hreftools/api/internal/config"
	"github.com/hreftools/api/internal/handlers"
	"github.com/hreftools/api/internal/store"
	"github.com/hreftools/api/internal/utils"
)

func seedMeUser(t *testing.T, s *store.Store) store.UserCreateParams {
	t.Helper()

	hash, err := utils.PasswordHash("SecurePass123!")
	if err != nil {
		t.Fatalf("failed to hash password: %v", err)
	}

	return store.UserCreateParams{
		Email:         "me@example.com",
		EmailVerified: true,
		Password:      hash,
		Username:      "meuser",
		IsAdmin:       false,
		IsPro:         true,
	}
}

func TestMeGet(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		s := setupTestStore(t)
		params := seedMeUser(t, s)

		user, err := s.Users.Create(t.Context(), params)
		if err != nil {
			t.Fatalf("failed to seed user: %v", err)
		}

		handler := handlers.MeGet(s)
		req := httptest.NewRequest(http.MethodGet, "/me", nil)
		ctx := context.WithValue(req.Context(), config.UserIDContextKey, user.ID)
		req = req.WithContext(ctx)
		rec := httptest.NewRecorder()

		handler.ServeHTTP(rec, req)

		if expected := http.StatusOK; rec.Code != expected {
			t.Errorf("expected %d, got %d", expected, rec.Code)
		}

		var res handlers.MeGetResponse
		if err := json.NewDecoder(rec.Body).Decode(&res); err != nil {
			t.Fatalf("failed to decode response: %v", err)
		}

		if expected := "ok"; res.Status != expected {
			t.Errorf("expected status: %s, got %s", expected, res.Status)
		}
		if expected := user.ID; res.Data.ID != expected {
			t.Errorf("expected ID: %s, got %s", expected, res.Data.ID)
		}
		if expected := params.Email; res.Data.Email != expected {
			t.Errorf("expected email: %s, got %s", expected, res.Data.Email)
		}
		if expected := params.Username; res.Data.Username != expected {
			t.Errorf("expected username: %s, got %s", expected, res.Data.Username)
		}
		if expected := params.IsPro; res.Data.IsPro != expected {
			t.Errorf("expected isPro: %v, got %v", expected, res.Data.IsPro)
		}
	})
}

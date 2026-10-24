package polar

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"net/http"
	"strconv"
	"testing"
	"time"
)

func signWebhook(key []byte, id string, timestamp string, body []byte) string {
	mac := hmac.New(sha256.New, key)
	mac.Write([]byte(id + "." + timestamp + "."))
	mac.Write(body)
	return "v1," + base64.StdEncoding.EncodeToString(mac.Sum(nil))
}

func TestVerifyWebhook(t *testing.T) {
	key := []byte("test-secret-key")
	client := NewClient("", "", "whsec_"+base64.StdEncoding.EncodeToString(key))

	now := time.Unix(1_800_000_000, 0)
	timestamp := strconv.FormatInt(now.Unix(), 10)
	body := []byte(`{"type":"subscription.created","timestamp":"2027-01-15T10:00:00Z","data":{}}`)
	valid := signWebhook(key, "msg_1", timestamp, body)

	tests := []struct {
		name      string
		id        string
		timestamp string
		signature string
		body      []byte
		wantErr   bool
	}{
		{"valid", "msg_1", timestamp, valid, body, false},
		{"valid among multiple signatures", "msg_1", timestamp, "v1,bm9wZQ== " + valid, body, false},
		{"missing signature", "msg_1", timestamp, "", body, true},
		{"tampered body", "msg_1", timestamp, valid, []byte(`{"type":"subscription.revoked"}`), true},
		{"different id", "msg_2", timestamp, valid, body, true},
		{"wrong key", "msg_1", timestamp, signWebhook([]byte("other"), "msg_1", timestamp, body), body, true},
		{"stale timestamp", "msg_1", strconv.FormatInt(now.Add(-6*time.Minute).Unix(), 10), signWebhook(key, "msg_1", strconv.FormatInt(now.Add(-6*time.Minute).Unix(), 10), body), body, true},
		{"future timestamp", "msg_1", strconv.FormatInt(now.Add(6*time.Minute).Unix(), 10), signWebhook(key, "msg_1", strconv.FormatInt(now.Add(6*time.Minute).Unix(), 10), body), body, true},
		{"unsupported version", "msg_1", timestamp, "v2" + valid[2:], body, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			header := http.Header{}
			header.Set("webhook-id", tt.id)
			header.Set("webhook-timestamp", tt.timestamp)
			header.Set("webhook-signature", tt.signature)

			err := client.verifyWebhook(header, tt.body, now)
			if tt.wantErr {
				if !errors.Is(err, ErrWebhookVerification) {
					t.Fatalf("expected ErrWebhookVerification, got %v", err)
				}
				return
			}
			if err != nil {
				t.Fatalf("expected no error, got %v", err)
			}
		})
	}
}

package polar

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json/jsontext"
	"encoding/json/v2"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"
)

const webhookTolerance = 5 * time.Minute

var (
	ErrWebhookVerification = errors.New("webhook verification failed")
)

type WebhookEvent struct {
	Type      string         `json:"type"`
	Timestamp time.Time      `json:"timestamp"`
	Data      jsontext.Value `json:"data"`
}

func (c *Client) ParseWebhook(header http.Header, body []byte) (WebhookEvent, error) {
	if err := c.verifyWebhook(header, body, time.Now()); err != nil {
		return WebhookEvent{}, err
	}

	var event WebhookEvent
	if err := json.Unmarshal(body, &event); err != nil {
		return WebhookEvent{}, fmt.Errorf("polar webhook: decode event: %w", err)
	}

	return event, nil
}

func (c *Client) verifyWebhook(header http.Header, body []byte, now time.Time) error {
	id := header.Get("webhook-id")
	timestamp := header.Get("webhook-timestamp")
	signatures := header.Get("webhook-signature")
	if id == "" || timestamp == "" || signatures == "" {
		return fmt.Errorf("%w: missing headers", ErrWebhookVerification)
	}

	unix, err := strconv.ParseInt(timestamp, 10, 64)
	if err != nil {
		return fmt.Errorf("%w: invalid timestamp", ErrWebhookVerification)
	}
	sentAt := time.Unix(unix, 0)
	if now.Sub(sentAt) > webhookTolerance || sentAt.Sub(now) > webhookTolerance {
		return fmt.Errorf("%w: timestamp outside tolerance", ErrWebhookVerification)
	}

	key, err := base64.StdEncoding.DecodeString(strings.TrimPrefix(c.webhookSecret, "whsec_"))
	if err != nil {
		return fmt.Errorf("polar webhook: decode secret: %w", err)
	}

	mac := hmac.New(sha256.New, key)
	mac.Write([]byte(id + "." + timestamp + "."))
	mac.Write(body)
	expected := mac.Sum(nil)

	for _, signature := range strings.Fields(signatures) {
		version, encoded, ok := strings.Cut(signature, ",")
		if !ok || version != "v1" {
			continue
		}

		decoded, err := base64.StdEncoding.DecodeString(encoded)
		if err != nil {
			continue
		}

		if hmac.Equal(decoded, expected) {
			return nil
		}
	}

	return fmt.Errorf("%w: signature mismatch", ErrWebhookVerification)
}

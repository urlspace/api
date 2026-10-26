package polar

import (
	"bytes"
	"context"
	"encoding/json/v2"
	"fmt"
	"io"
	"net/http"
)

type CustomerSessionCreateParams struct {
	ExternalCustomerID string `json:"external_customer_id"`
	ReturnURL          string `json:"return_url"`
}

type CustomerSession struct {
	CustomerPortalURL string `json:"customer_portal_url"`
}

func (c *Client) CreateCustomerSession(ctx context.Context, params CustomerSessionCreateParams) (CustomerSession, error) {
	body, err := json.Marshal(params)
	if err != nil {
		return CustomerSession{}, fmt.Errorf("polar create customer session: marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.apiURL+"/v1/customer-sessions/", bytes.NewReader(body))
	if err != nil {
		return CustomerSession{}, fmt.Errorf("polar create customer session: build request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+c.accessToken)
	req.Header.Set("Polar-Version", apiVersion)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	res, err := c.httpClient.Do(req)
	if err != nil {
		return CustomerSession{}, fmt.Errorf("polar create customer session: send request: %w", err)
	}
	defer res.Body.Close()

	if res.StatusCode == http.StatusNotFound {
		return CustomerSession{}, ErrCustomerNotFound
	}
	if res.StatusCode != http.StatusCreated {
		resBody, _ := io.ReadAll(io.LimitReader(res.Body, 4096))
		return CustomerSession{}, fmt.Errorf("polar create customer session: unexpected status %d: %s", res.StatusCode, resBody)
	}

	var session CustomerSession
	if err := json.UnmarshalRead(res.Body, &session); err != nil {
		return CustomerSession{}, fmt.Errorf("polar create customer session: decode response: %w", err)
	}

	return session, nil
}

package polar

import (
	"bytes"
	"context"
	"encoding/json/v2"
	"fmt"
	"io"
	"net/http"
	"time"
)

const apiVersion = "2027-01"

type Client struct {
	httpClient  *http.Client
	apiURL      string
	accessToken string
}

func NewClient(apiURL string, accessToken string) *Client {
	return &Client{
		httpClient:  &http.Client{Timeout: 10 * time.Second},
		apiURL:      apiURL,
		accessToken: accessToken,
	}
}

type CheckoutCreateParams struct {
	Products           []string `json:"products"`
	ExternalCustomerID string   `json:"external_customer_id"`
	CustomerEmail      string   `json:"customer_email"`
	SuccessURL         string   `json:"success_url"`
	ReturnURL          string   `json:"return_url"`
}

type Checkout struct {
	ID  string `json:"id"`
	URL string `json:"url"`
}

func (c *Client) CreateCheckout(ctx context.Context, params CheckoutCreateParams) (Checkout, error) {
	body, err := json.Marshal(params)
	if err != nil {
		return Checkout{}, fmt.Errorf("polar checkout: marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.apiURL+"/v1/checkouts/", bytes.NewReader(body))
	if err != nil {
		return Checkout{}, fmt.Errorf("polar checkout: build request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+c.accessToken)
	req.Header.Set("Polar-Version", apiVersion)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	res, err := c.httpClient.Do(req)
	if err != nil {
		return Checkout{}, fmt.Errorf("polar checkout: send request: %w", err)
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusCreated {
		resBody, _ := io.ReadAll(io.LimitReader(res.Body, 4096))
		return Checkout{}, fmt.Errorf("polar checkout: unexpected status %d: %s", res.StatusCode, resBody)
	}

	var checkout Checkout
	if err := json.UnmarshalRead(res.Body, &checkout); err != nil {
		return Checkout{}, fmt.Errorf("polar checkout: decode response: %w", err)
	}

	return checkout, nil
}

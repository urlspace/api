package polar

import (
	"bytes"
	"context"
	"encoding/json/v2"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
)

var (
	ErrCustomerNotFound = errors.New("customer not found")
)

type Customer struct {
	ID         string  `json:"id"`
	Email      string  `json:"email"`
	ExternalID *string `json:"external_id"`
}

type CustomerCreateParams struct {
	Email      string `json:"email"`
	ExternalID string `json:"external_id"`
}

type CustomerUpdateParams struct {
	Email string `json:"email"`
}

func (c *Client) GetCustomerByExternalID(ctx context.Context, externalID string) (Customer, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.apiURL+"/v1/customers/external/"+url.PathEscape(externalID), nil)
	if err != nil {
		return Customer{}, fmt.Errorf("polar get customer: build request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+c.accessToken)
	req.Header.Set("Polar-Version", apiVersion)
	req.Header.Set("Accept", "application/json")

	res, err := c.httpClient.Do(req)
	if err != nil {
		return Customer{}, fmt.Errorf("polar get customer: send request: %w", err)
	}
	defer res.Body.Close()

	if res.StatusCode == http.StatusNotFound {
		return Customer{}, ErrCustomerNotFound
	}
	if res.StatusCode != http.StatusOK {
		resBody, _ := io.ReadAll(io.LimitReader(res.Body, 4096))
		return Customer{}, fmt.Errorf("polar get customer: unexpected status %d: %s", res.StatusCode, resBody)
	}

	var customer Customer
	if err := json.UnmarshalRead(res.Body, &customer); err != nil {
		return Customer{}, fmt.Errorf("polar get customer: decode response: %w", err)
	}

	return customer, nil
}

func (c *Client) CreateCustomer(ctx context.Context, params CustomerCreateParams) (Customer, error) {
	body, err := json.Marshal(params)
	if err != nil {
		return Customer{}, fmt.Errorf("polar create customer: marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.apiURL+"/v1/customers/", bytes.NewReader(body))
	if err != nil {
		return Customer{}, fmt.Errorf("polar create customer: build request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+c.accessToken)
	req.Header.Set("Polar-Version", apiVersion)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	res, err := c.httpClient.Do(req)
	if err != nil {
		return Customer{}, fmt.Errorf("polar create customer: send request: %w", err)
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusCreated {
		resBody, _ := io.ReadAll(io.LimitReader(res.Body, 4096))
		return Customer{}, fmt.Errorf("polar create customer: unexpected status %d: %s", res.StatusCode, resBody)
	}

	var customer Customer
	if err := json.UnmarshalRead(res.Body, &customer); err != nil {
		return Customer{}, fmt.Errorf("polar create customer: decode response: %w", err)
	}

	return customer, nil
}

func (c *Client) UpdateCustomerByExternalID(ctx context.Context, externalID string, params CustomerUpdateParams) (Customer, error) {
	body, err := json.Marshal(params)
	if err != nil {
		return Customer{}, fmt.Errorf("polar update customer: marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPatch, c.apiURL+"/v1/customers/external/"+url.PathEscape(externalID), bytes.NewReader(body))
	if err != nil {
		return Customer{}, fmt.Errorf("polar update customer: build request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+c.accessToken)
	req.Header.Set("Polar-Version", apiVersion)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	res, err := c.httpClient.Do(req)
	if err != nil {
		return Customer{}, fmt.Errorf("polar update customer: send request: %w", err)
	}
	defer res.Body.Close()

	if res.StatusCode == http.StatusNotFound {
		return Customer{}, ErrCustomerNotFound
	}
	if res.StatusCode != http.StatusOK {
		resBody, _ := io.ReadAll(io.LimitReader(res.Body, 4096))
		return Customer{}, fmt.Errorf("polar update customer: unexpected status %d: %s", res.StatusCode, resBody)
	}

	var customer Customer
	if err := json.UnmarshalRead(res.Body, &customer); err != nil {
		return Customer{}, fmt.Errorf("polar update customer: decode response: %w", err)
	}

	return customer, nil
}

func (c *Client) DeleteCustomerByExternalID(ctx context.Context, externalID string) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodDelete, c.apiURL+"/v1/customers/external/"+url.PathEscape(externalID)+"?anonymize=true", nil)
	if err != nil {
		return fmt.Errorf("polar delete customer: build request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+c.accessToken)
	req.Header.Set("Polar-Version", apiVersion)
	req.Header.Set("Accept", "application/json")

	res, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("polar delete customer: send request: %w", err)
	}
	defer res.Body.Close()

	if res.StatusCode == http.StatusNotFound {
		return ErrCustomerNotFound
	}
	if res.StatusCode != http.StatusNoContent {
		resBody, _ := io.ReadAll(io.LimitReader(res.Body, 4096))
		return fmt.Errorf("polar delete customer: unexpected status %d: %s", res.StatusCode, resBody)
	}

	return nil
}

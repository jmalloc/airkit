package api

import (
	"context"
	"encoding/json"
	"net/http"
	"net/url"
)

// Client is is used to access the MyAir API.
type Client struct {
	// Address is the TCP address of the MyAir "wall mounted touch screen".
	// This is the Android tablet device connected to the air-conditioning unit.
	Address string

	// HTTPClient is the HTTP client used to make requests to the API. If it is
	// nil, http.DefaultClient is used.
	HTTPClient *http.Client
}

// System retrieves information about the entire system.
func (c *Client) System(ctx context.Context) (*System, error) {
	req, err := c.newRequest(
		http.MethodGet,
		"/getSystemData",
		nil,
	)
	if err != nil {
		return nil, err
	}

	res, err := c.do(ctx, req)
	if err != nil {
		return nil, err
	}

	defer res.Body.Close()

	var s System

	if err := json.NewDecoder(res.Body).Decode(&s); err != nil {
		return nil, err
	}

	return &s, nil
}

// do performs an HTTP request.
func (c *Client) do(ctx context.Context, req *http.Request) (*http.Response, error) {
	h := c.HTTPClient
	if h == nil {
		h = http.DefaultClient
	}

	return h.Do(
		req.WithContext(ctx),
	)
}

// newRequest returns a new HTTP request.
func (c *Client) newRequest(
	method string,
	path string,
	query map[string]string,
) (*http.Request, error) {
	u := url.URL{
		Scheme: "http",
		Host:   c.Address,
		Path:   path,
	}

	if len(query) > 0 {
		uv := url.Values{}
		for k, v := range query {
			uv.Set(k, v)
		}
		u.RawQuery = uv.Encode()
	}

	req, err := http.NewRequest(
		method,
		u.String(),
		nil,
	)

	if err != nil {
		return nil, err
	}

	return req, nil
}

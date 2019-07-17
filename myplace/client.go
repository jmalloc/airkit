package myplace

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"time"
)

// DefaultPort is the default port of the API server.
const DefaultPort = "2025"

// A Command is a request to change the state of the system in some way.
type Command func(map[string]*AirCon)

// Client is a client for the MyPlace API.
type Client struct {
	// Host is the hostname of the API server. It must not be empty.
	// The API server runs on the "wall mounted touch screen".
	Host string

	// Port is the TCP port of the HTTP server running on the API server. If it
	// is empty, DefaultPort is used.
	Port string

	// HTTPClient is the HTTP client used to access the API. If it is nil,
	// http.DefaultClient is used.
	HTTPClient *http.Client
}

// Read fetches the state of the entire system.
func (c *Client) Read(ctx context.Context) (*System, error) {
	for {
		res, err := c.get(ctx, "/getSystemData", nil)
		if err != nil {
			return nil, err
		}
		defer res.Body.Close()

		var s System

		if err := json.NewDecoder(res.Body).Decode(&s); err != nil {
			return nil, err
		}

		if s.Details.AppVersion == "" {
			// An empty result was returned. This happens after a successful
			// update for around 4 seconds. This is deliberate behavior of the
			// server API. We keep reading until we see a result, or the
			// deadline is exceeded.
			select {
			case <-ctx.Done():
				return nil, ctx.Err()
			case <-time.After(250 * time.Millisecond):
				continue
			}
		}

		s.populate()

		return &s, nil
	}
}

// Write updates the state of the system by performing one or more commands.
func (c *Client) Write(ctx context.Context, commands ...Command) error {
	req := map[string]*AirCon{}
	for _, c := range commands {
		c(req)
	}

	buf, err := json.Marshal(req)
	if err != nil {
		return err
	}

	fmt.Println(string(buf))

	res, err := c.get(
		ctx,
		"/setAircon",
		url.Values{
			"json": []string{
				string(buf),
			},
		},
	)
	if err != nil {
		return err
	}
	defer res.Body.Close()

	var result struct {
		Ack    bool
		Reason string
	}

	if err := json.NewDecoder(res.Body).Decode(&result); err != nil {
		return err
	}

	if result.Ack {
		return nil
	}

	return errors.New(result.Reason)
}

// get performs an HTTP GET request.
func (c *Client) get(
	ctx context.Context,
	path string,
	query url.Values,
) (*http.Response, error) {
	// build the request URL
	port := c.Port
	if port == "" {
		port = DefaultPort
	}

	u := url.URL{
		Scheme: "http",
		Host:   net.JoinHostPort(c.Host, port),
		Path:   path,
	}

	if query != nil {
		u.RawQuery = query.Encode()
	}

	// construct the GET request
	req, err := http.NewRequest(
		http.MethodGet,
		u.String(),
		nil,
	)
	if err != nil {
		return nil, err
	}

	// make the request
	cli := c.HTTPClient
	if cli == nil {
		cli = http.DefaultClient
	}

	res, err := cli.Do(
		req.WithContext(ctx),
	)

	// unwrap context errors
	if e, ok := err.(*url.Error); ok {
		switch e.Err {
		case context.Canceled,
			context.DeadlineExceeded:
			return res, e.Err
		}
	}

	return res, err
}

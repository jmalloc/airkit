package main

import (
	"errors"
	"os"

	"github.com/jmalloc/airkit/myplace"
)

func init() {
	container.Provide(func() (*myplace.Client, error) {
		host := os.Getenv("AIRKIT_API_HOST")
		if host == "" {
			return nil, errors.New("AIRKIT_API_HOST is not set")
		}

		return &myplace.Client{
			Host: host,
			Port: os.Getenv("AIRKIT_API_PORT"),
		}, nil
	})
}

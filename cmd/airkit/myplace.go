package main

import (
	"github.com/dogmatiq/imbue"
	"github.com/jmalloc/airkit/myplace"
)

func init() {
	imbue.With0(
		container,
		func(
			ctx imbue.Context,
		) (*myplace.Client, error) {
			return &myplace.Client{
				Host: apiHost.Value(),
				Port: apiPort.Value(),
			}, nil
		},
	)
}

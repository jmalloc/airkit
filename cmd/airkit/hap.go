package main

import (
	"github.com/brutella/hap"
	"github.com/dogmatiq/imbue"
)

func init() {
	imbue.With0(
		container,
		func(
			ctx imbue.Context,
		) (hap.Store, error) {
			return hap.NewFsStore("artifacts/db"), nil
		},
	)
}

package main

import (
	"github.com/dogmatiq/ferrite"
	"github.com/jmalloc/airkit/myplace"
)

var (
	apiHost = ferrite.
		String(
			"AIRKIT_API_HOST",
			"the IP address or hostname of the MyAir Touch Panel",
		).
		Required()

	apiPort = ferrite.
		NetworkPort(
			"AIRKIT_API_PORT",
			"the TCP port of the MyAir Touch Panel HTTP server",
		).
		WithDefault(myplace.DefaultPort).
		Required()

	dbPath = ferrite.
		File(
			"AIRKIT_DB_PATH",
			"the path where AirKit stores its data",
		).
		Required()

	homekitPIN = ferrite.
			String(
			"AIRKIT_HOMEKIT_PIN",
			"the PIN code required to pair HomeKit with the AirKit hub",
		).
		WithDefault("12340000").
		Required()
)

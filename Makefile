-include .makefiles/Makefile
-include .makefiles/pkg/go/v1/Makefile

run: artifacts/build/debug/$(GOHOSTOS)/$(GOHOSTARCH)/airkit
	AIRKIT_API_HOST=10.0.100.245 \
	AIRKIT_DB_PATH=artifacts/db \
		$< $(args)

serve: artifacts/build/debug/$(GOHOSTOS)/$(GOHOSTARCH)/airkit
	AIRKIT_API_HOST=10.0.100.245 \
	AIRKIT_DB_PATH=artifacts/db \
		$< serve $(args)

status: artifacts/build/debug/$(GOHOSTOS)/$(GOHOSTARCH)/airkit
	AIRKIT_API_HOST=10.0.100.245 \
	AIRKIT_DB_PATH=artifacts/db \
		$< status $(args)

.makefiles/%:
	curl -sfL https://makefiles.dev/v1 | bash /dev/stdin "$@"

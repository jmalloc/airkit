package main

import (
	"context"
	"math/rand"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/dogmatiq/ferrite"
	"github.com/dogmatiq/imbue"
	"github.com/spf13/cobra"
)

var root = &cobra.Command{
	Use:   "airkit",
	Short: "Integrate MyPlace air-conditioners with Apple HomeKit.",
}

var container = imbue.New()

// version is the current version of the CLI. It is set automatically set during
// build process.
var version = "0.0.0"

func main() {
	rand.Seed(time.Now().UnixNano())
	ferrite.Init()

	ctx, cancel := signal.NotifyContext(
		context.Background(),
		os.Interrupt,
		syscall.SIGTERM,
	)
	defer cancel()

	root.Version = version

	if err := root.ExecuteContext(ctx); err != nil {
		os.Exit(1)
	}
}

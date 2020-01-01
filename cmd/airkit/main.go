package main

import (
	"context"
	"os"

	"github.com/spf13/cobra"
	"go.uber.org/dig"
)

var root = &cobra.Command{
	Use:   "airkit",
	Short: "Integrate MyPlace air-conditioners with Apple HomeKit.",
}

var container = dig.New()

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	root.PersistentPreRunE = func(*cobra.Command, []string) error {
		return container.Provide(func() context.Context {
			return ctx
		})
	}

	if err := root.Execute(); err != nil {
		os.Exit(1)
	}
}

// runE wraps a command run function, causing its arguments to be provided by
// the DI container.
func runE(fn interface{}) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, args []string) error {
		prov := func() (*cobra.Command, []string) {
			return cmd, args
		}

		if err := container.Provide(prov); err != nil {
			return err
		}

		err := container.Invoke(fn)
		return dig.RootCause(err)
	}
}

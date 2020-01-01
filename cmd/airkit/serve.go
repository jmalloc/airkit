package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	dnslog "github.com/brutella/dnssd/log"
	"github.com/brutella/hc"
	"github.com/brutella/hc/accessory"
	"github.com/jmalloc/airkit/manager"
	"github.com/jmalloc/airkit/myplace"
	"github.com/spf13/cobra"
)

func init() {
	cmd := &cobra.Command{
		Use:   "serve",
		Short: "Run the HomeKit accessory server.",
		RunE: runE(func(
			ctx context.Context,
			cmd *cobra.Command,
			cli *myplace.Client,
		) error {
			verbose, err := cmd.Flags().GetBool("verbose")
			if err != nil {
				return err
			}
			if verbose {
				dnslog.Debug.Enable()
			}

			pin := os.Getenv("AIRKIT_PIN")
			fmtPin, err := hc.NewPin(pin)
			if err != nil {
				return fmt.Errorf("AIRKIT_PIN is invalid: %w", err)
			}
			cmd.Printf("HomeKit PIN is %s\n", fmtPin)

			sys, err := readInitialState(ctx, cmd, cli)
			if err != nil {
				return err
			}

			bridge := manager.NewBridge(sys)
			commands := make(chan []myplace.Command, 100)
			var managers []manager.AccessoryManager

			for _, ac := range sys.AirCons {
				cmd.Printf("adding HomeKit accessory for the '%s' air-conditioner\n", ac.Details.Name)
				managers = append(
					managers,
					manager.NewAirConManager(commands, ac),
				)

				managers = append(
					managers,
					manager.NewFanManager(commands, ac),
				)
			}

			var accessories []*accessory.Accessory
			for _, m := range managers {
				accessories = append(accessories, m.Accessories()...)
			}

			cmd.Println("starting HomeKit accessory server")
			xport, err := hc.NewIPTransport(
				hc.Config{
					StoragePath: "artifacts/db",
					Pin:         pin,
				},
				bridge.Accessory,
				accessories...,
			)
			if err != nil {
				return err
			}

			go func() {
				<-ctx.Done()
				xport.Stop()
			}()

			go func() {
				for {
					select {
					case <-ctx.Done():
						return

					case cmds := <-commands:
						err := cli.Write(ctx, cmds...)
						if err != nil {
							log.Print(err)
							continue
						}

					case <-time.After(2 * time.Second):
						s, err := cli.Read(ctx)
						if err != nil {
							log.Print(err)
							continue
						}

						for _, m := range managers {
							m.Update(s)
						}
					}
				}
			}()

			xport.Start()
			return nil
		}),
	}

	cmd.Flags().BoolP("verbose", "v", false, "Enable verbose bonjour logging")

	root.AddCommand(cmd)
}

// readInitialState reads the state of the MyPlace system.
//
// It retries until the state is read successfully or ctx is canceled.
func readInitialState(
	ctx context.Context,
	cmd *cobra.Command,
	cli *myplace.Client,
) (*myplace.System, error) {
	for {
		cmd.Println("reading MyPlace system information")

		sys, err := cli.Read(ctx)
		if err == nil || err == context.Canceled {
			return sys, err
		}

		cmd.PrintErr(err)

		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-time.After(1 * time.Second):
			continue
		}
	}
}

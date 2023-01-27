package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	dnslog "github.com/brutella/dnssd/log"
	"github.com/brutella/hap"
	"github.com/brutella/hap/accessory"
	"github.com/dogmatiq/imbue"
	"github.com/jmalloc/airkit/manager"
	"github.com/jmalloc/airkit/myplace"
	"github.com/spf13/cobra"
)

func init() {
	cmd := &cobra.Command{
		Use:   "serve",
		Short: "Run the HomeKit accessory server.",
		RunE: func(
			cmd *cobra.Command,
			args []string,
		) error {
			verbose, err := cmd.Flags().GetBool("verbose")
			if err != nil {
				return err
			}
			if verbose {
				dnslog.Debug.Enable()
			}

			cmd.SilenceUsage = true

			ctx, cancel := signal.NotifyContext(
				cmd.Context(),
				os.Interrupt,
				syscall.SIGTERM,
			)
			defer cancel()

			return imbue.Invoke2(
				ctx,
				container,
				func(
					ctx context.Context,
					st hap.Store,
					cli *myplace.Client,
				) error {
					sys, err := readInitialState(ctx, cmd, cli)
					if err != nil {
						return err
					}

					bridge := manager.NewBridge(version, sys)
					commands := make(chan []myplace.Command, 100)
					var managers []manager.AccessoryManager

					for _, ac := range sys.AirCons {
						log.Printf("adding HomeKit accessory for the '%s' air-conditioner\n", ac.Details.Name)
						managers = append(
							managers,
							manager.NewAirConManager(st, commands, ac),
						)

						managers = append(
							managers,
							manager.NewFanManager(commands, ac),
						)
					}

					var accessories []*accessory.A
					for _, m := range managers {
						accessories = append(accessories, m.Accessories()...)
					}

					srv, err := hap.NewServer(st, bridge.A, accessories...)
					srv.Pin = homekitPIN.Value()
					if err != nil {
						return err
					}

					go func() {
						for {
							select {
							case <-ctx.Done():
								return

							case cmds := <-commands:
								for _, cmd := range cmds {
									log.Print(cmd)
								}

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

					log.Printf("starting HomeKit accessory server, PIN is %s", srv.Pin)

					err = srv.ListenAndServe(ctx)
					if ctx.Err() != nil {
						return nil
					}

					return err
				},
			)
		},
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

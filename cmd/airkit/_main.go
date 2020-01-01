package main

import (
	"context"
	"log"
	"math/rand"
	"os"
	"time"

	// dnslog "github.com/brutella/dnssd/log"
	"github.com/brutella/hc"
	"github.com/brutella/hc/accessory"
	"github.com/jmalloc/airkit/manager"
	"github.com/jmalloc/airkit/myplace"
)

func main() {
	// dnslog.Debug.Enable()

	rand.Seed(time.Now().UnixNano())

	ctx := context.Background()
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	cli := &myplace.Client{
		Host: os.Getenv("AIRKIT_API_HOST"),
		Port: os.Getenv("AIRKIT_API_PORT"),
	}

	system, err := cli.Read(ctx)
	if err != nil {
		log.Fatal(err)
	}

	bridge := manager.NewBridge(system)

	commands := make(chan myplace.Command)
	var managers []manager.AccessoryManager

	for _, ac := range system.AirCons {
		for _, z := range ac.Zones {
			m := manager.NewZoneManager(commands, ac, z)
			managers = append(managers, m)
		}

		m := manager.NewAirConManager(commands, ac)
		managers = append(managers, m)
	}

	var accessories []*accessory.Accessory
	for _, m := range managers {
		accessories = append(accessories, m.Accessories()...)
	}

	xport, err := hc.NewIPTransport(
		hc.Config{
			StoragePath: "artifacts/db",
			Pin:         os.Getenv("AIRKIT_PIN"),
		},
		bridge.Accessory,
		accessories...,
	)
	if err != nil {
		log.Fatal(err)
	}

	hc.OnTermination(func() {
		cancel()
		<-xport.Stop()
	})

	go func() {
		for {
			select {
			case <-ctx.Done():
				return

			case c := <-commands:
				cmds := []myplace.Command{c}
				debounce := time.NewTimer(250 * time.Millisecond)

			loop:
				for {
					select {
					case c := <-commands:
						cmds = append(cmds, c)
					case <-debounce.C:
						break loop
					}
				}
				debounce.Stop()

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
}

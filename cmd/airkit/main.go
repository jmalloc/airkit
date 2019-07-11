package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/brutella/hc"
	"github.com/brutella/hc/accessory"
	"github.com/brutella/hc/characteristic"
	"github.com/jmalloc/airkit/api"
)

func main() {
	ctx := context.Background()
	cli := api.Client{
		Address: os.Getenv("MYAIR_API_ADDRESS"),
	}

	sys, err := cli.System(ctx)
	if err != nil {
		log.Fatal(err)
	}

	bridge := accessory.NewBridge(
		accessory.Info{
			Name:         "MyAir",
			Manufacturer: "Advantage Air / James Harris",
			Model:        sys.Details.TouchScreenModel,
			SerialNumber: sys.Details.TouchScreenID,
			FirmwareRevision: fmt.Sprintf(
				"MyPlace: %s (%s) / AirKit: %s",
				sys.Details.AppVersion,
				sys.Details.ServiceVersion,
				"0.0.0",
			),
		},
	)

	var accessories []*accessory.Accessory
	thermostats := map[int]*accessory.Thermostat{}

	for _, ac := range sys.AirCons {
		for _, z := range ac.Zones {
			acc := accessory.NewThermostat(
				accessory.Info{
					Name: z.Name + " Thermostat",
				},
				z.CurrentTemp,
				0,
				100,
				0.1,
			)

			if ac.IsOn() && z.IsOn() {
				switch ac.Details.Mode {
				case api.Heat:
					acc.Thermostat.CurrentHeatingCoolingState.SetValue(1)
				case api.Cool:
					acc.Thermostat.CurrentHeatingCoolingState.SetValue(2)
				case api.Vent:
					acc.Thermostat.CurrentHeatingCoolingState.SetValue(3) // reserved, what else do we do?
				case api.Dry:
					acc.Thermostat.CurrentHeatingCoolingState.SetValue(4) // reserved, what else do we do?
				}

			}

			CurrentHeatingCoolingState * characteristic.CurrentHeatingCoolingState
			TargetHeatingCoolingState * characteristic.TargetHeatingCoolingState
			CurrentTemperature * characteristic.CurrentTemperature
			TargetTemperature * characteristic.TargetTemperature
			TemperatureDisplayUnits * characteristic.TemperatureDisplayUnits

			thermostats[z.Number] = thermo
			accessories = append(accessories, acc.Accessory)
		}
	}

	xport, err := hc.NewIPTransport(
		hc.Config{
			StoragePath: "artifacts/db",
			Pin:         "32191111",
		},
		bridge.Accessory,
		accessories...,
	)
	if err != nil {
		log.Fatal(err)
	}

	hc.OnTermination(func() {
		<-xport.Stop()
	})

	xport.Start()
}

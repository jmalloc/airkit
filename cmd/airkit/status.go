package main

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"unicode/utf8"

	"github.com/jmalloc/airkit/myplace"
	"github.com/spf13/cobra"
)

func init() {
	root.AddCommand(&cobra.Command{
		Use:   "status",
		Short: "Print the status of air-conditioning units.",
		RunE: runE(func(
			ctx context.Context,
			cmd *cobra.Command,
			cli *myplace.Client,
		) error {
			sys, err := cli.Read(ctx)
			if err != nil {
				return err
			}

			for _, ac := range sys.AirCons {
				printAC(cmd, ac)
			}

			return nil
		}),
	})
}

func printAC(cmd *cobra.Command, ac *myplace.AirCon) {
	title := fmt.Sprintf(
		"%s (%s)",
		ac.Details.Name,
		ac.ID,
	)

	cmd.Println(title)
	cmd.Println(strings.Repeat("-", len(title)))
	cmd.Println("")

	cmd.Printf("Power:    %s\n", ac.Details.Power)

	if ac.Details.Mode == myplace.AirConModeAuto {
		cmd.Printf("Mode:     %s (%s)", ac.Details.Mode, ac.Details.MyAutoMode)
	} else {
		cmd.Printf("Mode:     %s", ac.Details.Mode)
	}
	if ac.Details.MyTempEnabled {
		cmd.Printf(" [mytemp enabled]")
	}
	if ac.Details.MyAutoEnabled {
		cmd.Printf(" [myauto enabled]")
	}
	if ac.Details.MySleepSaverRunning {
		cmd.Printf(" [mysleep$aver enabled - active]")
	} else if ac.Details.MySleepSaverEnabled {
		cmd.Printf(" [mysleep$aver enabled - inactive]")
	}
	cmd.Println("")

	cmd.Printf("Fan:      %s", ac.Details.FanSpeed)
	if ac.Details.MyFanEnabled {
		cmd.Print(" [myfan enabled]")
	}
	cmd.Println("")

	cmd.Printf("Firmware: v%d.%d\n", ac.Details.FirmwareMajorVersion, ac.Details.FirmwareMinorVersion)
	cmd.Println("")

	pad := zoneNamePadding(ac)
	for _, z := range ac.Zones {
		printZone(cmd, ac, z, pad)
	}

	cmd.Println("")
}

func printZone(cmd *cobra.Command, ac *myplace.AirCon, z *myplace.Zone, pad int) {
	cmd.Printf(
		"  %2d %-"+strconv.Itoa(pad)+"s",
		z.Number,
		z.Name,
	)

	if ac.IsConstantZone(z) {
		cmd.Print("  C")
	} else {
		cmd.Print("   ")
	}

	cmd.Print("  ")

	if z.HasTempControl == 0 {
		cmd.Print("??.?°         ")
	} else if z.State == myplace.ZoneStateClosed {
		cmd.Printf(
			"%2.1f°   %2.1f°",
			z.CurrentTemp,
			z.TargetTemp,
		)
	} else if z.CurrentTemp < z.TargetTemp {
		cmd.Printf(
			"%2.1f° < %2.1f°",
			z.CurrentTemp,
			z.TargetTemp,
		)
	} else if z.CurrentTemp > z.TargetTemp {
		cmd.Printf(
			"%2.1f° > %2.1f°",
			z.CurrentTemp,
			z.TargetTemp,
		)
	} else {
		cmd.Printf(
			"%2.1f° = %2.1f°",
			z.CurrentTemp,
			z.TargetTemp,
		)
	}

	if ac.IsMyZone(z) {
		cmd.Print(" MY")
	} else if z.State == myplace.ZoneStateOpen {
		cmd.Print(" ON")
	} else {
		cmd.Print("   ")
	}

	if z.State == myplace.ZoneStateOpen {
		cmd.Printf(
			" %3d%%",
			z.DamperPercentage,
		)
	} else {
		cmd.Print("    -")
	}

	if z.Error != myplace.ZoneErrorNone {
		cmd.Printf("  (error: %s)", z.Error)
	}

	cmd.Println("")
}

// zoneNamePadding returns the padding width to use for zone names of the given
// AC unit.
func zoneNamePadding(ac *myplace.AirCon) int {
	pad := 0

	for _, z := range ac.Zones {
		n := utf8.RuneCountInString(z.Name)
		if n > pad {
			pad = n
		}
	}

	return pad
}

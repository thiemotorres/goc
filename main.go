package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/thiemotorres/goc/cmd"
	"github.com/thiemotorres/goc/internal/tui"
)

func main() {
	if len(os.Args) < 2 {
		// No args - launch TUI
		if err := tui.Run(); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
		return
	}

	switch os.Args[1] {
	case "ride":
		rideCmd := flag.NewFlagSet("ride", flag.ExitOnError)
		gpxPath := rideCmd.String("gpx", "", "GPX file for route simulation")
		ergWatts := rideCmd.Int("erg", 0, "ERG mode target watts")
		mock := rideCmd.Bool("mock", false, "Use mock Bluetooth (for development)")
		rideCmd.Parse(os.Args[2:])

		opts := cmd.RideOptions{
			GPXPath:  *gpxPath,
			ERGWatts: *ergWatts,
			Mock:     *mock,
		}

		if err := cmd.Ride(opts); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}

	case "history":
		historyCmd := flag.NewFlagSet("history", flag.ExitOnError)
		limit := historyCmd.Int("n", 20, "Number of rides to show")
		historyCmd.Parse(os.Args[2:])

		opts := cmd.HistoryOptions{
			Limit: *limit,
		}

		if err := cmd.History(opts); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}

	case "help", "-h", "--help":
		printUsage()

	default:
		fmt.Fprintf(os.Stderr, "Unknown command: %s\n\n", os.Args[1])
		printUsage()
		os.Exit(1)
	}
}

func printUsage() {
	fmt.Println("goc - Indoor Cycling Trainer")
	fmt.Println()
	fmt.Println("Usage:")
	fmt.Println("  goc <command> [options]")
	fmt.Println()
	fmt.Println("Commands:")
	fmt.Println("  ride      Start a cycling session")
	fmt.Println("  history   View past rides")
	fmt.Println("  help      Show this help")
	fmt.Println()
	fmt.Println("Ride options:")
	fmt.Println("  -gpx <file>   Load GPX route for simulation mode")
	fmt.Println("  -erg <watts>  ERG mode with fixed target power")
	fmt.Println("  -mock         Use mock Bluetooth (for testing)")
	fmt.Println()
	fmt.Println("History options:")
	fmt.Println("  -n <count>    Number of rides to show (default: 20)")
}

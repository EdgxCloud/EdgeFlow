//go:build linux
// +build linux

package main

import (
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/warthog618/go-gpiocdev"
)

func main() {
	pin := flag.Int("pin", 17, "GPIO pin number (BCM)")
	chip := flag.String("chip", "", "GPIO chip name (auto-detect if empty)")
	blink := flag.Bool("blink", true, "Blink the LED on/off")
	interval := flag.Duration("interval", 500*time.Millisecond, "Blink interval")
	on := flag.Bool("on", false, "Turn LED on (no blink)")
	off := flag.Bool("off", false, "Turn LED off (no blink)")
	flag.Parse()

	// Auto-detect GPIO chip if not specified
	chipName := *chip
	if chipName == "" {
		chipName = detectGPIOChip()
	}

	fmt.Printf("GPIO Test Tool\n")
	fmt.Printf("  Chip: %s\n", chipName)
	fmt.Printf("  Pin:  GPIO%d\n", *pin)
	fmt.Println()

	// Request the line as output
	line, err := gpiocdev.RequestLine(chipName, *pin, gpiocdev.AsOutput(0))
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: Failed to request GPIO%d on %s: %v\n", *pin, chipName, err)
		fmt.Fprintf(os.Stderr, "\nTroubleshooting:\n")
		fmt.Fprintf(os.Stderr, "  1. Make sure your user is in the 'gpio' group: sudo usermod -aG gpio $USER\n")
		fmt.Fprintf(os.Stderr, "  2. Log out and log back in after adding to group\n")
		fmt.Fprintf(os.Stderr, "  3. Check if the pin is already in use: cat /sys/kernel/debug/gpio\n")
		os.Exit(1)
	}
	defer func() {
		line.SetValue(0) // Turn off on exit
		line.Close()
		fmt.Println("\nGPIO pin released, LED off.")
	}()

	fmt.Printf("  GPIO%d requested successfully!\n\n", *pin)

	// Handle Ctrl+C for clean shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	if *on {
		// Just turn on
		if err := line.SetValue(1); err != nil {
			fmt.Fprintf(os.Stderr, "Error setting pin HIGH: %v\n", err)
			os.Exit(1)
		}
		fmt.Println("LED ON (GPIO HIGH). Press Ctrl+C to turn off and exit.")
		<-sigChan
		return
	}

	if *off {
		// Just turn off
		if err := line.SetValue(0); err != nil {
			fmt.Fprintf(os.Stderr, "Error setting pin LOW: %v\n", err)
			os.Exit(1)
		}
		fmt.Println("LED OFF (GPIO LOW). Press Ctrl+C to exit.")
		<-sigChan
		return
	}

	if *blink {
		fmt.Printf("Blinking LED at %v interval. Press Ctrl+C to stop.\n\n", *interval)
		state := false
		ticker := time.NewTicker(*interval)
		defer ticker.Stop()

		for {
			select {
			case <-sigChan:
				return
			case <-ticker.C:
				state = !state
				val := 0
				if state {
					val = 1
				}
				if err := line.SetValue(val); err != nil {
					fmt.Fprintf(os.Stderr, "Error writing to pin: %v\n", err)
					return
				}
				if state {
					fmt.Print("ON  ")
				} else {
					fmt.Print("OFF ")
				}
			}
		}
	}
}

// detectGPIOChip finds the main GPIO chip by reading sysfs labels
func detectGPIOChip() string {
	for _, chip := range []string{"gpiochip0", "gpiochip4"} {
		labelPath := fmt.Sprintf("/sys/bus/gpio/devices/%s/label", chip)
		data, err := os.ReadFile(labelPath)
		if err != nil {
			continue
		}
		label := string(data)
		// Pi 5 uses pinctrl-rp1, Pi 4 and earlier use pinctrl-bcm2835
		if contains(label, "pinctrl-rp1") || contains(label, "pinctrl-bcm2") {
			fmt.Printf("  Auto-detected: %s (%s)\n", chip, trim(label))
			return chip
		}
	}
	fmt.Println("  Auto-detect failed, using gpiochip0")
	return "gpiochip0"
}

func contains(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

func trim(s string) string {
	// Trim whitespace and newlines
	start, end := 0, len(s)-1
	for start <= end && (s[start] == ' ' || s[start] == '\n' || s[start] == '\r' || s[start] == '\t') {
		start++
	}
	for end >= start && (s[end] == ' ' || s[end] == '\n' || s[end] == '\r' || s[end] == '\t') {
		end--
	}
	if start > end {
		return ""
	}
	return s[start : end+1]
}

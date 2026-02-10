//go:build linux
// +build linux

package main

import (
	"log"
	"runtime"

	"github.com/edgeflow/edgeflow/internal/hal"
)

func initHAL() {
	if runtime.GOARCH == "arm64" || runtime.GOARCH == "arm" {
		rpiHAL, err := hal.NewRaspberryPiHAL()
		if err != nil {
			log.Printf("Warning: Failed to initialize RPi HAL: %v", err)
			log.Println("GPIO nodes will not be available, using Mock HAL")
			hal.SetGlobalHAL(hal.NewMockHAL())
			return
		}
		log.Printf("Raspberry Pi HAL initialized (%s via %s)",
			rpiHAL.Info().Name, rpiHAL.Info().GPIOChip)
		hal.SetGlobalHAL(rpiHAL)
	} else {
		log.Println("Non-ARM platform detected, using Mock HAL for GPIO")
		hal.SetGlobalHAL(hal.NewMockHAL())
	}
}

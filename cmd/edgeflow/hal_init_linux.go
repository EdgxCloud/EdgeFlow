//go:build linux
// +build linux

package main

import (
	"runtime"

	"github.com/edgeflow/edgeflow/internal/hal"
	"github.com/edgeflow/edgeflow/internal/logger"
	"go.uber.org/zap"
)

func initHAL() {
	if runtime.GOARCH == "arm64" || runtime.GOARCH == "arm" {
		rpiHAL, err := hal.NewRaspberryPiHAL()
		if err != nil {
			logger.Warn("Failed to initialize RPi HAL, using Mock HAL", zap.Error(err))
			hal.SetGlobalHAL(hal.NewMockHAL())
			return
		}
		logger.Info("Raspberry Pi HAL initialized",
			zap.String("board", rpiHAL.Info().Name),
			zap.String("gpio_chip", rpiHAL.Info().GPIOChip))
		hal.SetGlobalHAL(rpiHAL)
	} else {
		logger.Info("Non-ARM platform detected, using Mock HAL for GPIO")
		hal.SetGlobalHAL(hal.NewMockHAL())
	}
}

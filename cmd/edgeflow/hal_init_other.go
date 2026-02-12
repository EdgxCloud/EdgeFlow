//go:build !linux
// +build !linux

package main

import (
	"github.com/edgeflow/edgeflow/internal/hal"
	"github.com/edgeflow/edgeflow/internal/logger"
)

func initHAL() {
	logger.Info("Non-Linux platform detected, using Mock HAL for GPIO")
	hal.SetGlobalHAL(hal.NewMockHAL())
}

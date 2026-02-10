//go:build !linux
// +build !linux

package main

import (
	"log"

	"github.com/edgeflow/edgeflow/internal/hal"
)

func initHAL() {
	log.Println("Non-Linux platform detected, using Mock HAL for GPIO")
	hal.SetGlobalHAL(hal.NewMockHAL())
}

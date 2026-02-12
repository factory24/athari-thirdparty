package dtos

import (
	"github.com/chirpstack/chirpstack/api/go/v4/common"
)

// GatewayDTO represents the data needed to create or update a gateway
type GatewayDTO struct {
	SerialNumber  string
	Name          string
	Description   string
	Location      *common.Location
	StatsInterval uint32 // Stats interval in seconds
}

// DeviceDTO represents the data needed to create or update a device
type DeviceDTO struct {
	Eui          string
	Name         string
	Description  string
	ItemType     string
	SerialNumber string
}

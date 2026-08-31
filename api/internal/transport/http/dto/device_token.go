package dto

import "time"

type RegisterDeviceTokenRequest struct {
	Token    string `json:"token"`
	Platform string `json:"platform"` // ios | android
}

type UnregisterDeviceTokenRequest struct {
	Token string `json:"token"`
}

type DeviceTokenDTO struct {
	Token      string    `json:"token"`
	Platform   string    `json:"platform"`
	LastSeenAt time.Time `json:"last_seen_at"`
}

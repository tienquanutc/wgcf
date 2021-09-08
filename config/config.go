package config

import "github.com/ViRb3/wgcf/openapi"

const (
	DeviceId    = "device_id"
	AccessToken = "access_token"
	PrivateKey  = "private_key"
	LicenseKey  = "license_key"
)

type Context struct {
	DeviceId    string
	AccessToken string
	PrivateKey  string
	LicenseKey  string
	ApiClientAuth *openapi.APIClient
}

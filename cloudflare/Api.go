package cloudflare

import (
	"crypto/tls"
	"net/http"
	"time"

	"github.com/ViRb3/wgcf/config"
	"github.com/ViRb3/wgcf/openapi"
	"github.com/ViRb3/wgcf/util"
	"github.com/ViRb3/wgcf/wireguard"
)

const (
	ApiUrl     = "https://api.cloudflareclient.com"
	ApiVersion = "v0a1922"
)

var (
	DefaultHeaders = map[string]string{
		"User-Agent":        "okhttp/3.12.1",
		"CF-Client-Version": "a-6.3-1922",
	}
	DefaultTransport = &http.Transport{
		// Match app's TLS config or API will reject us with code 403 error 1020
		TLSClientConfig: &tls.Config{
			MinVersion: tls.VersionTLS12,
			MaxVersion: tls.VersionTLS12},
		ForceAttemptHTTP2: false,
		// From http.DefaultTransport
		Proxy:                 http.ProxyFromEnvironment,
		MaxIdleConns:          100,
		IdleConnTimeout:       90 * time.Second,
		TLSHandshakeTimeout:   10 * time.Second,
		ExpectContinueTimeout: 1 * time.Second,
	}
)

func MakeApiClient(authToken *string) *openapi.APIClient {
	httpClient := http.Client{Transport: DefaultTransport}
	apiClient := openapi.NewAPIClient(&openapi.Configuration{
		DefaultHeader: DefaultHeaders,
		UserAgent:     DefaultHeaders["User-Agent"],
		Debug:         false,
		Servers: []openapi.ServerConfiguration{
			{URL: ApiUrl},
		},
		HTTPClient: &httpClient,
	})
	if authToken != nil {
		apiClient.GetConfig().DefaultHeader["Authorization"] = "Bearer " + *authToken
	}
	return apiClient
}

func Register(publicKey *wireguard.Key, deviceModel string) (openapi.Register200Response, error) {
	timestamp := util.GetTimestamp()
	result, _, err := MakeApiClient(nil).DefaultApi.
		Register(nil, ApiVersion).
		RegisterRequest(openapi.RegisterRequest{
			FcmToken:  "", // not empty on actual client
			InstallId: "", // not empty on actual client
			Key:       publicKey.String(),
			Locale:    "en_US",
			Model:     deviceModel,
			Tos:       timestamp,
			Type:      "Android",
		}).Execute()
	return result, err
}

type Device openapi.UpdateSourceDevice200Response

func GetSourceDevice(ctx *config.Context) (*Device, error) {
	result, _, err := ctx.ApiClientAuth.DefaultApi.
		GetSourceDevice(nil, ApiVersion, ctx.DeviceId).
		Execute()
	castResult := Device{}
	if err := util.Restructure(&result, &castResult); err != nil {
		return nil, err
	}
	return &castResult, err
}

type Account openapi.GetAccount200Response

func GetAccount(ctx *config.Context) (*Account, error) {
	result, _, err := ctx.ApiClientAuth.DefaultApi.
		GetAccount(nil, ctx.DeviceId, ApiVersion).
		Execute()
	castResult := Account(result)
	return &castResult, err
}

func UpdateLicenseKey(ctx *config.Context, newPublicKey string) (*openapi.UpdateAccount200Response, *Device, error) {
	result, _, err := ctx.ApiClientAuth.DefaultApi.
		UpdateAccount(nil, ctx.DeviceId, ApiVersion).
		UpdateAccountRequest(openapi.UpdateAccountRequest{License: ctx.LicenseKey}).
		Execute()
	if err != nil {
		return nil, nil, err
	}
	// change public key as per official client
	result2, _, err := ctx.ApiClientAuth.DefaultApi.
		UpdateSourceDevice(nil, ApiVersion, ctx.DeviceId).
		UpdateSourceDeviceRequest(openapi.UpdateSourceDeviceRequest{Key: newPublicKey}).
		Execute()
	castResult := Device(result2)
	if err != nil {
		return nil, nil, err
	}
	return &result, &castResult, nil
}

type BoundDevice openapi.GetBoundDevices200Response

func GetBoundDevices(ctx *config.Context) ([]BoundDevice, error) {
	result, _, err := ctx.ApiClientAuth.DefaultApi.
		GetBoundDevices(nil, ctx.DeviceId, ApiVersion).
		Execute()
	if err != nil {
		return nil, err
	}
	var castResult []BoundDevice
	for _, device := range result {
		castResult = append(castResult, BoundDevice(device))
	}
	return castResult, nil
}

func GetSourceBoundDevice(ctx *config.Context) (*BoundDevice, error) {
	result, err := GetBoundDevices(ctx)
	if err != nil {
		return nil, err
	}
	return FindDevice(result, ctx.DeviceId)
}

func UpdateSourceBoundDeviceName(ctx *config.Context, newName string) (*BoundDevice, error) {
	return UpdateSourceBoundDevice(ctx, openapi.UpdateBoundDeviceRequest{
		Name: &newName,
	})
}

func UpdateSourceBoundDeviceActive(ctx *config.Context, active bool) (*BoundDevice, error) {
	return UpdateSourceBoundDevice(ctx, openapi.UpdateBoundDeviceRequest{
		Active: &active,
	})
}

func UpdateSourceBoundDevice(ctx *config.Context, data openapi.UpdateBoundDeviceRequest) (*BoundDevice, error) {
	result, _, err := ctx.ApiClientAuth.DefaultApi.
		UpdateBoundDevice(nil, ctx.DeviceId, ApiVersion, ctx.DeviceId).
		UpdateBoundDeviceRequest(data).
		Execute()
	if err != nil {
		return nil, err
	}
	var castResult []BoundDevice
	for _, device := range result {
		castResult = append(castResult, BoundDevice(device))
	}
	return FindDevice(castResult, ctx.DeviceId)
}

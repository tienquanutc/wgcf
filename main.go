package main

import (
	"github.com/ViRb3/wgcf/cloudflare"
	. "github.com/ViRb3/wgcf/cmd/shared"
	"github.com/ViRb3/wgcf/config"
	"github.com/ViRb3/wgcf/wireguard"
	"github.com/pkg/errors"
	"github.com/spf13/viper"
	"log"
	"net/http"
)

func main() {
	http.ListenAndServe("0.0.0.0:6688", http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		if req.URL.Path != "/warpplus.conf" {
			rw.Write([]byte("\n<h1><a href=\"/warpplus.conf\">Download a warpplus.conf now!</a></h1>"))
			return
		}
		ctx, err := registerAccount()
		if err != nil {
			println("failure register profile %s", err.Error())
			rw.WriteHeader(500)
			rw.Write([]byte("<h1>Internal Server Error</h1>"))
			return
		}
		profile, err := generateProfile(ctx)
		if err != nil {
			println("failure generate profile %s", err.Error())
			rw.WriteHeader(500)
			rw.Write([]byte("<h1>Internal Server Error</h1>"))
			return
		}

		rw.Header().Set("Content-Disposition", "attachment; filename=\"warpplus.conf\"")
		rw.Header().Set("Content-Type", "text/plain")
		rw.Write([]byte(profile))
		log.Println("Successfully created Cloudflare Warp account")
	}))
}

func registerAccount() (*config.Context, error) {
	privateKey, err := wireguard.NewPrivateKey()
	if err != nil {
		return nil, err
	}

	device, err := cloudflare.Register(privateKey.Public(), "PC")
	if err != nil {
		return nil, err
	}

	ctx := config.Context{
		PrivateKey:    privateKey.String(),
		DeviceId:      device.Id,
		AccessToken:   device.Token,
		LicenseKey:    device.Account.License,
		ApiClientAuth: cloudflare.MakeApiClient(&device.Token),
	}

	_, err = SetDeviceName(&ctx, "deviceName")
	if err != nil {
		return nil, err
	}
	thisDevice, err := cloudflare.GetSourceDevice(&ctx)
	if err != nil {
		return nil, err
	}

	boundDevice, err := cloudflare.UpdateSourceBoundDeviceActive(&ctx, true)
	if err != nil {
		return nil, err
	}
	if !boundDevice.Active {
		return nil, errors.New("failed to activate device")
	}

	PrintDeviceData(thisDevice, boundDevice)
	log.Println("Successfully created Cloudflare Warp account")
	return &ctx, nil
}

func generateProfile(ctx *config.Context) (string, error) {
	thisDevice, err := cloudflare.GetSourceDevice(ctx)
	if err != nil {
		return "", err
	}
	boundDevice, err := cloudflare.GetSourceBoundDevice(ctx)
	if err != nil {
		return "", err
	}

	profile, err := wireguard.NewProfile(&wireguard.ProfileData{
		PrivateKey: viper.GetString(config.PrivateKey),
		Address1:   thisDevice.Config.Interface.Addresses.V4,
		Address2:   thisDevice.Config.Interface.Addresses.V6,
		PublicKey:  thisDevice.Config.Peers[0].PublicKey,
		Endpoint:   thisDevice.Config.Peers[0].Endpoint.Host,
	})
	if err != nil {
		return "", err
	}
	PrintDeviceData(thisDevice, boundDevice)
	return profile.String(), nil
}

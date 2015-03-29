package main

import (
	"time"

	"github.com/appleague/go-enigma2"
	"github.com/ninjasphere/go-ninja/api"
	"github.com/ninjasphere/go-ninja/config"
	"github.com/ninjasphere/go-ninja/devices"
	"github.com/ninjasphere/go-ninja/model"
)

type Device struct {
	devices.MediaPlayerDevice
	stb *enigma2.STB
}

func (d *Device) updateHost(host string) {
	d.stb.Host = host
}

func newDevice(driver ninja.Driver, conn *ninja.Connection, cfg *STBConfig) (*Device, error) {

	player, err := devices.CreateMediaPlayerDevice(driver, &model.Device{
		NaturalID:     cfg.ID,
		NaturalIDType: "enigma2-stb",
		Name:          &cfg.Name,
		Signatures: &map[string]string{
			"ninja:manufacturer": "VUPlus",
			"ninja:productName":  "Smart STB",
			"ninja:thingType":    "mediaplayer",
			"ip:mac":             cfg.ID,
		},
	}, conn)

	if err != nil {
		return nil, err
	}

	enigma2.EnableLogging = true

	stb := enigma2.STB{
		Host:            cfg.Host,
		ApplicationID:   config.MustString("userId"),
		ApplicationName: "Ninja Sphere         ",
	}

	// On-off Channel
	player.ApplyOff = func() error {
		return stb.SendCommand("POWEROFF")
	}

	if err := player.EnableOnOffChannel("state"); err != nil {
		player.Log().Fatalf("Failed to enable control channel: %s", err)
	}

	go func() {

		// Continuous updates as STB goes online and offline
		for online := range stb.OnlineState(time.Second * 15) {
			player.UpdateOnOffState(online)
		}
	}()

	return &Device{*player, &stb}, nil
}

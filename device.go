package main

import (
	"time"

	"github.com/appleague/go-NS-enigma2"
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
	log.Infof("updateHost")
	d.stb.Host = host
}

func newDevice(driver ninja.Driver, conn *ninja.Connection, cfg *STBConfig) (*Device, error) {
	log.Infof("newDevice")
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

	stb := enigma2.STB{
		Host:            cfg.Host,
		ApplicationID:   config.MustString("userId"),
		ApplicationName: cfg.Name,
		//ApplicationName: "Ninja Sphere         ",
	}

	stb.SendMessage("INFO VUPlus Driver loaded")

	// Volume Channel
	player.ApplyVolumeUp = func() error {
		stb.SendCommand("VOLUP")
		return nil
	}

	player.ApplyVolumeDown = func() error {
		stb.SendCommand("VOLDOWN")
		return nil
	}

	player.ApplyToggleMuted = func() error {
		stb.SendCommand("MUTE")
		return nil
	}

	if err := player.EnableVolumeChannel(false); err != nil {
		player.Log().Fatalf("Failed to enable volume channel: %s", err)
	}

	// Media Control Channel
	player.ApplyPlayPause = func(play bool) error {
		stb.SendCommand("TOGGLEONOFF")
		return nil
	}

	player.ApplyStop = func() error {
		stb.SendCommand("POWEROFF")
		return nil
	}

	if err := player.EnableControlChannel([]string{}); err != nil {
		player.Log().Fatalf("Failed to enable control channel: %s", err)
	}

	// Toggle On-off Channel
	player.ApplyToggleOnOff = func() error {
		return stb.SendCommand("TOGGLEONOFF")
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
		for online := range stb.OnlineState(time.Second * 60) {
			player.UpdateOnOffState(online)
		}
	}()

	return &Device{*player, &stb}, nil
}

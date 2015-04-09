package main

import (
	"fmt"
	"os"
	"time"

	"github.com/appleague/go-enigma2"
	"github.com/mostlygeek/arp"
	"github.com/ninjasphere/go-ninja/api"
	"github.com/ninjasphere/go-ninja/logger"
	// actually the config export for the stb wont work
	//"github.com/ninjasphere/go-ninja/model"
	"github.com/ninjasphere/go-ninja/support"
)

var info = ninja.LoadModuleInfo("./package.json")
var log = logger.GetLogger(info.Name)

type Driver struct {
	support.DriverSupport
	config  Config
	devices map[string]*Device
}

type Config struct {
	STBs map[string]*STBConfig
}

func (c *Config) get(id string) *STBConfig {
	for _, stb := range c.STBs {
		if stb.ID == id {
			return stb
		}
	}
	return nil
}

type STBConfig struct {
	ID   string
	Name string
	Host string
}

func NewDriver() (*Driver, error) {
	log.Infof("NewDriver")
	driver := &Driver{
		devices: make(map[string]*Device),
	}

	err := driver.Init(info)
	if err != nil {
		log.Fatalf("Failed to initialize driver: %s", err)
	}

	err = driver.Export(driver)
	if err != nil {
		log.Fatalf("Failed to export driver: %s", err)
	}

	return driver, nil
}

func (d *Driver) deleteSTB(id string) error {
	log.Infof("deleteSTB")
	delete(d.config.STBs, id)

	err := d.SendEvent("config", &d.config)

	// TODO: Can't unexport devices at the moment, so we should restart the driver...
	go func() {
		time.Sleep(time.Second * 2)
		os.Exit(0)
	}()

	return err
}

func (d *Driver) saveSTB(stb STBConfig) error {
	log.Infof("saveSTB")
	if !(&enigma2.STB{Host: stb.Host}).Online(time.Second * 5) {
		return fmt.Errorf("Could not connect to STB. Is it online?")
	}

	mac, err := getMACAddress(stb.Host, time.Second*10)

	if err != nil {
		return fmt.Errorf("Failed to get mac address for STB. Is it online?")
	}

	existing := d.config.get(mac)

	if existing != nil {
		existing.Host = stb.Host
		existing.Name = stb.Name
		device, ok := d.devices[mac]
		if ok {
			device.stb.Host = stb.Host
		}
	} else {
		stb.ID = mac
		d.config.STBs[mac] = &stb

		go d.createSTBDevice(&stb)
	}

	return d.SendEvent("config", d.config)
}

func (d *Driver) Start(config *Config) error {
	log.Infof("Driver Starting with config %+v", config)

	if config.STBs == nil {
		log.Infof("config.STBs == nil")
		config.STBs = make(map[string]*STBConfig)

		var stbcfg STBConfig

		// actually the config export for the stb wont work
		//stb := enigma2.STB{}

		// actually the config export for the stb wont work
		// setting the stb config manually
		stbcfg.Host = "10.0.0.20"
		stbcfg.ID = "VU+ Solo 2"

		config.STBs[stbcfg.ID] = &stbcfg

		fmt.Println("Added STB config with Id " + stbcfg.ID)
	}

	d.config = *config

	for _, cfg := range config.STBs {
		log.Infof("createSTBDevice %+v", cfg)
		d.createSTBDevice(cfg)
	}

	// actually the config export for the stb wont work
	//log.Infof("MustExportService")
	//d.Conn.MustExportService(&configService{d}, "$driver/"+info.ID+"/configure", &model.ServiceAnnouncement{
	//	Schema: "/protocol/configuration",
	//})

	return nil
}

func (d *Driver) createSTBDevice(cfg *STBConfig) {
	device, err := newDevice(d, d.Conn, cfg)

	if err != nil {
		log.Fatalf("Failed to create new VUPlus STB device host:%s id:%s name:%s : %s", cfg.Host, cfg.ID, cfg.Name, err)
	}

	d.devices[cfg.ID] = device
}

func getMACAddress(host string, timeout time.Duration) (string, error) {
	log.Infof("getMACAddress")
	timedOut := false
	success := make(chan string, 1)

	go func() {
		for {
			if timedOut {
				break
			}
			id := arp.Search(host)
			if id != "" && id != "(incomplete)" {
				success <- id
			}

			time.Sleep(time.Millisecond * 500)
		}
	}()

	select {
	case mac := <-success:
		return mac, nil
	case <-time.After(timeout):
		timedOut = true
		return "", fmt.Errorf("Timed out searching for MAC address")
	}
}

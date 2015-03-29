package main

import (
	"encoding/json"
	"fmt"

	"github.com/ninjasphere/go-ninja/model"
	"github.com/ninjasphere/go-ninja/suit"
)

type configService struct {
	driver *Driver
}

func (c *configService) GetActions(request *model.ConfigurationRequest) (*[]suit.ReplyAction, error) {
	return &[]suit.ReplyAction{
		suit.ReplyAction{
			Name:  "",
			Label: "VUPlus STBs",
		},
	}, nil
}

func (c *configService) Configure(request *model.ConfigurationRequest) (*suit.ConfigurationScreen, error) {
	log.Infof("Incoming configuration request. Action:%s Data:%s", request.Action, string(request.Data))

	switch request.Action {
	case "list":
		return c.list()
	case "":
		if len(c.driver.config.STBs) > 0 {
			return c.list()
		}
		fallthrough
	case "new":
		return c.edit(STBConfig{})
	case "edit":

		var vals map[string]string
		json.Unmarshal(request.Data, &vals)
		config := c.driver.config.get(vals["stb"])

		if config == nil {
			return c.error(fmt.Sprintf("Could not find stb with id: %s", vals["stb"]))
		}

		return c.edit(*config)
	case "delete":

		var vals map[string]string
		json.Unmarshal(request.Data, &vals)

		err := c.driver.deleteSTB(vals["stb"])

		if err != nil {
			return c.error(fmt.Sprintf("Failed to delete stb: %s", err))
		}

		return c.list()
	case "save":
		var cfg STBConfig
		err := json.Unmarshal(request.Data, &cfg)
		if err != nil {
			return c.error(fmt.Sprintf("Failed to unmarshal save config request %s: %s", request.Data, err))
		}

		err = c.driver.saveSTB(cfg)

		if err != nil {
			return c.error(fmt.Sprintf("Could not save stb: %s", err))
		}

		return c.list()
	default:
		return c.error(fmt.Sprintf("Unknown action: %s", request.Action))
	}
}

func (c *configService) error(message string) (*suit.ConfigurationScreen, error) {

	return &suit.ConfigurationScreen{
		Sections: []suit.Section{
			suit.Section{
				Contents: []suit.Typed{
					suit.Alert{
						Title:        "Error",
						Subtitle:     message,
						DisplayClass: "danger",
					},
				},
			},
		},
		Actions: []suit.Typed{
			suit.ReplyAction{
				Label: "Cancel",
				Name:  "list",
			},
		},
	}, nil
}
func (c *configService) list() (*suit.ConfigurationScreen, error) {

	var stbs []suit.ActionListOption

	for _, stb := range c.driver.config.STBs {
		stbs = append(stbs, suit.ActionListOption{
			Title: stb.Name,
			//Subtitle: tv.ID,
			Value: stb.ID,
		})
	}

	screen := suit.ConfigurationScreen{
		Title: "VUPlus STBs",
		Sections: []suit.Section{
			suit.Section{
				Contents: []suit.Typed{
					suit.ActionList{
						Name:    "stb",
						Options: stbs,
						PrimaryAction: &suit.ReplyAction{
							Name:        "edit",
							DisplayIcon: "pencil",
						},
						SecondaryAction: &suit.ReplyAction{
							Name:         "delete",
							Label:        "Delete",
							DisplayIcon:  "trash",
							DisplayClass: "danger",
						},
					},
				},
			},
		},
		Actions: []suit.Typed{
			suit.CloseAction{
				Label: "Close",
			},
			suit.ReplyAction{
				Label:        "New STB",
				Name:         "new",
				DisplayClass: "success",
				DisplayIcon:  "star",
			},
		},
	}

	return &screen, nil
}

func (c *configService) edit(config STBConfig) (*suit.ConfigurationScreen, error) {

	title := "New VUplus STB"
	if config.ID != "" {
		title = "Editing VUplus STB"
	}

	screen := suit.ConfigurationScreen{
		Title: title,
		Sections: []suit.Section{
			suit.Section{
				Contents: []suit.Typed{
					suit.InputHidden{
						Name:  "id",
						Value: config.ID,
					},
					suit.InputText{
						Name:        "name",
						Before:      "Name",
						Placeholder: "My VUPlus",
						Value:       config.Name,
					},
					suit.InputText{
						Name:        "host",
						Before:      "Host",
						Placeholder: "IP or Hostname",
						Value:       config.Host,
					},
				},
			},
		},
		Actions: []suit.Typed{
			suit.CloseAction{
				Label: "Cancel",
			},
			suit.ReplyAction{
				Label:        "Save",
				Name:         "save",
				DisplayClass: "success",
				DisplayIcon:  "star",
			},
		},
	}

	return &screen, nil
}

func i(i int) *int {
	return &i
}

func contains(s []string, e string) bool {
	for _, a := range s {
		if a == e {
			return true
		}
	}
	return false
}

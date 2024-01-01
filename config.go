package main

import (
	"encoding/json"
	"errors"
	"os"
	"path"

	"github.com/bwmarrin/discordgo"
)

const ConfigFilePath = "data/config.json"

type Config struct {
	Token            string                     `json:"token"`
	Status           discordgo.UpdateStatusData `json:"status"`
	DMEnable         bool                       `json:"dm_enable"`
	BlacklistServers []string                   `json:"blacklist_servers"`
}

func ReadConfig() (config Config, error error) {
	if file, err := os.ReadFile(ConfigFilePath); err != nil {
		if !errors.Is(err, os.ErrNotExist) {
			return
		}

		createNewConfigFile()
	} else {
		if err := json.Unmarshal(file, &config); err != nil {
			createNewConfigFile()

			return ReadConfig()
		}
	}

	return
}

func NewConfig() Config {
	return Config{
		Token: "",
		Status: discordgo.UpdateStatusData{
			Status: "dnd",
			AFK:    false,
		},
		DMEnable:         false,
		BlacklistServers: []string{},
	}
}

func createNewConfigFile() {
	data, _ := json.MarshalIndent(NewConfig(), "", "  ")

	os.MkdirAll(path.Dir(ConfigFilePath), 0o644)
	os.WriteFile(ConfigFilePath, data, 0o644)
}

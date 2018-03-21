package configs

import (
	"io/ioutil"
	"encoding/json"
	"os"
	"log"
)

type Configs struct {
	Port int `json:"port"`
	RestOnly bool `json:"rest_only"`
	SteamUser string `json:"steam_user"`
	Blacklist []string `json:"genre_blacklist"`
}

var configFile = "config.json"
var activeConfigs *Configs = nil

func (c *Configs) store(fileName string) error {
	data, err := json.MarshalIndent(c, "", "\t")
	if err != nil {
		return err
	}

	return ioutil.WriteFile(fileName, data, 0600)
}

func getDefaultConfigs() *Configs {
	bl := make([]string, 1)
	bl[0] = "none"
	return &Configs{8080, false, "", bl}
}

func GetConfigs() *Configs {
	var err error = nil
	if activeConfigs == nil {
		data, err := ioutil.ReadFile(configFile)
		if err == nil {
			var configs Configs
			err = json.Unmarshal(data, &configs)
			if err == nil {
				activeConfigs = &configs
			}
		} else if os.IsNotExist(err) {
			activeConfigs = getDefaultConfigs()
		}
		err = activeConfigs.store(configFile)
	}
	if err != nil {
		log.Fatalf("Error reading configurations: %s\n", err.Error())
	}
	return activeConfigs
}
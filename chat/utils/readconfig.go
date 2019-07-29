package utils

import (
	"encoding/json"
	"io/ioutil"
	"libruary/models"
	"log"
)

// ReadConfig reading json file "config.json"
// parse it into json struct and returns new models Config
func ReadConfig() *models.Config {
	var config models.Config

	file, err := ioutil.ReadFile("config.json")

	if err != nil {
		log.Panic(err)
	}

	if err := json.Unmarshal(file, &config); err != nil {
		panic(err)
	}

	return &config
}

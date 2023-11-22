package config

import (
	"errors"
	"log"
	"os"

	"gopkg.in/yaml.v3"
)

func defaultConfig() Config {
	log.Println("setting config to default")
	return Config{
		OrderbookFrames:  10 * 60 * 5,
		ChangeoverFrames: 10 * 10,
		Filepath:         "temp",
	}
}

func ReadConfigFromFile(filepath string) Config {
	config := Config{}

	log.Printf("Attempting to read config %s... ", filepath)

	if _, err := os.Stat(filepath); errors.Is(err, os.ErrNotExist) {
		log.Printf("File does not exist")
		return defaultConfig()
	}

	bytes, err := os.ReadFile(filepath)
	if err != nil {
		log.Printf("Failed to read config: %s \n", err)
		return defaultConfig()
	}

	err = yaml.Unmarshal(bytes, &config)
	if err != nil {
		log.Printf("Failed to decode config: %s \n", err)
		return defaultConfig()
	}

	log.Printf("Config successfully read. %#v \n", config)

	return config
}

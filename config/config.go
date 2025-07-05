package config

import (
	"log"

	"github.com/spf13/viper"
)

type Configuration struct {
	Server      ServerConfiguration
	Database    DatabaseConfiguration
	Application ApplicationConfiguration
}

// SetupConfig configuration
func SetupConfig() error {
	var configuration *Configuration

	viper.SetConfigFile(".env")
	if err := viper.ReadInConfig(); err != nil {
		log.Printf("Error to reading config file, %s", err)
		return err
	}

	err := viper.Unmarshal(&configuration)
	if err != nil {
		log.Printf("error to decode, %v", err)
		return err
	}

	return nil
}

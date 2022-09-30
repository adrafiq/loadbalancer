package utils

import (
	"fmt"

	"github.com/spf13/viper"
)

var Config = viper.New()

func init() {
	Config.SetConfigFile(".env")
	Config.AddConfigPath("./")
	if err := Config.ReadInConfig(); err != nil {
		panic(fmt.Errorf("couldnt find .env configuration file"))
	}
}

func init() {
	// To set defaults
	Config.SetDefault("env", "dev")
	Config.SetDefault("logMode", "info")
}

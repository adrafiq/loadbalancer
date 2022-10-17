package utils

import (
	"fmt"

	"github.com/spf13/viper"
)

var Config = viper.New()

func InitConfig() {
	Config.SetConfigFile("config.yaml")
	Config.SetConfigType("yaml")
	Config.AddConfigPath(".")
	if err := Config.ReadInConfig(); err != nil {
		panic(fmt.Errorf("couldnt find config.yaml configuration file\n", err))
	}
	Config.SetDefault("env", "dev")
	Config.SetDefault("logLevel", "info")
	Config.SetDefault("port", 3000)

}

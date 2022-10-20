package config

import (
	"fmt"

	"github.com/spf13/viper"
)

func NewConfig(fileName string) *viper.Viper {
	var config = viper.New()
	config.SetConfigFile(fileName)
	if err := config.ReadInConfig(); err != nil {
		panic(fmt.Errorf("couldnt find config.yaml configuration file\n", err))
	}
	config.SetDefault("env", "dev")
	config.SetDefault("logLevel", "info")
	config.SetDefault("port", 3000)
	return config
}

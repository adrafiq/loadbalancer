package utils

import (
	"fmt"

	"github.com/sirupsen/logrus"
)

var Logger = logrus.New()

func init() {
	if logLevel, err := logrus.ParseLevel(string(Config.GetString("logLevel"))); err != nil {
		panic(fmt.Errorf("couldnt find .env configuration file"))

	} else {
		Logger.SetLevel(logLevel)
	}
}

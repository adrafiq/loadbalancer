package utils

import (
	"fmt"

	"github.com/sirupsen/logrus"
)

var Logger = logrus.New()

func NewLogger() logrus.Logger {
	if logLevel, err := logrus.ParseLevel(string(Config.GetString("logLevel"))); err != nil {
		panic(fmt.Errorf("bad value for logLevel. must be info or debug"))

	} else {
		Logger.SetLevel(logLevel)
	}
	return *Logger
}

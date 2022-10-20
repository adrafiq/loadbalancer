package log

import (
	"fmt"

	"github.com/sirupsen/logrus"
)

func NewLogger(logLevel string) *logrus.Logger {
	logger := logrus.New()
	logrusLogLevel, err := logrus.ParseLevel(logLevel)
	if err != nil {
		panic(fmt.Errorf("bad value for logLevel. must be info or debug", err))
	}
	logger.SetLevel(logrusLogLevel)
	return logger
}

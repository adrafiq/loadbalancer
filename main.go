package main

import (
	"infrastructure/loadbalancer/utils"
)

var Logger = utils.Logger
var Config = utils.Config

func main() {
	Logger.Info("Hello World")
	Logger.Debug("Debug Log")
}

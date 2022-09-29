package main

import (
	"infrastructure/loadbalancer/utils"
)

var Logger = utils.Logger

func main() {
	Logger.Info("Hello World")
}

package main

import (
	"fmt"
	"infrastructure/loadbalancer/utils"
)

var Logger = utils.Logger
var Config = utils.Config

func main() {
	fmt.Println(Config.Get("env"))
	Logger.Info("Hello World")
}

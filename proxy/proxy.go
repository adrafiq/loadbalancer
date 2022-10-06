package proxy

import (
	"infrastructure/loadbalancer/utils"
)

type Url string
type Host struct {
	Name    Url   `yaml:"name"`
	Servers []Url `yaml:"servers"`
}

var HostConfigured Host

func init() {
	utils.Config.UnmarshalKey("host", &HostConfigured)
}

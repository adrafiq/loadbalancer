package proxy

import (
	"errors"
	"fmt"
	"infrastructure/loadbalancer/utils"
	"math/rand"
	"time"
)

const (
	Random             = "random"
	RoundRobin         = "roundrobin"
	WeightedRoundRobin = "weightedroundrobin"
)

type Url string
type Host struct {
	Name    Url    `yaml:"name"`
	Servers []Url  `yaml:"servers"`
	Scheme  string `yaml:"scheme"`
}

var HostConfigured Host

func init() {
	utils.Config.UnmarshalKey("host", &HostConfigured)
	fmt.Println(HostConfigured)
}

func (h *Host) GetNext() (Url, error) {
	switch h.Scheme {
	case Random:
		rand.Seed(time.Now().Unix())
		return h.Servers[rand.Intn(len(h.Servers))], nil
	}
	return "", errors.New("unrecognized scheme")
}

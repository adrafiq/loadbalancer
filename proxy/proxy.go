package proxy

import (
	"errors"
	"infrastructure/loadbalancer/utils"
	"math/rand"
	"net/url"
	"time"
)

const (
	Random             = "random"
	RoundRobin         = "roundrobin"
	WeightedRoundRobin = "weightedroundrobin"
)

type Url string
type Host struct {
	Name    string     `yaml:"name"`
	Servers []*url.URL `yaml:"servers"`
	Scheme  string     `yaml:"scheme"`
}

var HostConfigured Host
var iterator int = 0

func init() {
	utils.Config.UnmarshalKey("host", &HostConfigured)
	targetServers := utils.Config.GetStringSlice("host.servers")
	HostConfigured.Servers = nil
	for _, s := range targetServers {
		targetUrl, _ := url.Parse(s)
		HostConfigured.Servers = append(HostConfigured.Servers, targetUrl)
	}
}

func (h *Host) GetNext() (*url.URL, error) {
	switch h.Scheme {
	case Random:
		rand.Seed(time.Now().Unix())
		return h.Servers[rand.Intn(len(h.Servers))], nil
	case RoundRobin:
		if iterator == len(h.Servers) {
			iterator = 0
		}
		targetIndex := iterator
		iterator++
		return h.Servers[targetIndex], nil
	}
	return nil, errors.New("unrecognized scheme")
}

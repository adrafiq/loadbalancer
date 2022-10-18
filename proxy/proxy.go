package proxy

import (
	"errors"
	"infrastructure/loadbalancer/utils"
	"math/rand"
	"net/http"
	"time"
)

const (
	Random             = "random"
	RoundRobin         = "roundrobin"
	WeightedRoundRobin = "weightedroundrobin"
	First              = 0
	Reset              = 0
)

type Server struct {
	Name   string  `yaml:"name"`
	Weight float32 `yaml:"weight"`
}
type Host struct {
	Name           string   `yaml:"name"`
	Servers        []Server `yaml:"servers"`
	HealthyServers []Server
	Scheme         string `yaml:"scheme"`
	Health         string `yaml:"health"`
}

var HostConfigured Host
var iterator int
var serversProgress []float32
var roundSize int
var currentRound int

func InitHost() {
	utils.Config.UnmarshalKey("host", &HostConfigured)
	utils.Logger.Info(HostConfigured)
	serversProgress = make([]float32, len(HostConfigured.Servers))
	for _, v := range HostConfigured.Servers {
		roundSize += int(v.Weight)
	}
}

// Refactor: Use closure to return a function instead of cases
func (h *Host) GetNext() (string, error) {
	switch h.Scheme {
	case Random:
		rand.Seed(time.Now().Unix())
		return h.Servers[rand.Intn(len(h.Servers))].Name, nil
	case RoundRobin:
		if iterator == len(h.Servers) {
			iterator = Reset
		}
		targetIndex := iterator
		iterator++
		return h.Servers[targetIndex].Name, nil
	case WeightedRoundRobin:
		if currentRound == roundSize {
			serversProgress = make([]float32, len(h.Servers))
			currentRound = Reset
		}
		var minProgress = serversProgress[First] / h.Servers[First].Weight
		var minProgressIndex int
		for index, server := range h.Servers {
			progress := serversProgress[index] / server.Weight
			if progress <= minProgress {
				minProgressIndex = index
				minProgress = progress
			}
		}
		serversProgress[minProgressIndex]++
		currentRound++
		return h.Servers[minProgressIndex].Name, nil
	}
	return "", errors.New("unrecognized scheme")
}

func (h *Host) UpdateHealthyServer() {
	var healthyServers []Server
	scheme := "http"
	client := http.DefaultClient
	req, _ := http.NewRequest("GET", "", nil)
	for _, server := range h.Servers {
		req.URL.Host = server.Name
		req.URL.Path = h.Health
		req.URL.Scheme = scheme
		res, err := client.Do(req)
		if err != nil {
			utils.Logger.Errorln("error in calling health endpoint", err)
<<<<<<< HEAD
			continue
		}
		defer res.Body.Close()
=======
		} else {
			defer res.Body.Close()
		}
>>>>>>> 86dca9b (Adds UpdateHealthyServer method)
		if res.StatusCode == 200 {
			healthyServers = append(healthyServers, server)
		}
	}
	h.HealthyServers = healthyServers
}

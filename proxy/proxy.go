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

type Server struct {
	Name   string  `yaml:"name"`
	Weight float32 `yaml:"weight"`
}
type Host struct {
	Name    string   `yaml:"name"`
	Servers []Server `yaml:"servers"`
	Scheme  string   `yaml:"scheme"`
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

func (h *Host) GetNext() (string, error) {
	switch h.Scheme {
	case Random:
		rand.Seed(time.Now().Unix())
		return h.Servers[rand.Intn(len(h.Servers))].Name, nil
	case RoundRobin:
		if iterator == len(h.Servers) {
			iterator = 0
		}
		targetIndex := iterator
		iterator++
		return h.Servers[targetIndex].Name, nil
	case WeightedRoundRobin:
		if currentRound == roundSize {
			serversProgress = make([]float32, len(h.Servers))
			currentRound = 0
			fmt.Print("Round finished")
		}
		var minProgress = serversProgress[0] / h.Servers[0].Weight
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

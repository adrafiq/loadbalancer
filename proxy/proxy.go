package proxy

import (
	"errors"
	"math/rand"
	"net/http"
	"time"

	"github.com/sirupsen/logrus"
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
	Name            string   `yaml:"name"`
	Servers         []Server `yaml:"servers"`
	HealthyServers  []Server
	Scheme          string `yaml:"scheme"`
	Health          string `yaml:"health"`
	Interval        int    `yaml:"interval"`
	iterator        int
	serversProgress []float32
	roundSize       int
	currentRound    int
	logger          *logrus.Logger
}

func (h *Host) resetState() {
	h.serversProgress = make([]float32, len(h.HealthyServers))
	h.currentRound = Reset
	h.iterator = Reset
	h.roundSize = Reset
	for _, server := range h.HealthyServers {
		h.roundSize += int(server.Weight)
	}
}

func NewHost(l *logrus.Logger) *Host {
	host := Host{logger: l}
	return &host
}

// Refactor: Use closure to return a function instead of cases
func (h *Host) GetNext() (string, error) {
	switch h.Scheme {
	case Random:
		rand.Seed(time.Now().Unix())
		return h.HealthyServers[rand.Intn(len(h.HealthyServers))].Name, nil
	case RoundRobin:
		if h.iterator == len(h.HealthyServers) {
			h.iterator = Reset
		}
		targetIndex := h.iterator
		h.iterator++
		return h.HealthyServers[targetIndex].Name, nil
	case WeightedRoundRobin:
		if h.currentRound == h.roundSize {
			h.resetState()
		}
		minProgress := h.serversProgress[First] / h.HealthyServers[First].Weight
		minProgressIndex := 0
		for index, server := range h.HealthyServers {
			progress := h.serversProgress[index] / server.Weight
			if progress <= minProgress {
				minProgressIndex = index
				minProgress = progress
			}
		}
		h.serversProgress[minProgressIndex]++
		h.currentRound++
		return h.HealthyServers[minProgressIndex].Name, nil
	}
	return "", errors.New("unrecognized scheme, check host configuration")
}

func (h *Host) CheckHealth() {
	var healthyServers []Server
	currentCount := len(h.HealthyServers)
	scheme := "http"
	client := http.DefaultClient
	req, _ := http.NewRequest("GET", "", nil)
	req.URL.Path = h.Health
	req.URL.Scheme = scheme
	for _, server := range h.Servers {
		req.URL.Host = server.Name
		res, err := client.Do(req)
		if err != nil {
			h.logger.Errorln("error in calling health endpoint", err)
			continue
		}
		defer res.Body.Close()
		if res.StatusCode == 200 {
			healthyServers = append(healthyServers, server)
		}
	}
	h.HealthyServers = healthyServers
	if currentCount != len(healthyServers) {
		h.resetState()
	}
}

package proxy

import (
	"errors"
	"net/http"

	"github.com/sirupsen/logrus"
)

const (
	Random             = "random"
	RoundRobin         = "roundrobin"
	WeightedRoundRobin = "weightedroundrobin"
	First              = 0
	Reset              = 0
)

type LoadBalancer interface {
	resetState()
	GetNext(randInt func(int) int) (string, error)
}

type Server struct {
	Name   string `yaml:"name"`
	Weight int    `yaml:"weight"`
}
type Host struct {
	Name            string   `yaml:"name"`
	Servers         []Server `yaml:"servers"`
	HealthyServers  []Server
	Scheme          string `yaml:"scheme"`
	Health          string `yaml:"health"`
	Interval        int    `yaml:"interval"`
	Port            string `yaml:"port"`
	cursor          int
	serversProgress []int
	roundSize       int
	currentRound    int
	logger          *logrus.Logger
}

func (h *Host) resetState() {
	h.serversProgress = make([]int, len(h.HealthyServers))
	h.currentRound = Reset
	h.cursor = Reset
	h.roundSize = Reset
	for _, server := range h.HealthyServers {
		h.roundSize += int(server.Weight)
	}
}

func NewHost(l *logrus.Logger) *Host {
	host := Host{logger: l}
	return &host
}

func (h *Host) SetLogger(l *logrus.Logger) {
	h.logger = l
}

func (h *Host) GetNext(randInt func(int) int) (string, error) {
	switch h.Scheme {
	case Random:

		return h.HealthyServers[randInt(len(h.HealthyServers))].Name, nil
	case RoundRobin:
		if h.cursor == len(h.HealthyServers) {
			h.cursor = Reset
		}
		targetIndex := h.cursor
		h.cursor++
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
	current := h.HealthyServers
	scheme := "http"
	req, _ := http.NewRequest("GET", "", nil)
	client := http.DefaultClient
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
	if len(current) != len(healthyServers) {
		h.resetState()
		return
	}
	for index, server := range current {
		if server != healthyServers[index] {
			h.resetState()
		}
	}
}

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

type LoadBalancer interface {
	resetState()
	GetNext() (string, error)
}

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
	Port            string `yaml:"port"`
	iterator        int
	serversProgress []float32
	roundSize       int
	currentRound    int
	logger          *logrus.Logger
}

// Unit is method
// Type is command
// Exit points
//
//	Check that all states have been changed of a given object.
func (h *Host) resetState() {
	h.serversProgress = make([]float32, len(h.HealthyServers))
	h.currentRound = Reset
	h.iterator = Reset
	h.roundSize = Reset
	for _, server := range h.HealthyServers {
		h.roundSize += int(server.Weight)
	}
}

// Unit is function
// TYpe is command
// Exit points
//
//	Check if the returned value should be as expected
func NewHost(l *logrus.Logger) *Host {
	host := Host{logger: l}
	return &host
}

// Unit is method
// TYpe is command
// Exit points
//
//	Check if the state has been changed as exptected

func (h *Host) SetLogger(l *logrus.Logger) {
	h.logger = l
}

// Unit is method
// TYpe is query
// Exit points
//	Check if random case is selected and right query is returned
//	check if roundrobin case is selected and right query is returned
// check if weightedRoundRobin is selected and right query is returned
// Check if in absence of scheme, new error is returned

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

// Unit is method
// TYpe is both query and command
// Exit points
//	check if all third party are invoked with right parameters in right sequence.
//

func (h *Host) CheckHealth() {
	var healthyServers []Server
	current := h.HealthyServers
	scheme := "http"
	req, _ := http.NewRequest("GET", "", nil)
	client := http.DefaultClient
	req.URL.Path = h.Health
	req.URL.Scheme = scheme
	// it should patch request object and call client.Do for all server list
	for _, server := range h.Servers {
		req.URL.Host = server.Name
		res, err := client.Do(req)
		// check in case of error from mock, logger is invoked
		if err != nil {
			h.logger.Errorln("error in calling health endpoint", err)
			continue
		}
		// it should invoke body.Close by deferring
		defer res.Body.Close()
		// it should append server to healthyserver if response status is 200
		if res.StatusCode == 200 {
			healthyServers = append(healthyServers, server)
		}
	}
	// it should assign new servers to current object
	h.HealthyServers = healthyServers
	// check wether third party is invoked on desired conditions
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

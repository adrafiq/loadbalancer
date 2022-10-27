package proxy

import (
	"net/http"
	"net/http/httptest"
	"reflect"
	"strings"
	"testing"

	"github.com/sirupsen/logrus"
)

func TestHost(t *testing.T) {

	t.Run("it creates a new host with logger", func(t *testing.T) {
		logger := logrus.New()
		host := NewHost(logger)
		if host == nil {
			t.Error("expected host, got nil")
		}
		if host.logger != logger {
			t.Errorf("expected host logger be equal to argument %v", host.logger)
		}
	})
	t.Run("it adds logger to host", func(t *testing.T) {
		var host Host
		logger := logrus.New()
		host.SetLogger(logger)
		if host.logger != logger {
			t.Errorf("expected host logger be equal to argument %v", host.logger)
		}
	})
}

func TestHostGetHealth(t *testing.T) {
	t.Run("it invokes http health check for servers and add them to healthyservers", func(t *testing.T) {

		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
			w.WriteHeader(http.StatusOK)
		}))
		serverAddress := strings.Split(server.URL, "//")[1]
		defer server.Close()
		logger := logrus.New()
		host := Host{
			Name:     "localhost",
			Scheme:   RoundRobin,
			Port:     "3000",
			Health:   "/health",
			Interval: 10,
			Servers: []Server{
				{Name: serverAddress},
				{Name: serverAddress},
			},
			iterator:     1,
			roundSize:    1,
			currentRound: 2,
		}
		host.SetLogger(logger)
		host.CheckHealth()
		if len(host.HealthyServers) < 1 {
			t.Error("healthy servers list should have updated")
		}
		if !(host.iterator == 0 ||
			host.currentRound == 0 ||
			host.roundSize == 0 ||
			reflect.DeepEqual(host.serversProgress, []float32{})) {
			t.Error("the state was not reset, when heatlhservers were updated")
		}

	})
}

func TestGetNext(t *testing.T) {
	t.Run("it returns error if routing scheme is missing", func(t *testing.T) {
		host := Host{}
		if _, err := host.GetNext(); err == nil {
			t.Errorf("should return error in case of missing scheme")
		}
	})
	t.Run("it returns the server specified by iterator and increments it when scheme is rr", func(t *testing.T) {
		host := Host{
			Scheme:   RoundRobin,
			Interval: 10,
			HealthyServers: []Server{
				{Name: "server1"},
				{Name: "server2"},
				{Name: "server3"},
			},
			iterator: 1,
		}
		server, _ := host.GetNext()
		if server != host.HealthyServers[1].Name {
			t.Errorf("should return server from index specified by iterator")
		}
		if host.iterator != 2 {
			t.Errorf("should have incremented iterator")
		}
	})
	t.Run("it returns the server with least progress and add to round progress when scheme is wrr", func(t *testing.T) {
		host := Host{
			Scheme:       WeightedRoundRobin,
			roundSize:    10,
			currentRound: 5,
			HealthyServers: []Server{
				{Name: "server1", Weight: 1},
				{Name: "server2", Weight: 3},
				{Name: "server3", Weight: 6},
			},
			serversProgress: []float32{1, 1, 1},
		}
		server, _ := host.GetNext()
		expectedServer := 2
		if server != host.HealthyServers[expectedServer].Name {
			t.Errorf("should return first server with least progress")
		}
		if host.serversProgress[expectedServer] != 2 {
			t.Errorf("should have incremented progress for returned server")
		}
	})
	t.Run("it returns a server specified by random index within bound when scheme is random)", func(t *testing.T) {
		host := Host{
			Scheme: Random,
			HealthyServers: []Server{
				{Name: "server1"},
				{Name: "server2"},
				{Name: "server3"},
			},
			serversProgress: []float32{1, 1, 1},
		}
		server, _ := host.GetNext()
		serverAsExpected := false
		for _, s := range host.HealthyServers {
			if server == s.Name {
				serverAsExpected = true
			}
		}
		if !serverAsExpected {
			t.Error("server should have been from healthservers")
		}
	})
}

package proxy

import (
	"math/rand"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strings"
	"testing"
	"time"

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
			t.Errorf("attribute logger should be equal to logger passed")
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
		cursor:       1,
		roundSize:    1,
		currentRound: 2,
	}
	host.SetLogger(logger)
	t.Run("it adds servers to healthy list, if healthcheck returns http 200", func(t *testing.T) {
		host.CheckHealth()
		unexpected := 0
		if len(host.HealthyServers) == unexpected {
			t.Error("healthy servers list should have updated")
		}
	})
	t.Run("it resets state when HealthyServers were updated", func(t *testing.T) {
		host.CheckHealth()
		if !(host.cursor == 0 ||
			host.currentRound == 0 ||
			host.roundSize == 0 ||
			reflect.DeepEqual(host.serversProgress, []int{})) {
			t.Error("the state was not reset")
		}
	})
	server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	t.Run("it doesnot add servers for which healthcheck doesnt return http 200", func(t *testing.T) {
		host.CheckHealth()
		host.HealthyServers = []Server{}
		expected := 0
		if len(host.HealthyServers) != expected {
			t.Error("unhealthy servers should'nt be added")
		}
	})
}

func TestHostGetNext(t *testing.T) {
	randInt := func(n int) int {
		return n
	}
	t.Run("it returns error if routing scheme is missing", func(t *testing.T) {
		host := Host{}
		if _, err := host.GetNext(randInt); err == nil {
			t.Errorf("should return error in case of missing scheme")
		}
	})
	t.Run("it returns the server specified by cursor and increments it when scheme is rr", func(t *testing.T) {
		host := Host{
			Scheme: RoundRobin,
			HealthyServers: []Server{
				{Name: "server1"},
				{Name: "server2"},
				{Name: "server3"},
			},
			cursor: 1,
		}
		expectCursor := 1
		server, _ := host.GetNext(randInt)
		if server != host.HealthyServers[expectCursor].Name {
			t.Errorf("should return server from index specified by cursor")
		}
		if host.cursor != 2 {
			t.Errorf("should have incremented cursor")
		}
	})
	t.Run("it resets cursor when RR round is complete", func(t *testing.T) {
		host := Host{
			Scheme: RoundRobin,
			HealthyServers: []Server{
				{Name: "server1"},
				{Name: "server2"},
				{Name: "server3"},
			},
		}
		host.cursor = len(host.HealthyServers)
		_, err := host.GetNext(randInt)
		if err != nil {
			t.Error(err)
		}
		expectedCursor := 1

		if host.cursor != expectedCursor {
			t.Errorf("should have reseted cursor")
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
			serversProgress: []int{1, 1, 1},
		}
		server, _ := host.GetNext(randInt)
		expectedServer := 2
		expectedProgress := 2
		expectedCurrentRound := 6
		if server != host.HealthyServers[expectedServer].Name {
			t.Errorf("should return server with least progress")
		}
		if host.serversProgress[expectedServer] != expectedProgress {
			t.Errorf("should have incremented progress for returned server")
		}
		if host.currentRound != expectedCurrentRound {
			t.Errorf("should have incremented current round")
		}
	})
	t.Run("it resets the round when round is finished", func(t *testing.T) {
		host := Host{
			Scheme:       WeightedRoundRobin,
			roundSize:    10,
			currentRound: 10,
			HealthyServers: []Server{
				{Name: "server1", Weight: 1},
				{Name: "server2", Weight: 3},
				{Name: "server3", Weight: 6},
			},
			serversProgress: []int{1, 1, 1},
		}
		_, err := host.GetNext(randInt)
		if err != nil {
			t.Error(err)
		}
		expectRound := 1
		if host.currentRound != expectRound {
			t.Errorf("should have reset the currentRound")
		}
	})
	t.Run("it returns a server specified by random index", func(t *testing.T) {
		host := Host{
			Scheme: Random,
			HealthyServers: []Server{
				{Name: "server1"},
				{Name: "server2"},
				{Name: "server3"},
			},
			serversProgress: []int{1, 1, 1},
		}
		expectedIndex := 0
		randInt = func(n int) int {
			return expectedIndex
		}
		server, _ := host.GetNext(randInt)
		if server != host.HealthyServers[expectedIndex].Name {
			t.Error("should return server from index specified by randInt")
		}
	})
}

func BenchmarkGetNext(b *testing.B) {
	b.Run("it benchmarks random routing scheme", func(b *testing.B) {
		host := Host{
			Scheme: Random,
			HealthyServers: []Server{
				{Name: "server1"},
				{Name: "server2"},
				{Name: "server3"},
			},
			serversProgress: []int{1, 1, 1},
		}
		rand.Seed(time.Now().Unix())
		b.ResetTimer()
		for n := 0; n < b.N; n++ {
			host.GetNext(rand.Intn)
		}
	})
	b.Run("it benchmarks round robin routing scheme", func(b *testing.B) {
		host := Host{
			Scheme: RoundRobin,
			HealthyServers: []Server{
				{Name: "server1"},
				{Name: "server2"},
				{Name: "server3"},
			},
			serversProgress: []int{1, 1, 1},
		}
		for n := 0; n < b.N; n++ {
			host.GetNext(rand.Intn)
		}
	})
	b.Run("it benchmarks weighted round robin routing scheme", func(b *testing.B) {
		host := Host{
			Scheme: RoundRobin,
			HealthyServers: []Server{
				{Name: "server1", Weight: 1},
				{Name: "server2", Weight: 3},
				{Name: "server3", Weight: 6},
				{Name: "server4", Weight: 1},
				{Name: "server5", Weight: 3},
				{Name: "server6", Weight: 6},
			},
			serversProgress: []int{0, 0, 0},
		}
		for n := 0; n < b.N; n++ {
			host.GetNext(rand.Intn)
		}
	})
}

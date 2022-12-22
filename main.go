package main

import (
	"context"
	cfg "infrastructure/loadbalancer/internals/config"
	log "infrastructure/loadbalancer/internals/log"
	"infrastructure/loadbalancer/proxy"
	"infrastructure/loadbalancer/server"
	"os"
	"os/signal"
	"sync"
	"syscall"

	"math/rand"
	"net/http"
	"time"
)

var config = cfg.NewConfig("config.yaml")
var logger = log.NewLogger(config.GetString("logLevel"))

func main() {
	var hosts []proxy.Host
	config.UnmarshalKey("hosts", &hosts)
	var wg sync.WaitGroup
	var hostsChan []chan os.Signal
	rand.Seed(time.Now().Unix())

	for i := 0; i < len(hosts); i++ {
		exit := make(chan os.Signal, 1)
		signal.Notify(exit, os.Interrupt, syscall.SIGKILL)
		hostsChan = append(hostsChan, exit)

		hosts[i].SetUtils(logger, rand.Intn)
		wg.Add(1)
		go func(h *proxy.Host, exit chan os.Signal) {
			serverChan := make(chan *http.Server)
			go server.StartServer(h, serverChan, logger)
			go server.StartSchedular(h, logger)
			server := <-serverChan
			// Blocking till os.Interrupt
			<-exit
			server.Shutdown(context.Background())
		}(&hosts[i], hostsChan[i])
	}
	wg.Wait()
}

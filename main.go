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
	"time"
)

var config = cfg.NewConfig("config.yaml")
var logger = log.NewLogger(config.GetString("logLevel"))

func main() {
	var hosts []proxy.Host
	config.UnmarshalKey("hosts", &hosts)
	var wg sync.WaitGroup
	rand.Seed(time.Now().Unix())

	for idx, _ := range hosts {
		exit := make(chan os.Signal, 1)
		signal.Notify(exit, os.Interrupt, syscall.SIGKILL)
		hosts[idx].SetUtils(logger, rand.Intn)
		wg.Add(1)
		go func(h *proxy.Host, exit chan os.Signal) {
			server := server.NewServer(h, logger)
			go server.Start()
			go server.ScheduleHealthCheck()
			// GracefulExit: Blocking till os.Interrupt
			<-exit
			server.Instance.Shutdown(context.Background())
		}(&hosts[idx], exit)
	}
	wg.Wait()
}

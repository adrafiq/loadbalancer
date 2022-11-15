package main

import (
	"context"
	cfg "infrastructure/loadbalancer/internals/config"
	log "infrastructure/loadbalancer/internals/log"
	"infrastructure/loadbalancer/proxy"
	"io"
	"os"
	"os/signal"
	"sync"
	"syscall"

	"math/rand"
	"net/http"
	"strings"
	"time"
)

var config = cfg.NewConfig("config.yaml")
var logger = log.NewLogger(config.GetString("logLevel"))

func main() {
	var hosts []proxy.Host
	config.UnmarshalKey("hosts", &hosts)
	var wg sync.WaitGroup
	var hostsChan []chan os.Signal
	for i := 0; i < len(hosts); i++ {
		exit := make(chan os.Signal, 1)
		signal.Notify(exit, os.Interrupt, syscall.SIGKILL)
		hostsChan = append(hostsChan, exit)
		hosts[i].SetLogger(logger)
		wg.Add(1)
		go func(h *proxy.Host, exit chan os.Signal) {
			serverChan := make(chan *http.Server)
			go startServer(h, serverChan)
			go schedular(h)
			server := <-serverChan
			// Blocking till os.Interrupt
			<-exit
			server.Shutdown(context.Background())
		}(&hosts[i], hostsChan[i])
	}
	wg.Wait()
}

func makeHandler(
	host *proxy.Host,
	writeString func(w io.Writer, s string) (n int, err error),
) func(res http.ResponseWriter, req *http.Request) {
	rand.Seed(time.Now().Unix())
	return func(res http.ResponseWriter, req *http.Request) {
		ctx, cancel := context.WithTimeout(req.Context(), host.Timeout*time.Second)
		defer cancel()
		logger.Debugf("Request %+v", req)
		hostName := strings.Split(req.Host, ":")[0]
		if hostName != host.Name {
			res.WriteHeader(403)
			writeString(res, "unrecognized host")
			return
		}
		if len(host.HealthyServers) == 0 {
			res.WriteHeader(503)
			writeString(res, "server not ready. no healthy upstream")
			return
		}
		proxyTarget, err := host.Next(rand.Intn)
		if err != nil {
			logger.Error(err)
			res.WriteHeader(500)
			writeString(res, "internal server error")
		}
		request, _ := http.NewRequestWithContext(ctx, req.Method, "", req.Body)
		request.URL.Host = proxyTarget
		request.URL.Scheme = "http" // only http now
		request.URL.Path = req.URL.Path
		client := http.DefaultClient
		proxyRes, err := client.Do(request)
		if err != nil {
			logger.Error(err)
			res.WriteHeader(403)
			writeString(res, err.Error())
			return
		}
		defer proxyRes.Body.Close()
		logger.Debugf("Response %+v", proxyRes)
		res.WriteHeader(proxyRes.StatusCode)
		io.Copy(res, proxyRes.Body)

	}
}

func schedular(host *proxy.Host) {
	intervals := time.Tick(time.Duration(host.Interval) * time.Second)
	for next := range intervals {
		logger.Debugln("health check interval ", next)
		host.CheckHealth()
	}
}

func startServer(hostConfigured *proxy.Host, serverChan chan *http.Server) {
	var writeString = io.WriteString

	router := http.NewServeMux()
	router.HandleFunc("/", makeHandler(hostConfigured, writeString))
	server := http.Server{
		Addr:    ":" + hostConfigured.Port,
		Handler: router,
	}
	serverChan <- &server
	logger.Infof("Server is starting at %s ", hostConfigured.Port)
	logger.Fatal(server.ListenAndServe())
}

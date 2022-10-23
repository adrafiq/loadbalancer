package main

import (
	cfg "infrastructure/loadbalancer/internals/config"
	log "infrastructure/loadbalancer/internals/log"
	"infrastructure/loadbalancer/proxy"
	"io"
	"net/http"
	"strings"
	"sync"
	"time"
)

var config = cfg.NewConfig("config.yaml")
var logger = log.NewLogger(config.GetString("logLevel"))

func main() {
	var wg sync.WaitGroup
	hostConfigured := proxy.NewHost(logger)
	config.UnmarshalKey("host", &hostConfigured)
	// Make a cron schedular and send this to it
	hostConfigured.CheckHealth()
	// This code piece explictily declares ServeMux and default Server to elaborate internals
	router := http.NewServeMux()
	router.HandleFunc("/", makeHandler(hostConfigured))
	server := http.Server{
		Addr:    ":" + config.GetString("port"),
		Handler: router,
	}
	wg.Add(1)
	go startServer(&server, config.GetInt("port"))
	healthCheckChan := make(chan []proxy.Server)
	wg.Add(1)
	go schedular(hostConfigured.Interval, *hostConfigured, healthCheckChan)
	for updatedServers := range healthCheckChan {
		hostConfigured.HealthyServers = updatedServers
	}
	wg.Wait()
}

func makeHandler(host *proxy.Host) func(res http.ResponseWriter, req *http.Request) {
	return func(res http.ResponseWriter, req *http.Request) {
		logger.Debugf("Request %+v", req)
		if len(host.HealthyServers) == 0 {
			res.WriteHeader(503)
			io.WriteString(res, "server not ready. no healthy upstream")
			return
		}
		hostName := strings.Split(req.Host, ":")[0]
		if hostName != host.Name {
			res.WriteHeader(403)
			io.WriteString(res, "unrecognized host")
			return
		}
		proxyTarget, _ := host.GetNext()
		req.URL.Host = proxyTarget
		req.Host = "39.45.128.173" //Place holder for exteral LB
		req.RequestURI = ""
		client := http.DefaultClient
		proxyRes, err := client.Do(req)
		if err != nil {
			logger.Error(err)
			res.WriteHeader(403)
			io.WriteString(res, err.Error())
			return
		}
		defer proxyRes.Body.Close()
		logger.Debugf("Response %+v", proxyRes)
		res.WriteHeader(proxyRes.StatusCode)
		io.Copy(res, proxyRes.Body)
	}
}

func schedular(intervalSeconds int, host proxy.Host, updateChan chan []proxy.Server) {
	intervals := time.Tick(time.Duration(intervalSeconds) * time.Second)
	for next := range intervals {
		logger.Debugln("health check interval ", next)
		host.CheckHealth()
		updateChan <- host.HealthyServers
	}
}

func startServer(server *http.Server, port int) {
	logger.Infof("Server is starting at %v ", port)
	logger.Fatal(server.ListenAndServe())
}

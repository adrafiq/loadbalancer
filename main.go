package main

import (
	"fmt"
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
	var hosts []proxy.Host
	config.UnmarshalKey("hosts", &hosts)
	var wg sync.WaitGroup
	for _, host := range hosts {
		host.SetLogger(logger)
		wg.Add(1)
		go func(h proxy.Host) {
			wg.Add(1)
			go startServer(&h)
			wg.Add(1)
			go schedular(h.Interval, &h)
		}(host)
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

func schedular(intervalSeconds int, host *proxy.Host) {
	intervals := time.Tick(time.Duration(intervalSeconds) * time.Second)
	for next := range intervals {

		logger.Debugln("health check interval ", next)
		host.CheckHealth()
		fmt.Println("schedular called. health checked")
	}
}

func startServer(hostConfigured *proxy.Host) {
	router := http.NewServeMux()
	router.HandleFunc("/", makeHandler(hostConfigured))
	server := http.Server{
		Addr:    ":" + hostConfigured.Port,
		Handler: router,
	}
	logger.Infof("Server is starting at %s ", hostConfigured.Port)
	logger.Fatal(server.ListenAndServe())
}

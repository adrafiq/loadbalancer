package main

import (
	cfg "infrastructure/loadbalancer/internals/config"
	log "infrastructure/loadbalancer/internals/log"
	"infrastructure/loadbalancer/proxy"
	"io"
	"net/http"
	"strings"
)

var config = cfg.NewConfig("config.yaml")
var logger = log.NewLogger(config.GetString("logLevel"))

func main() {
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
	logger.Infof("Server is starting at %s ", config.GetString("port"))
	logger.Fatal(server.ListenAndServe())
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

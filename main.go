package main

import (
	c "infrastructure/loadbalancer/config"
	log "infrastructure/loadbalancer/log"
	"infrastructure/loadbalancer/proxy"
	"io"
	"net/http"
	"strings"
)

func init() {
	c.InitConfig()
}

var logger = log.NewLogger(c.Config.GetString("logLevel"))

func main() {
	logger := log.NewLogger(c.Config.GetString("logLevel"))
	hostConfigured := proxy.NewHost(logger)
	c.Config.UnmarshalKey("host", &hostConfigured)
	// Make a cron schedular and send this to it
	hostConfigured.CheckHealth()
	// This code piece explictily declares ServeMux and default Server to elaborate internals
	router := http.NewServeMux()
	router.HandleFunc("/", makeHandler(hostConfigured))
	server := http.Server{
		Addr:    ":" + c.Config.GetString("port"),
		Handler: router,
	}
	logger.Infof("Server is starting at %s ", c.Config.GetString("port"))
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

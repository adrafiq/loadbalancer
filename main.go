package main

import (
	"infrastructure/loadbalancer/proxy"
	"infrastructure/loadbalancer/utils"
	"io"
	"net/http"
)

var logger = utils.Logger
var config = utils.Config
var host = proxy.HostConfigured

func main() {
	// This code piece explictily declares ServeMux and default Server to elaborate internals
	logger.Infof("Server is starting at %s ", config.GetString("port"))
	router := http.NewServeMux()
	router.HandleFunc("/", defaultHandler)
	server := http.Server{
		Addr:    ":" + config.GetString("port"),
		Handler: router,
	}
	logger.Fatal(server.ListenAndServe())
}

func defaultHandler(res http.ResponseWriter, req *http.Request) {
	// check if host registered
	// get the proxy url
	// prepare http request and send it
	// copy the proxy response even error to handler response and send back
	io.WriteString(res, "preparing server. please check back in 30 seconds")
	res.WriteHeader(503)
}

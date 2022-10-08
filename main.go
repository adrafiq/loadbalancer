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
	host = proxy.HostConfigured
	// Debug output the request
	// check if host registered
	// get the proxy url
	// prepare http request and send it
	// Debug output the response
	// copy the proxy response even error to handler response and send back
	logger.Debugf("Request %+v", req)
	client := http.DefaultClient
	proxyTarget, _ := host.GetNext()
	req.URL = proxyTarget
	req.Host = "39.45.128.173"
	req.RequestURI = ""
	proxyRes, err := client.Do(req)

	if err != nil {
		logger.Fatal(err)
	} else {
		defer proxyRes.Body.Close()
	}
	logger.Debugf("Response %+v", proxyRes)
	res.WriteHeader(proxyRes.StatusCode)
	io.Copy(res, proxyRes.Body)
}

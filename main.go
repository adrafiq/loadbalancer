package main

import (
	"infrastructure/loadbalancer/proxy"
	"infrastructure/loadbalancer/utils"
	"io"
	"net/http"
	"strings"
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

	logger.Debugf("Request %+v", req)
	host = proxy.HostConfigured
	hostName := strings.Split(req.Host, ":")[0]
	if hostName != host.Name {
		res.WriteHeader(403)
		io.WriteString(res, "unrecognized host")
	}
	proxyTarget, _ := host.GetNext()
	req.URL = proxyTarget
	req.Host = "39.45.128.173" //Place holder for externla LB
	req.RequestURI = ""
	client := http.DefaultClient
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

package main

import (
	"infrastructure/loadbalancer/utils"
	"io"
	"net/http"
)

var logger = utils.Logger
var config = utils.Config

func main() {
	// This code piece delibertly doesnt use default ServeMux and default Server to elaborate internals
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
	io.WriteString(res, "preparing server. please check back in 30 seconds")
	res.WriteHeader(503)
}

package main

import (
	"infrastructure/loadbalancer/utils"
	"io"
	"net/http"
)

var logger = utils.Logger
var config = utils.Config

func main() {
	logger.Infof("Server is starting at %s ", config.GetString("port"))
	logger.Fatal(http.ListenAndServe(":"+config.GetString("port"), defaultHandler()))
}

func defaultHandler() http.HandlerFunc {
	return func(res http.ResponseWriter, req *http.Request) {
		io.WriteString(res, "Default Page")
		res.WriteHeader(200)
	}
}

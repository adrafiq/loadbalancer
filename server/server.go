package server

import (
	"infrastructure/loadbalancer/proxy"
	"io"
	"net/http"
	"time"

	"github.com/sirupsen/logrus"
)

type server struct {
	logger *logrus.Logger
}

func StartSchedular(host *proxy.Host, logger *logrus.Logger) {
	intervals := time.Tick(time.Duration(host.Interval) * time.Second)
	for next := range intervals {
		logger.Debugln("health check interval ", next)
		host.CheckHealth()
	}
}

func StartServer(hostConfigured *proxy.Host, serverChan chan *http.Server, logger *logrus.Logger) {
	var writeString = io.WriteString

	router := http.NewServeMux()
	router.HandleFunc("/", makeHandler(hostConfigured, writeString, logger))
	server := http.Server{
		Addr:    ":" + hostConfigured.Port,
		Handler: router,
	}
	serverChan <- &server
	logger.Infof("Server is starting at %s ", hostConfigured.Port)
	logger.Fatal(server.ListenAndServe())
}

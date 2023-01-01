package server

import (
	"infrastructure/loadbalancer/proxy"
	"io"
	"net/http"
	"time"

	"github.com/sirupsen/logrus"
)

type Server struct {
	host     *proxy.Host
	logger   *logrus.Logger
	Instance *http.Server
}

func NewServer(host *proxy.Host, logger *logrus.Logger) Server {
	server := Server{
		host:   host,
		logger: logger,
		Instance: &http.Server{
			Addr: ":" + host.Port,
		},
	}
	return server
}

func (s *Server) ScheduleHealthCheck() {
	intervals := time.Tick(time.Duration(s.host.Interval) * time.Second)
	for next := range intervals {
		s.logger.Debugln("health check interval ", next)
		s.host.CheckHealth()
	}
}

func (s *Server) Start() {
	var writeString = io.WriteString
	router := http.NewServeMux()
	router.HandleFunc("/", makeHandler(s.host, writeString, s.logger))
	s.Instance.Handler = router
	s.logger.Infof("Server is starting at %s ", s.host.Port)
	s.logger.Fatal(s.Instance.ListenAndServe())
}

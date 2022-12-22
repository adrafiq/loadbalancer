package server

import (
	"context"
	"infrastructure/loadbalancer/proxy"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/sirupsen/logrus"
)

func makeHandler(
	host *proxy.Host,
	writeString func(w io.Writer, s string) (n int, err error),
	logger *logrus.Logger,
) func(res http.ResponseWriter, req *http.Request) {
	return func(res http.ResponseWriter, req *http.Request) {
		ctx, cancel := context.WithTimeout(req.Context(), host.Timeout*time.Second)
		defer cancel()
		logger.Debugf("Request %+v", req)
		hostName := strings.Split(req.Host, ":")[0]
		if hostName != host.Name {
			res.WriteHeader(403)
			writeString(res, "unrecognized host")
			return
		}
		if len(host.HealthyServers) == 0 {
			res.WriteHeader(503)
			writeString(res, "server not ready. no healthy upstream")
			return
		}
		proxyTarget, err := host.Next()
		if err != nil {
			logger.Error(err)
			res.WriteHeader(500)
			writeString(res, "internal server error")
		}
		request, _ := http.NewRequestWithContext(ctx, req.Method, "", req.Body)
		request.URL.Host = proxyTarget
		request.URL.Scheme = "http" // only http now
		request.URL.Path = req.URL.Path
		client := http.DefaultClient
		proxyRes, err := client.Do(request)
		if err != nil {
			logger.Error(err)
			res.WriteHeader(403)
			writeString(res, err.Error())
			return
		}
		defer proxyRes.Body.Close()
		logger.Debugf("Response %+v", proxyRes)
		res.WriteHeader(proxyRes.StatusCode)
		io.Copy(res, proxyRes.Body)

	}
}

package main

import (
	"infrastructure/loadbalancer/proxy"
	"infrastructure/loadbalancer/utils"
	"io"
	"net/http"
	"strings"
)

func init() {
	utils.InitConfig()
	utils.InitLogger()
}

func main() {
	// This code piece explictily declares ServeMux and default Server to elaborate internals

	hostConfigured := proxy.NewHost()
	utils.Config.UnmarshalKey("host", &hostConfigured)
	// Make a cron schedular and send this to it
	hostConfigured.CheckHealth()
	router := http.NewServeMux()
	router.HandleFunc("/", makeHandler(hostConfigured))
	server := http.Server{
		Addr:    ":" + utils.Config.GetString("port"),
		Handler: router,
	}
	utils.Logger.Infof("Server is starting at %s ", utils.Config.GetString("port"))
	utils.Logger.Fatal(server.ListenAndServe())
}

func makeHandler(host *proxy.Host) func(res http.ResponseWriter, req *http.Request) {
	return func(res http.ResponseWriter, req *http.Request) {
		utils.Logger.Debugf("Request %+v", req)
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
			utils.Logger.Error(err)
			res.WriteHeader(403)
			io.WriteString(res, err.Error())
			return
		}
		defer proxyRes.Body.Close()
		utils.Logger.Debugf("Response %+v", proxyRes)
		res.WriteHeader(proxyRes.StatusCode)
		io.Copy(res, proxyRes.Body)
	}
}

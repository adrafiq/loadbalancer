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
	proxy.InitHost()
	// Make a cron schedular and send this to it
	proxy.HostConfigured.UpdateHealthyServer()
}

func main() {
	// This code piece explictily declares ServeMux and default Server to elaborate internals
	utils.Logger.Infof("Server is starting at %s ", utils.Config.GetString("port"))
	router := http.NewServeMux()
	router.HandleFunc("/", defaultHandler)
	server := http.Server{
		Addr:    ":" + utils.Config.GetString("port"),
		Handler: router,
	}
	utils.Logger.Fatal(server.ListenAndServe())
}

func defaultHandler(res http.ResponseWriter, req *http.Request) {
	utils.Logger.Debugf("Request %+v", req)
	if len(proxy.HostConfigured.HealthyServers) == 0 {
		res.WriteHeader(503)
		io.WriteString(res, "server not ready. no healthy upstream")
		return
	}
	hostName := strings.Split(req.Host, ":")[0]
	if hostName != proxy.HostConfigured.Name {
		res.WriteHeader(403)
		io.WriteString(res, "unrecognized host")
		return
	}
	proxyTarget, _ := proxy.HostConfigured.GetNext()
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

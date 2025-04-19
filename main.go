package main

import (
	"fmt"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
)

type RouteConfig struct {
	Host   string
	Target string
	Port   string
}

type ProxyHandler struct {
	routes map[string]*httputil.ReverseProxy
}

const (
	HOST_LOCAL  = "192.168.254.155"
	PORT_CAMPUS = "8014"
)

func main() {
	// Define all your routes here
	routes := []RouteConfig{
		{Host: "app.gopher.wtf", Target: HOST_LOCAL, Port: "8015"},
		//ww.gopher.wtf", Target: HOST_LOCAL, Port: PORT_CAMPUS},
		{Host: "200.37.144.19", Target: HOST_LOCAL, Port: PORT_CAMPUS},
		{Host: HOST_LOCAL, Target: HOST_LOCAL, Port: PORT_CAMPUS},
	}

	// Create proxy handler with routes
	proxyHandler := &ProxyHandler{
		routes: make(map[string]*httputil.ReverseProxy),
	}

	// Configure all routes
	for _, route := range routes {
		target, err := url.Parse(fmt.Sprintf("http://%s:%s", route.Target, route.Port))
		if err != nil {
			panic(err)
		}
		proxyHandler.routes[route.Host] = httputil.NewSingleHostReverseProxy(target)
	}

	http.Handle("/", proxyHandler)
	err := http.ListenAndServe(":80", nil)
	if err != nil {
		panic(err)
	}
}

func (ph *ProxyHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	log.Printf("Incoming request: %s, Host: %s", r.URL, r.Host)

	proxy, exists := ph.routes[r.Host]
	if !exists {
		http.Error(w, "Not Found", http.StatusNotFound)
		return
	}

	proxy.ErrorHandler = func(w http.ResponseWriter, r *http.Request, err error) {
		log.Printf("Proxy error: %v", err)
		http.Error(w, "Proxy Error", http.StatusBadGateway)
	}

	proxy.ServeHTTP(w, r)
}

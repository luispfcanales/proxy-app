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
	HOST_LOCAL = "172.31.91.58"
	// PORT_CAMPUS = "8001"
	CONFERICIS_PORT     = "8001"
	API_CONFERICIS_PORT = "8002"
)

func main() {
	// Define all your routes here
	routes := []RouteConfig{
		// {Host: "app.gopher.wtf", Target: HOST_LOCAL, Port: "8015"},
		{Host: "api.luispf.org", Target: HOST_LOCAL, Port: API_CONFERICIS_PORT},
		{Host: "confericis.luispf.org", Target: HOST_LOCAL, Port: CONFERICIS_PORT},
		{Host: "35.173.114.173", Target: HOST_LOCAL, Port: CONFERICIS_PORT},
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
		proxy := httputil.NewSingleHostReverseProxy(target)

		// Add headers for Cloudflare
		originalDirector := proxy.Director
		proxy.Director = func(req *http.Request) {
			originalDirector(req)
			req.Header.Set("X-Forwarded-Host", req.Host)
			req.Header.Set("X-Forwarded-Proto", "https")
			req.Header.Set("X-Real-IP", req.RemoteAddr)
		}

		proxyHandler.routes[route.Host] = proxy
	}

	http.Handle("/", proxyHandler)
	log.Printf("Starting server on :80")
	if err := http.ListenAndServe(":80", nil); err != nil {
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

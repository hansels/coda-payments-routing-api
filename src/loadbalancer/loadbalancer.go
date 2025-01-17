package loadbalancer

import (
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"sync"
	"time"
)

type Server struct {
	URL     *url.URL
	healthy bool
}

type LoadBalancer struct {
	servers []*Server
	current int
	mx      sync.Mutex
}

func NewLoadBalancer(servers []string) *LoadBalancer {
	serverList := make([]*Server, len(servers))
	for i, server := range servers {
		urlData, err := url.Parse(server)
		if err != nil {
			log.Println("Failed to parse server URL: ", err)
			continue
		}

		serverList[i] = &Server{URL: urlData, healthy: true}
	}

	lb := &LoadBalancer{servers: serverList}

	// Initiate Health Check Function
	go lb.healthCheck()

	return lb
}

func (lb *LoadBalancer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	target := lb.getNextServer()
	if target == nil {
		http.Error(w, "No healthy servers available", http.StatusServiceUnavailable)
		return
	}

	log.Printf("Forwarding request to %s", target)
	proxy := httputil.NewSingleHostReverseProxy(target)
	proxy.ServeHTTP(w, r)
}

func (lb *LoadBalancer) healthCheck() {
	const maxRetries = 3

	for {
		for _, server := range lb.servers {
			healthy := false
			slowResponses := 0

			for i := 0; i < maxRetries; i++ {
				start := time.Now()
				resp, err := http.Get(server.URL.String() + "/ping")
				responseTime := time.Since(start)

				if err == nil && resp.StatusCode == http.StatusOK && responseTime <= 5*time.Second {
					healthy = true
					resp.Body.Close()
					break
				}

				if responseTime > 5*time.Second {
					slowResponses++
				}

				if resp != nil {
					resp.Body.Close()
				}

				time.Sleep(1 * time.Second)
			}

			if healthy && slowResponses < maxRetries {
				server.healthy = true
				log.Printf("Server %s is healthy", server.URL)
			} else {
				server.healthy = false
				log.Printf("Server %s is unhealthy", server.URL)
			}
		}

		time.Sleep(30 * time.Second)
	}
}

func (lb *LoadBalancer) getNextServer() *url.URL {
	count := 0
	for {
		if count >= len(lb.servers) {
			return nil
		}

		lb.current++
		server := lb.servers[lb.current%len(lb.servers)]
		if server.healthy {
			return server.URL
		}

		count++
	}
}

func (lb *LoadBalancer) RegisterServer(serverURL string) {
	lb.mx.Lock()
	defer lb.mx.Unlock()

	url, err := url.Parse(serverURL)
	if err != nil {
		log.Printf("Failed to parse server URL: %v", err)
		return
	}

	lb.servers = append(lb.servers, &Server{URL: url, healthy: true})

	log.Printf("Server %s registered", serverURL)
}

func (lb *LoadBalancer) UnregisterServer(serverURL string) {
	lb.mx.Lock()
	defer lb.mx.Unlock()

	for i, server := range lb.servers {
		if server.URL.String() == serverURL {
			lb.servers = append(lb.servers[:i], lb.servers[i+1:]...)
			log.Printf("Server %s unregistered", serverURL)
			return
		}
	}

	log.Printf("Server %s not found", serverURL)
}

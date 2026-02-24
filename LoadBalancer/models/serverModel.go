package models

import (
	"log"
	"net/http"
	"net/http/httputil"
	"sync"
	"time"
)

type SimpleServer struct {
	Addr  string
	Proxy *httputil.ReverseProxy
	Alive bool
	Mux   sync.RWMutex
}

type LoadBalancer struct {
	Port            string
	RoundRobinCount int
	Servers         []*SimpleServer
}

func (s *SimpleServer) SetAlive(alive bool) {
	s.Mux.Lock() // Mutex lock to ensure thread safety when updating the Alive status
	s.Alive = alive
	s.Mux.Unlock()
}

// IsAlive checks if the server is healthy *_*
func (s *SimpleServer) IsAlive() bool {
	s.Mux.RLock()
	defer s.Mux.RUnlock()
	return s.Alive
}

func isServerResponsive(addr string) bool {
	conn, err := http.Get(addr + "/users/productview") // Pinging a public route that doesn't require authentication
	if err != nil || conn.StatusCode != http.StatusOK {
		return false
	}
	return true
}

// Picking next alive serve
// Used Round Robin here
// Will try to implement other algos too
func (lb *LoadBalancer) GetNextServer() *SimpleServer {
	for i := 0; i < len(lb.Servers); i++ {
		idx := lb.RoundRobinCount % len(lb.Servers)
		lb.RoundRobinCount++

		if lb.Servers[idx].IsAlive() {
			return lb.Servers[idx]
		}
	}
	return nil // Man down!
}

func (lb *LoadBalancer) ServeProxy(w http.ResponseWriter, r *http.Request) {
	targetServer := lb.GetNextServer()
	if targetServer == nil {
		http.Error(w, "Service Unavailable", http.StatusServiceUnavailable)
		return
	}
	start := time.Now()
	targetServer.Proxy.ServeHTTP(w, r)
	duration := time.Since(start)
	log.Printf("[LoadBalanced] %s %s -> %s | Duration: %v",
		r.Method,
		r.URL.Path,
		targetServer.Addr,
		duration,
	)
}

func (lb *LoadBalancer) HealthCheck() {
	for {
		for _, s := range lb.Servers {
			alive := isServerResponsive(s.Addr)
			s.SetAlive(alive)
			if !alive {
				log.Printf("Server %s is DOWN", s.Addr)
			}
		}
		time.Sleep(10 * time.Second) // Check health every 10 seconds
	}
}

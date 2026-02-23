package main

import (
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"

	"github.com/Bhanubpsn/e-commerce-backend/loadBalancer/models"
	"github.com/joho/godotenv"
)

func NewLoadBalancer(port string, serverList []string) *models.LoadBalancer {
	var servers []*models.SimpleServer
	for _, addr := range serverList {
		serverUrl, _ := url.Parse(addr)
		s := &models.SimpleServer{
			Addr:  addr,
			Proxy: httputil.NewSingleHostReverseProxy(serverUrl),
			Alive: true,
		}
		servers = append(servers, s)
	}
	return &models.LoadBalancer{
		Port:            port,
		RoundRobinCount: 0,
		Servers:         servers,
	}
}

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}
	port1 := os.Getenv("PORT1")
	port2 := os.Getenv("PORT2")
	port3 := os.Getenv("PORT3")
	loadbalancerport := os.Getenv("LOAD_BALANCER_PORT")
	servers := []string{"http://localhost:" + port1, "http://localhost:" + port2, "http://localhost:" + port3}
	lb := NewLoadBalancer(loadbalancerport, servers)

	log.Printf("Load Balancer started at :%s\n", loadbalancerport)

	// Go routine to periodically check server health it will not block the main thread
	go lb.HealthCheck()

	http.HandleFunc("/", lb.ServeProxy)
	log.Fatal(http.ListenAndServe(":"+loadbalancerport, nil))
}

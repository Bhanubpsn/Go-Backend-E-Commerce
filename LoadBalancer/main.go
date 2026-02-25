package main

import (
	"bufio"
	"log"
	"net"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"

	"github.com/Bhanubpsn/e-commerce-backend/loadBalancer/models"
	"github.com/joho/godotenv"
)

func NewLimiterClient(addr string) *models.LimiterClient {
	limiterConn, err := net.Dial("tcp", addr)
	if err != nil {
		log.Fatal("Could not connect to TCP Limiter")
	}
	limiterReader := bufio.NewReader(limiterConn)
	return &models.LimiterClient{Conn: limiterConn, Reader: limiterReader}
}

func NewLoadBalancer(port string, serverList []string, lc *models.LimiterClient) *models.LoadBalancer {
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
		Limiter:         lc,
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
	rateLimiterPort := os.Getenv("RATE_LIMITER_PORT")
	loadbalancerport := os.Getenv("LOAD_BALANCER_PORT")

	servers := []string{"http://localhost:" + port1, "http://localhost:" + port2, "http://localhost:" + port3}

	limiterClient := NewLimiterClient("localhost:" + rateLimiterPort)
	lb := NewLoadBalancer(loadbalancerport, servers, limiterClient)

	log.Printf("Load Balancer started at :%s\n", loadbalancerport)

	// Go routine to periodically check server health it will not block the main thread
	go lb.HealthCheck()

	http.HandleFunc("/", lb.ServeProxy)
	log.Fatal(http.ListenAndServe(":"+loadbalancerport, nil))
}

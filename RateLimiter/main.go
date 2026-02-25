package main

import (
	"bufio"
	"fmt"
	"log"
	"net"
	"os"
	"sync"
	"time"

	"github.com/joho/godotenv"
)

// For now using Token Bucket algorithm for rate limiting, can be extended to leaky bucket or fixed window counter as needed
// @TODO: Make an interface for different rate limiting algorithms and implement them as needed
type TokenBucket struct {
	tokens     float64
	lastRefill time.Time
	mutexLock  sync.Mutex
}

var (
	buckets = make(map[string]*TokenBucket)
	mu      sync.RWMutex
	rate    = 25.0 // 25 requests per second per IP
	burst   = 10.0 // 10 tokens max in the bucket, allows for short bursts of traffic
)

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}
	PORT := os.Getenv("PORT")
	listener, _ := net.Listen("tcp", ":"+PORT)
	fmt.Println("TCP Rate Limiter active on :" + PORT)

	for {
		conn, _ := listener.Accept()
		go handleConnection(conn) // Handle multiple LB connections
	}
}

func handleConnection(conn net.Conn) {
	defer conn.Close()
	scanner := bufio.NewScanner(conn)
	for scanner.Scan() {
		ip := scanner.Text()

		mu.RLock()
		bucket, exists := buckets[ip]
		mu.RUnlock()

		if !exists {
			mu.Lock()
			bucket = &TokenBucket{tokens: burst, lastRefill: time.Now()}
			buckets[ip] = bucket
			mu.Unlock()
		}

		if allow(bucket) {
			conn.Write([]byte("1\n")) // 1 = Allow
		} else {
			conn.Write([]byte("0\n")) // 0 = Deny
		}
	}
}

func allow(b *TokenBucket) bool {
	b.mutexLock.Lock()
	defer b.mutexLock.Unlock()
	now := time.Now()
	b.tokens += now.Sub(b.lastRefill).Seconds() * rate
	if b.tokens > burst {
		b.tokens = burst
	}
	b.lastRefill = now
	if b.tokens >= 1 {
		b.tokens--
		return true
	}
	return false
}

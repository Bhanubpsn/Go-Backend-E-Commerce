package main

import (
	"log"
	"os"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
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
	rate    = 25.0 // 25 requests per second
	burst   = 10.0 // Max 10 tokens
)

func main() {
	router := gin.New()

	router.GET("/check", func(c *gin.Context) {
		ip := c.Query("ip")
		log.Printf("Received rate limit check for IP: %s", ip)
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
			c.Status(200) // OK: Request allowed
		} else {
			c.Status(429) // Too Many Requests
		}
	})
	router.Run(":" + os.Getenv("PORT"))
}

func allow(b *TokenBucket) bool {
	b.mutexLock.Lock()
	defer b.mutexLock.Unlock()

	now := time.Now()
	elapsed := now.Sub(b.lastRefill).Seconds()
	b.tokens += elapsed * rate

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

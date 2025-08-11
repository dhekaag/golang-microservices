package gateway

import (
	"net/http"
	"sync"
	"time"

	"github.com/dhekaag/golang-microservices/shared/pkg/utils"
)

type RateLimiter struct {
	clients map[string]*Client
	mutex   sync.RWMutex
	limit   int
	window  time.Duration
}

type Client struct {
	requests []time.Time
	mutex    sync.Mutex
}

type RateLimitConfig struct {
	RequestsPerMinute int
	WindowSize        time.Duration
}

func NewRateLimiter(config RateLimitConfig) *RateLimiter {
	return &RateLimiter{
		clients: make(map[string]*Client),
		limit:   config.RequestsPerMinute,
		window:  config.WindowSize,
	}
}

func RateLimit(next http.Handler, config RateLimitConfig) http.Handler {
	limiter := NewRateLimiter(config)

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		clientIP := getClientIP(r)

		if !limiter.Allow(clientIP) {
			w.Header().Set("X-RateLimit-Limit", string(rune(config.RequestsPerMinute)))
			w.Header().Set("X-RateLimit-Remaining", "0")
			w.Header().Set("Retry-After", "60")
			utils.SendError(w, http.StatusTooManyRequests, "Rate limit exceeded")
			return
		}

		next.ServeHTTP(w, r)
	})
}

func (rl *RateLimiter) Allow(clientID string) bool {
	rl.mutex.Lock()
	defer rl.mutex.Unlock()

	client, exists := rl.clients[clientID]
	if !exists {
		client = &Client{requests: make([]time.Time, 0)}
		rl.clients[clientID] = client
	}

	client.mutex.Lock()
	defer client.mutex.Unlock()

	now := time.Now()

	// Remove old requests outside the window
	cutoff := now.Add(-rl.window)
	newRequests := make([]time.Time, 0)
	for _, req := range client.requests {
		if req.After(cutoff) {
			newRequests = append(newRequests, req)
		}
	}
	client.requests = newRequests

	// Check if we can accept new request
	if len(client.requests) >= rl.limit {
		return false
	}

	// Add current request
	client.requests = append(client.requests, now)
	return true
}

func getClientIP(r *http.Request) string {
	// Check X-Forwarded-For header
	forwarded := r.Header.Get("X-Forwarded-For")
	if forwarded != "" {
		return forwarded
	}

	// Check X-Real-IP header
	realIP := r.Header.Get("X-Real-IP")
	if realIP != "" {
		return realIP
	}

	// Use remote address
	return r.RemoteAddr
}

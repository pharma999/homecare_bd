package middleware

import (
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"golang.org/x/time/rate"
)

// ipLimiter stores per-IP rate limiters.
type ipLimiter struct {
	limiter  *rate.Limiter
	lastSeen time.Time
}

type rateLimitStore struct {
	mu       sync.Mutex
	visitors map[string]*ipLimiter
	r        rate.Limit
	b        int
}

func newStore(r rate.Limit, b int) *rateLimitStore {
	s := &rateLimitStore{
		visitors: make(map[string]*ipLimiter),
		r:        r,
		b:        b,
	}
	// Prune stale entries every 5 minutes
	go func() {
		for range time.Tick(5 * time.Minute) {
			s.mu.Lock()
			for ip, v := range s.visitors {
				if time.Since(v.lastSeen) > 10*time.Minute {
					delete(s.visitors, ip)
				}
			}
			s.mu.Unlock()
		}
	}()
	return s
}

func (s *rateLimitStore) get(ip string) *rate.Limiter {
	s.mu.Lock()
	defer s.mu.Unlock()
	v, ok := s.visitors[ip]
	if !ok {
		v = &ipLimiter{limiter: rate.NewLimiter(s.r, s.b)}
		s.visitors[ip] = v
	}
	v.lastSeen = time.Now()
	return v.limiter
}

// OTPRateLimit allows at most 5 OTP requests per IP per minute.
var otpStore = newStore(rate.Every(time.Minute/5), 5)

func OTPRateLimit() gin.HandlerFunc {
	return func(c *gin.Context) {
		ip := c.ClientIP()
		if !otpStore.get(ip).Allow() {
			c.JSON(http.StatusTooManyRequests, gin.H{
				"status":  429,
				"message": "Too many OTP requests. Please wait a minute before trying again.",
				"error":   "rate_limited",
			})
			c.Abort()
			return
		}
		c.Next()
	}
}

// APIRateLimit allows 120 requests per IP per minute for general API use.
var apiStore = newStore(rate.Every(time.Minute/120), 120)

func APIRateLimit() gin.HandlerFunc {
	return func(c *gin.Context) {
		ip := c.ClientIP()
		if !apiStore.get(ip).Allow() {
			c.JSON(http.StatusTooManyRequests, gin.H{
				"status":  429,
				"message": "Too many requests. Please slow down.",
				"error":   "rate_limited",
			})
			c.Abort()
			return
		}
		c.Next()
	}
}

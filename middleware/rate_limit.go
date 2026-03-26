package middleware

import (
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

type ipEntry struct {
	count     int
	windowEnd time.Time
	mu        sync.Mutex
}

var ipStore sync.Map

// RateLimit limits requests per IP to `max` per `window` duration.
func RateLimit(max int, window time.Duration) gin.HandlerFunc {
	return func(c *gin.Context) {
		ip := c.ClientIP()

		val, _ := ipStore.LoadOrStore(ip, &ipEntry{windowEnd: time.Now().Add(window)})
		entry := val.(*ipEntry)

		entry.mu.Lock()
		now := time.Now()
		if now.After(entry.windowEnd) {
			entry.count = 0
			entry.windowEnd = now.Add(window)
		}
		entry.count++
		count := entry.count
		entry.mu.Unlock()

		if count > max {
			c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{
				"error": "Demasiadas solicitudes. Intentá de nuevo en unos minutos.",
			})
			return
		}

		c.Next()
	}
}

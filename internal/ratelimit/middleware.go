package ratelimit

import (
	"net/http"
	"strings"
)

func (rl *RateLimiter) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Получаем IP клиента (упрощенно)
		clientIP := strings.Split(r.RemoteAddr, ":")[0]

		// Проверяем лимит
		if !rl.Allow(clientIP) {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusTooManyRequests)
			w.Write([]byte(`{"error": "rate limit exceeded"}`))
			return
		}

		next.ServeHTTP(w, r)
	})
}

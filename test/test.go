package proxy_test

import (
	"balancer/internal/balancer"
	"balancer/internal/proxy"
	"balancer/internal/ratelimit"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

// Тестируем Round Robin балансировку
func TestLoadBalancer_RoundRobin(t *testing.T) {
	// Создаем фальшивые бэкенды с помощью httptest
	backend1 := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("response from backend 1"))
	}))
	defer backend1.Close()

	backend2 := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("response from backend 2"))
	}))
	defer backend2.Close()

	// Список бэкендов
	backends := []*balancer.Backend{
		{Url: backend1.URL, State: true},
		{Url: backend2.URL, State: true},
	}

	// Балансировщик Round Robin
	b := balancer.NewRR(backends)

	// Создаем rate limiter (по умолчанию 1 запрос в секунду)
	rateLimiter := ratelimit.NewRateLimiter(10, time.Second)

	// Создаем экземпляр load balancer
	lb := proxy.NewLoadBalancer(b, rateLimiter)

	// Тестируем, что балансировщик перенаправляет запросы к бэкендам по очереди
	tests := []struct {
		name        string
		expectedURL string
	}{
		{"request to backend 1", "response from backend 1"},
		{"request to backend 2", "response from backend 2"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Создаем запрос
			req, err := http.NewRequest("GET", "http://localhost:8080", nil)
			if err != nil {
				t.Fatal(err)
			}

			// Отправляем запрос через LoadBalancer
			rr := httptest.NewRecorder()
			lb.ServeHTTP(rr, req)

			// Проверяем, что балансировщик перенаправил запрос к нужному бэкенду
			if rr.Body.String() != tt.expectedURL {
				t.Errorf("expected %v, got %v", tt.expectedURL, rr.Body.String())
			}
		})
	}
}

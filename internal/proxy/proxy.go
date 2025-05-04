package proxy

import (
	"balancer/internal/balancer"
	"balancer/internal/ratelimit"
	"balancer/pkg/logger"
	"context"
	"net"
	"net/http"
	"net/http/httputil"
	"net/url"
)

type LoadBalancer struct {
	balancer     balancer.Balancer
	server       *http.Server
	reverseProxy *httputil.ReverseProxy
	rateManager  *ratelimit.RateLimiter
}

type contextKey string

const backendKey contextKey = "selected-backend"

func NewLoadBalancer(balance balancer.Balancer, rm *ratelimit.RateLimiter) *LoadBalancer {
	lb := &LoadBalancer{
		balancer:    balance,
		rateManager: rm,
	}
	logger.Logger.Info().Msg("создан новый LoadBalancer")
	lb.reverseProxy = &httputil.ReverseProxy{
		Director: func(req *http.Request) {
			backend := lb.balancer.NextBackendRR()
			if backend == nil {
				return
			}

			backend.IncrementConn()

			logger.Logger.Info().Msgf("выбран бэкенд: %s", backend.Url)
			// Сохраняем бэкенд в контексте запроса
			ctx := context.WithValue(req.Context(), backendKey, backend)
			*req = *req.WithContext(ctx)

			backendUrl, err := url.Parse(backend.Url)
			if err != nil {
				logger.Logger.Err(err).Msg("ошибка парсинга URL бэкенда")
				backend.DecrementConn()
				return
			}

			req.URL.Scheme = backendUrl.Scheme
			req.URL.Host = backendUrl.Host
			req.Host = backendUrl.Host
			req.RequestURI = ""

			req.Header.Set("X-Forwarded-For", req.RemoteAddr)
			req.Header.Set("X-Forwarded-Host", req.Host)
			req.Header.Set("X-Forwarded-Proto", req.URL.Scheme)
			logger.Logger.Info().Msgf("перенаправляем запрос на бэкенд: %s", backend.Url)
		},
		ErrorHandler: func(w http.ResponseWriter, r *http.Request, err error) {
			w.WriteHeader(http.StatusBadGateway)
			_, _ = w.Write([]byte("ошибка прокси: " + err.Error()))

			logger.Logger.Err(err).Msg("ошибка прокси")
			// Декремент соединений даже в случае ошибки
			if backendVal := r.Context().Value(backendKey); backendVal != nil {
				if backend, ok := backendVal.(*balancer.Backend); ok {
					backend.DecrementConn()
				}
			}
		},
		ModifyResponse: func(resp *http.Response) error {
			// Уменьшаем счётчик после получения ответа
			if backendVal := resp.Request.Context().Value(backendKey); backendVal != nil {
				if backend, ok := backendVal.(*balancer.Backend); ok {
					backend.DecrementConn()
				}
			}
			return nil
		},
	}
	logger.Logger.Info().Msg("создан новый ReverseProxy")
	return lb
}

func (lb *LoadBalancer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	clientIP, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		clientIP = r.RemoteAddr
	}
	logger.Logger.Info().Msgf("получен запрос от клиента: %s", clientIP)
	if lb.rateManager != nil && !lb.rateManager.Allow(clientIP) {
		w.WriteHeader(http.StatusTooManyRequests)
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("Retry-After", "60")
		_, _ = w.Write([]byte("too many requests. Please, try later"))
		return
	}

	lb.reverseProxy.ServeHTTP(w, r)
}

func (lb *LoadBalancer) Start(addr string) *http.Server {
	lb.server = &http.Server{
		Addr:    addr,
		Handler: lb,
	}

	go func() {
		if err := lb.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Logger.Err(err).Msg("ошибка запуска сервера")
		}
	}()

	return lb.server
}

func (lb *LoadBalancer) Shutdown(ctx context.Context) error {
	logger.Logger.Info().Msg("остановка сервера")
	return lb.server.Shutdown(ctx)
}

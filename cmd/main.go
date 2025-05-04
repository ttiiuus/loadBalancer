package main

import (
	"balancer/internal/balancer"
	"balancer/internal/config"
	"balancer/internal/health"
	"balancer/internal/proxy"
	"balancer/internal/ratelimit"
	"balancer/pkg/logger"
	"context"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func main() {
	configPath := flag.String("config", "configs/config.json", "путь к конфигурационному файлу")
	flag.Parse()

	logcfg := logger.Config{
		Level:  "info",
		Pretty: true,
	}
	logger.InitGlobalLogger(&logcfg)

	cfg, err := config.LoadConfig(*configPath)

	if err != nil {
		logger.Logger.Err(err).Msg("ошибка загрузки конфигурации")
		os.Exit(1)
	}

	var backends []*balancer.Backend
	for _, backendCfg := range cfg.Backends {
		backend := &balancer.Backend{
			Url:   backendCfg.Url,
			State: true,
		}
		backends = append(backends, backend)
		back := fmt.Sprintf("добавлен бэкенд : %s", backend.Url)
		logger.Logger.Info().Msg(back)
	}
	b := balancer.NewRR(backends)
	//используем алгоритм балансировки round robin

	rateLimiter := ratelimit.NewRateLimiter(cfg.RateLim.Cap, cfg.RateLim.Rate)
	rl := fmt.Sprintf("включили rate limiting. Стандартный лимит %d запросов в секунду", cfg.RateLim.Rate)
	logger.Logger.Info().Msg(rl)

	healtChecker := health.NewChecker(b, cfg.HealthCheck.Interval, cfg.HealthCheck.Timeout)
	healtChecker.Start()
	//запустили хелсчекер

	//запускаем прокси
	proxy := proxy.NewLoadBalancer(b, rateLimiter)

	//запускаем сервер
	serverAddr := fmt.Sprintf(":%d", cfg.Server.Number)
	server := proxy.Start(serverAddr)
	startsrv := fmt.Sprintf("logger started on %s", serverAddr)
	logger.Logger.Info().Msg(startsrv)

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	logger.Logger.Info().Msg("get sig stop, starting graceful shutdown")

	healtChecker.Stop()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		logger.Logger.Err(err).Msg("ошибка при остановке сервера")
		//log.Error("Ошибка при остановке сервера:", err)
	}
	logger.Logger.Info().Msg("сервер остановлен")

}

package health

import (
	"balancer/internal/balancer"
	"balancer/pkg/logger"
	"net/http"
	"net/url"
	"sync"
	"time"
)

type Checker struct {
	balancer      balancer.Balancer
	CheckInterval time.Duration
	timeout       time.Duration
	stopCh        chan struct{}
	wg            sync.WaitGroup
	httpClient    *http.Client
}

func NewChecker(balancer balancer.Balancer, checkInterval, timeout time.Duration) *Checker {
	httpClient := &http.Client{
		Timeout: timeout,
	}
	logger.Logger.Info().Msgf("запускаем health checker с интервалом %s", checkInterval.String())
	logger.Logger.Info().Msgf("таймаут health checker %s", timeout.String())

	return &Checker{
		balancer:      balancer,
		CheckInterval: checkInterval,
		timeout:       timeout,
		stopCh:        make(chan struct{}),
		httpClient:    httpClient,
	}
}

func (c *Checker) Start() {
	// Запускаем горутину для периодической проверки бэкендов
	// и добавляем её в WaitGroup
	logger.Logger.Info().Msg("запускаем health checker")
	c.wg.Add(1)
	go func() {
		defer c.wg.Done()

		c.checkAllBackends()

		ticker := time.NewTicker(c.CheckInterval)

		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				c.checkAllBackends()
			case <-c.stopCh:
				return
			}
		}
	}()

}

func (c *Checker) Stop() {
	logger.Logger.Info().Msg("остановка health checker")
	// Закрываем канал остановки и ждём завершения горутины
	close(c.stopCh)
	c.wg.Wait()
}

func (c *Checker) checkAllBackends() {
	// Получаем список всех бэкендов из балансировщика
	// и запускаем горутины для проверки каждого бэкенда
	// с использованием WaitGroup для ожидания завершения всех горутин
	logger.Logger.Info().Msg("проверяем состояние бэкендов")
	backends := c.balancer.AllBackend()

	var wg sync.WaitGroup
	for _, backend := range backends {
		wg.Add(1)
		go func(b *balancer.Backend) {
			defer wg.Done()
			currentState := b.State
			newState := c.checkBackend(b)
			// Если состояние бэкенда изменилось, обновляем его в балансировщике
			if newState != currentState {
				logger.Logger.Info().Msgf("статус бэкенда %s изменился с %t на %t", b.Url, currentState, newState)
				if newState {
					c.balancer.MarkBackendUp(b.Url)
				} else {
					c.balancer.MarkBackendDown(b.Url)
				}
			} else {
				logger.Logger.Info().Msgf("статус бэкенда %s остался %t", b.Url, currentState)
			}
		}(backend)
	}
	wg.Wait()
	logger.Logger.Info().Msg("все бэкенды проверены")
}

func (c *Checker) checkBackend(backend *balancer.Backend) bool {
	// Проверяем, что бэкенд активен
	backendURL, err := url.Parse(backend.Url)
	if err != nil {
		logger.Logger.Error().Msgf("ошибка парсинга URL бэкенда %s: %v", backend.Url, err)
		return false
	}

	// Формируем URL для проверки
	healthURL := *backendURL
	healthURL.Path = "/health" // Можно настроить путь для проверки

	// Отправляем GET запрос для проверки доступности
	resp, err := c.httpClient.Get(healthURL.String())

	if err != nil {
		// Обработка ошибки
		// Проверяем, что ошибка имеет тип *url.Error
		if urlErr, ok := err.(*url.Error); ok {
			// Проверяем, является ли ошибка таймаутом
			if urlErr.Timeout() {
				// Обработка ошибки таймаута
				logger.Logger.Error().Msgf("таймаут при проверке бэкенда %s: %v", backend.Url, err)
			} else {
				// Другая ошибка
				logger.Logger.Error().Msgf("ошибка при проверке бэкенда %s: %v", backend.Url, err)
			}
		} else {
			// Ошибка другого типа
			logger.Logger.Error().Msgf("ошибка при проверке бэкенда %s: %v", backend.Url, err)
		}
		return false
	}

	defer resp.Body.Close()

	// Проверяем статус код ответа
	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		logger.Logger.Info().Msgf("бэкенд %s доступен", backend.Url)
		return true
	} else {
		logger.Logger.Error().Msgf("бэкенд %s недоступен, статус код: %d", backend.Url, resp.StatusCode)
		return false
	}
}

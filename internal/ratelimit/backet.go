package ratelimit

import (
	"sync"
	"time"
)

type TokenBucket struct {
	capacity   int           // Максимальное количество токенов
	rate       time.Duration // Интервал пополнения (напр. 1 токен/секунду)
	tokens     int           // Текущее количество токенов
	lastRefill time.Time     // Время последнего пополнения
	mu         sync.Mutex    // Для потокобезопасности
}

func NewBucket(capacity int, rate time.Duration) *TokenBucket {
	return &TokenBucket{
		capacity:   capacity,
		rate:       rate,
		tokens:     capacity,
		lastRefill: time.Now(),
	}
}

func (tb *TokenBucket) Allow() bool {
	tb.mu.Lock()
	defer tb.mu.Unlock()

	timeNow := time.Now()
	elapsed := timeNow.Sub(tb.lastRefill)

	tokensToAdd := int(elapsed / tb.rate)
	if tokensToAdd > 0 {
		tb.tokens = min(tb.tokens+tokensToAdd, tb.capacity)
		tb.lastRefill = timeNow
	}

	if tb.tokens > 0 {
		tb.tokens--
		return true
	}
	return false
}

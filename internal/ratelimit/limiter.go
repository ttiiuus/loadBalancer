package ratelimit

import (
	"sync"
	"time"
)

type RateLimiter struct {
	buckets     sync.Map      // key: clientID (string), value: *TokenBucket
	defaultCap  int           // Дефолтная емкость
	defaultRate time.Duration // Дефолтная скорость пополнения
}

func NewRateLimiter(defaultCap int, defaultRate time.Duration) *RateLimiter {
	return &RateLimiter{
		defaultCap:  defaultCap,
		defaultRate: defaultRate,
	}
}

func (rl *RateLimiter) Allow(clientID string) bool {
	bucket, _ := rl.buckets.LoadOrStore(clientID,
		NewBucket(rl.defaultCap, rl.defaultRate))

	return bucket.(*TokenBucket).Allow()
}

// Для кастомных лимитов (если у клиента особые правила)
func (rl *RateLimiter) SetCustomLimit(clientID string, cap int, rate time.Duration) {
	rl.buckets.Store(clientID, NewBucket(cap, rate))
}

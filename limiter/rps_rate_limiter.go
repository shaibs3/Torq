package limiter

import (
	"go.uber.org/zap"
	"sync"
	"time"
)

type TimeProvider func() time.Time

type RpsRateLimiter struct {
	mu           sync.Mutex
	lastReset    time.Time
	count        int
	limit        int
	logger       *zap.Logger
	timeProvider TimeProvider
}

func NewRateLimiter(rps int, logger *zap.Logger) *RpsRateLimiter {
	return NewRateLimiterWithTimeProvider(rps, logger, time.Now)
}

func NewRateLimiterWithTimeProvider(rps int, logger *zap.Logger, tp TimeProvider) *RpsRateLimiter {
	return &RpsRateLimiter{
		limit:        rps,
		lastReset:    tp(),
		logger:       logger,
		timeProvider: tp,
	}
}

func (l *RpsRateLimiter) Allow() bool {
	l.mu.Lock()
	defer l.mu.Unlock()

	now := l.timeProvider()

	if now.Sub(l.lastReset) >= time.Second {
		l.logger.Debug("Resetting rate limiter", zap.Time("time", now))
		l.lastReset = now
		l.count = 0
	}

	if l.count >= l.limit {
		l.logger.Warn("Rate limit exceeded", zap.Int("limit", l.limit))
		return false
	}

	l.count++
	l.logger.Debug("Request allowed", zap.Int("count", l.count))
	return true
}

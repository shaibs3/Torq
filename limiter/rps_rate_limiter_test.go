package limiter

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap/zaptest"
)

type fakeTime struct {
	current time.Time
}

func (f *fakeTime) Now() time.Time {
	return f.current
}

func (f *fakeTime) Advance(d time.Duration) {
	f.current = f.current.Add(d)
}

func TestRpsRateLimiter_Allow(t *testing.T) {
	ft := &fakeTime{current: time.Unix(0, 0)}
	logger := zaptest.NewLogger(t)
	limiter := NewRateLimiterWithTimeProvider(2, logger, ft.Now)

	require.True(t, limiter.Allow(), "expected first request to be allowed")
	assert.True(t, limiter.Allow(), "expected second request to be allowed")
	assert.False(t, limiter.Allow(), "expected third request to be denied")
}

func TestRpsRateLimiter_Reset(t *testing.T) {
	ft := &fakeTime{current: time.Unix(0, 0)}
	logger := zaptest.NewLogger(t)
	limiter := NewRateLimiterWithTimeProvider(1, logger, ft.Now)

	require.True(t, limiter.Allow(), "expected first request to be allowed")
	assert.False(t, limiter.Allow(), "expected second request to be denied")

	// Advance time by just over 1 second to trigger reset
	ft.Advance(time.Second + 10*time.Millisecond)
	assert.True(t, limiter.Allow(), "expected request after reset to be allowed")
}

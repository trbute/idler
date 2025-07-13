package ratelimit

import (
	"context"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/redis/go-redis/v9"
)

type RateLimiter struct {
	redis *redis.Client
}

func NewRateLimiter(redis *redis.Client) *RateLimiter {
	return &RateLimiter{redis: redis}
}

type RateLimit struct {
	Requests int
	Window   time.Duration
}

func (rl *RateLimiter) Allow(ctx context.Context, key string, limit RateLimit) (bool, int, time.Duration, error) {
	now := time.Now()
	bucketKey := fmt.Sprintf("rate_limit:%s", key)
	
	script := `
		local bucket_key = KEYS[1]
		local capacity = tonumber(ARGV[1])
		local refill_rate = tonumber(ARGV[2])
		local now = tonumber(ARGV[3])
		local window_ms = tonumber(ARGV[4])
		
		local bucket = redis.call('HMGET', bucket_key, 'tokens', 'last_refill')
		local tokens = tonumber(bucket[1]) or capacity
		local last_refill = tonumber(bucket[2]) or now
		
		local elapsed = now - last_refill
		local tokens_to_add = math.floor(elapsed * refill_rate / window_ms)
		tokens = math.min(capacity, tokens + tokens_to_add)
		
		if tokens >= 1 then
			tokens = tokens - 1
			redis.call('HMSET', bucket_key, 'tokens', tokens, 'last_refill', now)
			redis.call('EXPIRE', bucket_key, math.ceil(window_ms / 1000))
			return {1, tokens}
		else
			redis.call('HMSET', bucket_key, 'tokens', tokens, 'last_refill', now)
			redis.call('EXPIRE', bucket_key, math.ceil(window_ms / 1000))
			return {0, tokens}
		end
	`
	
	refillRate := float64(limit.Requests) / float64(limit.Window.Milliseconds())
	result, err := rl.redis.Eval(ctx, script, []string{bucketKey}, 
		limit.Requests, refillRate, now.UnixMilli(), limit.Window.Milliseconds()).Result()
	
	if err != nil {
		return false, 0, 0, err
	}
	
	resultSlice := result.([]interface{})
	allowed := resultSlice[0].(int64) == 1
	remaining := int(resultSlice[1].(int64))
	
	var resetTime time.Duration
	if !allowed {
		resetTime = time.Duration(float64(time.Millisecond) / refillRate)
	}
	
	return allowed, remaining, resetTime, nil
}

func (rl *RateLimiter) Middleware(limit RateLimit) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			key := r.RemoteAddr
			if forwarded := r.Header.Get("X-Forwarded-For"); forwarded != "" {
				key = forwarded
			}
			
			allowed, remaining, resetTime, err := rl.Allow(r.Context(), key, limit)
			if err != nil {
				http.Error(w, "Rate limiting error", http.StatusInternalServerError)
				return
			}
			
			w.Header().Set("X-RateLimit-Limit", strconv.Itoa(limit.Requests))
			w.Header().Set("X-RateLimit-Remaining", strconv.Itoa(remaining))
			
			if !allowed {
				w.Header().Set("X-RateLimit-Reset", strconv.FormatInt(time.Now().Add(resetTime).Unix(), 10))
				w.Header().Set("Retry-After", strconv.FormatInt(int64(resetTime.Seconds()), 10))
				http.Error(w, "Rate limit exceeded", http.StatusTooManyRequests)
				return
			}
			
			next.ServeHTTP(w, r)
		})
	}
}
package ratelimit

import (
	"context"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
	"github.com/trbute/idler/server/internal/auth"
)

const (
	defaultLimit     = 100
	defaultWindow    = time.Minute
	websocketLimit   = 50
	websocketWindow  = time.Minute
	unauthLimit      = 30
	unauthWindow     = time.Minute
)

type Limiter struct {
	redis *redis.Client
}

func NewLimiter(redis *redis.Client) *Limiter {
	return &Limiter{redis: redis}
}

func (l *Limiter) Allow(ctx context.Context, userID uuid.UUID, limit int, window time.Duration) (bool, error) {
	key := fmt.Sprintf("rate_limit:%s", userID.String())
	
	pipe := l.redis.Pipeline()
	incr := pipe.Incr(ctx, key)
	pipe.Expire(ctx, key, window)
	
	_, err := pipe.Exec(ctx)
	if err != nil {
		return false, fmt.Errorf("failed to execute rate limit check: %w", err)
	}
	
	count := incr.Val()
	return count <= int64(limit), nil
}

func (l *Limiter) Middleware(jwtSecret string, limit int, window time.Duration) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			token := r.Header.Get("Authorization")
			if token == "" {
				allowed, err := l.AllowUnauth(r.Context(), unauthLimit, unauthWindow)
				if err != nil {
					http.Error(w, "Rate limit check failed", http.StatusInternalServerError)
					return
				}
				
				if !allowed {
					w.Header().Set("X-RateLimit-Limit", strconv.Itoa(unauthLimit))
					w.Header().Set("X-RateLimit-Window", unauthWindow.String())
					http.Error(w, "Rate limit exceeded", http.StatusTooManyRequests)
					return
				}
				
				next.ServeHTTP(w, r)
				return
			}
			
			bearerToken, err := auth.GetBearerToken(r.Header)
			if err != nil {
				next.ServeHTTP(w, r)
				return
			}
			
			userID, err := auth.ValidateJWTWithBlacklist(r.Context(), bearerToken, jwtSecret, l.redis)
			if err != nil {
				next.ServeHTTP(w, r)
				return
			}
			
			allowed, err := l.Allow(r.Context(), userID, limit, window)
			if err != nil {
				http.Error(w, "Rate limit check failed", http.StatusInternalServerError)
				return
			}
			
			if !allowed {
				w.Header().Set("X-RateLimit-Limit", strconv.Itoa(limit))
				w.Header().Set("X-RateLimit-Window", window.String())
				http.Error(w, "Rate limit exceeded", http.StatusTooManyRequests)
				return
			}
			
			next.ServeHTTP(w, r)
		})
	}
}

func (l *Limiter) CheckWebSocketRateLimit(ctx context.Context, userID uuid.UUID) (bool, error) {
	return l.Allow(ctx, userID, websocketLimit, websocketWindow)
}

func (l *Limiter) AllowUnauth(ctx context.Context, limit int, window time.Duration) (bool, error) {
	key := "rate_limit:unauth"
	
	pipe := l.redis.Pipeline()
	incr := pipe.Incr(ctx, key)
	pipe.Expire(ctx, key, window)
	
	_, err := pipe.Exec(ctx)
	if err != nil {
		return false, fmt.Errorf("failed to execute unauthenticated rate limit check: %w", err)
	}
	
	count := incr.Val()
	return count <= int64(limit), nil
}
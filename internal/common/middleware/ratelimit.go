package middleware

import (
	"fmt"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/henryzhuhr/iam-superpowers/internal/common/errors"
	"github.com/redis/go-redis/v9"
)

func RateLimiter(rdb *redis.Client, limit int, window time.Duration) gin.HandlerFunc {
	return func(c *gin.Context) {
		key := fmt.Sprintf("ratelimit:%s", c.ClientIP())
		now := time.Now().UnixNano()
		windowStart := now - window.Nanoseconds()

		pipe := rdb.Pipeline()
		pipe.ZRemRangeByScore(c.Request.Context(), key, "-inf", fmt.Sprintf("%d", windowStart))
		pipe.ZAdd(c.Request.Context(), key, redis.Z{Score: float64(now), Member: fmt.Sprintf("%d", now)})
		pipe.ZCard(c.Request.Context(), key)
		pipe.Expire(c.Request.Context(), key, window)

		results, err := pipe.Exec(c.Request.Context())
		if err != nil {
			c.Next()
			return
		}

		count, _ := results[2].(*redis.IntCmd).Result()
		if int(count) > limit {
			errors.RespondError(c, &errors.APIError{
				Code:    errors.ErrTooManyRequests,
				Message: "too many requests",
			})
			c.Abort()
			return
		}

		c.Next()
	}
}

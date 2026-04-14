package middleware

import (
	"log"

	"github.com/gin-gonic/gin"
	"github.com/henryzhuhr/iam-superpowers/internal/common/errors"
)

func Recovery() gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if err := recover(); err != nil {
				log.Printf("panic: %v", err)
				errors.RespondError(c, errors.NewInternalError("internal server error"))
				c.Abort()
			}
		}()
		c.Next()
	}
}

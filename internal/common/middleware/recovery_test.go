package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestRecovery_CatchesPanic(t *testing.T) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	router := gin.New()
	router.Use(Recovery())
	router.GET("/test", func(c *gin.Context) {
		panic("test panic")
	})

	req, _ := http.NewRequest("GET", "/test", nil)
	assert.NotPanics(t, func() {
		router.ServeHTTP(w, req)
	})
	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

func TestRecovery_NoPanic(t *testing.T) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	router := gin.New()
	router.Use(Recovery())
	router.GET("/test", func(c *gin.Context) {
		c.String(http.StatusOK, "ok")
	})

	req, _ := http.NewRequest("GET", "/test", nil)
	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)
}

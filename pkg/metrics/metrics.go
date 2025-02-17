package metrics

import (
	"fmt"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var (
	RequestCounter = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "staking_api_http_requests_total",
			Help: "Total number of HTTP requests processed by staking-api",
		},
		[]string{"method", "path", "status"},
	)

	RequestDuration = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name: "staking_api_http_request_duration_seconds",
			Help: "Histogram of the response duration for HTTP requests processed by staking-api",
		},
		[]string{"method", "path", "status"},
	)
)

func init() {
	prometheus.MustRegister(RequestCounter)
	prometheus.MustRegister(RequestDuration)
}

func Middleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		defer func() {
			duration := time.Since(start).Seconds()
			statusCode := c.Writer.Status()
			path := c.FullPath()
			if path == "" {
				path = "unknown"
			}
			RequestCounter.WithLabelValues(c.Request.Method, path, fmt.Sprintf("%d", statusCode)).Inc()
			RequestDuration.WithLabelValues(c.Request.Method, path, fmt.Sprintf("%d", statusCode)).Observe(duration)
		}()
		c.Next()
	}
}

func Handler() gin.HandlerFunc {
	h := promhttp.Handler()
	return func(c *gin.Context) {
		h.ServeHTTP(c.Writer, c.Request)
	}
}

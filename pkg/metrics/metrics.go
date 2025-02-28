package metrics

import (
	"fmt"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var (
	DBErrorCounter = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "staking_api_db_errors_total",
			Help: "Total number of DB ops errors encountered",
		},
		[]string{},
	)

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

	RPCRequestErrorCounter = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "staking_api_story_api_req_errors_total",
			Help: "Total number of RPC request errors",
		},
		[]string{"story_endpoint"},
	)
)

func init() {
	prometheus.MustRegister(DBErrorCounter)
	prometheus.MustRegister(RequestCounter)
	prometheus.MustRegister(RequestDuration)
	prometheus.MustRegister(RPCRequestErrorCounter)
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

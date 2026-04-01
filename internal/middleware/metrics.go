package middleware

import (
	"fmt"
	"strconv"
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/middleware/adaptor"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var (
	httpRequestsTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "http_requests_total",
			Help: "Total number of HTTP requests processed.",
		},
		[]string{"method", "path", "status"},
	)
	httpRequestDurationSeconds = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "http_request_duration_seconds",
			Help:    "Duration of HTTP requests in seconds.",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"method", "path"},
	)
	httpRequestsInFlight = prometheus.NewGauge(
		prometheus.GaugeOpts{
			Name: "http_requests_in_flight",
			Help: "Current number of in-flight HTTP requests.",
		},
	)
)

func init() {
	prometheus.MustRegister(httpRequestsTotal, httpRequestDurationSeconds, httpRequestsInFlight)
}

// NewMetrics collects request count, duration, and in-flight metrics.
func NewMetrics() fiber.Handler {
	return func(ctx fiber.Ctx) error {
		startTime := time.Now()
		httpRequestsInFlight.Inc()
		defer httpRequestsInFlight.Dec()

		nextError := ctx.Next()

		path := string(ctx.Request().URI().Path())
		method := string(ctx.Method())
		status := strconv.Itoa(ctx.Response().StatusCode())
		durationSeconds := time.Since(startTime).Seconds()

		httpRequestsTotal.WithLabelValues(method, path, status).Inc()
		httpRequestDurationSeconds.WithLabelValues(method, path).Observe(durationSeconds)

		if nextError != nil {
			return fmt.Errorf("metricsMiddleware: %w", nextError)
		}
		return nil
	}
}

// MetricsHandler exposes Prometheus metrics for scraping.
func MetricsHandler() fiber.Handler {
	return adaptor.HTTPHandler(promhttp.Handler())
}

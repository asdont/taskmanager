package app

import (
	"net/http"
	"strconv"

	"github.com/prometheus/client_golang/prometheus"
)

type Metrics struct {
	MetricsRoute  string
	RequestsTotal *prometheus.CounterVec
}

func CreatePrometheusMetrics(metricsRoute string) Metrics {
	metrics := Metrics{
		MetricsRoute: metricsRoute,
		RequestsTotal: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Namespace: "api_tm",
				Subsystem: "request",
				Name:      "http_requests_total",
				Help:      "Total number of requests, their methods and paths",
			},
			[]string{"code", "method", "path"},
		),
	}

	prometheus.MustRegister(
		metrics.RequestsTotal,
	)

	metrics.RequestsTotal.WithLabelValues(strconv.Itoa(http.StatusOK), http.MethodGet, "/api/v1/tasks").Add(0)

	return metrics
}

package xgin

import (
	"browsertools/pkg/errors"
	"strconv"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

const serverNamespace = "http_server"

var (
	metricServerRequestDurations = promauto.NewHistogramVec(prometheus.HistogramOpts{
		Namespace:   serverNamespace,
		Subsystem:   "requests",
		Name:        "duration_ms",
		Help:        "backend server requests duration(ms).",
		ConstLabels: map[string]string{},
		Buckets:     []float64{5, 10, 25, 50, 100, 250, 500, 1000, 5000, 10000, 30000, 60000},
	}, []string{"path"})

	metricServerRequestCodeTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Namespace:   serverNamespace,
		Subsystem:   "requests",
		Name:        "code_total",
		Help:        "backend server requests error count.",
		ConstLabels: map[string]string{},
	}, []string{"path", "code"})

	metricServerRequestTimeoutTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Namespace:   serverNamespace,
		Subsystem:   "requests",
		Name:        "timeout_total",
		Help:        "backend server requests timeout count.",
		ConstLabels: map[string]string{},
	}, []string{"path"})
)

func TriggerErrorCode(value string, err errors.CodeError) {
	metricServerRequestCodeTotal.WithLabelValues(
		value, strconv.Itoa(err.Code())).Inc()
}

package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
)

var MetricsTotalRequestDuration = prometheus.NewHistogramVec(
	prometheus.HistogramOpts{
		Namespace: "encrypting_rpc_server",
		Subsystem: "request",
		Name:      "total_request_duration",
		Help:      "Histogram of the time it takes for all requests.",
		Buckets:   prometheus.DefBuckets,
	},
	[]string{"request"},
)

var MetricsEncryptionDuration = prometheus.NewHistogramVec(
	prometheus.HistogramOpts{
		Namespace: "encrypting_rpc_server",
		Subsystem: "request",
		Name:      "encryption_duration",
		Help:      "Histogram of the time it takes for encrypting a tx",
		Buckets:   prometheus.DefBuckets,
	},
	[]string{"request"},
)

var MetricsRequestedGasLimit = prometheus.NewHistogramVec(
	prometheus.HistogramOpts{
		Namespace: "encrypting_rpc_server",
		Subsystem: "request",
		Name:      "requested_gas_limit",
		Help:      "Histogram of the gas limit requested in tx",
		Buckets:   prometheus.DefBuckets,
	},
	[]string{"request"},
)

func InitMetrics() {
	prometheus.MustRegister(MetricsTotalRequestDuration)
	prometheus.MustRegister(MetricsEncryptionDuration)
	prometheus.MustRegister(MetricsRequestedGasLimit)
}

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
	[]string{"encrypted_tx_hash", "tx_hash"},
)

var MetricsEncryptionDuration = prometheus.NewHistogramVec(
	prometheus.HistogramOpts{
		Namespace: "encrypting_rpc_server",
		Subsystem: "request",
		Name:      "encryption_duration",
		Help:      "Histogram of the time it takes for encrypting a tx",
		Buckets:   prometheus.DefBuckets,
	},
	[]string{"encrypted_tx_hash"},
)

var gasLimitBuckets = []float64{21000, 25000, 35000, 50000, 70000, 100000, 200000, 500000, 1000000, 10000000, 30000000}

var MetricsRequestedGasLimit = prometheus.NewHistogramVec(
	prometheus.HistogramOpts{
		Namespace: "encrypting_rpc_server",
		Subsystem: "request",
		Name:      "requested_gas_limit",
		Help:      "Histogram of the gas limit requested in tx",
		Buckets:   gasLimitBuckets,
	},
	[]string{"encrypted_tx_hash"},
)

func InitMetrics() {
	prometheus.MustRegister(MetricsTotalRequestDuration)
	prometheus.MustRegister(MetricsEncryptionDuration)
	prometheus.MustRegister(MetricsRequestedGasLimit)
}

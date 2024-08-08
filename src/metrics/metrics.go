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

var MetricsUpstreamRequestDuration = prometheus.NewHistogramVec(
	prometheus.HistogramOpts{
		Namespace: "encrypting_rpc_server",
		Subsystem: "upstream_request",
		Name:      "duration",
		Help:      "Histogram of the request duration for upstream request",
		Buckets:   prometheus.DefBuckets,
	},
	[]string{"method"},
)

var MetricsCancellationTxCounter = prometheus.NewCounterVec(
	prometheus.CounterOpts{
		Namespace: "encrypting_rpc_server",
		Subsystem: "request",
		Name:      "cancellation_tx_counter",
		Help:      "Counter of tx which were cancelled",
	},
	[]string{"tx_hash"},
)

var MetricsErrorReturnedCounter = prometheus.NewCounter(
	prometheus.CounterOpts{
		Namespace: "encrypting_rpc_server",
		Subsystem: "request",
		Name:      "error_returned",
		Help:      "Counter of error returned",
	},
)

var MetricsERPCBalance = prometheus.NewGauge(
	prometheus.GaugeOpts{
		Namespace: "encrypting_rpc_server",
		Subsystem: "balance",
		Name:      "erpc_address_balance_xdai",
		Help:      "Native token balance",
	},
)

func InitMetrics() {
	prometheus.MustRegister(MetricsTotalRequestDuration)
	prometheus.MustRegister(MetricsEncryptionDuration)
	prometheus.MustRegister(MetricsRequestedGasLimit)
	prometheus.MustRegister(MetricsUpstreamRequestDuration)
	prometheus.MustRegister(MetricsCancellationTxCounter)
	prometheus.MustRegister(MetricsErrorReturnedCounter)
	prometheus.MustRegister(MetricsERPCBalance)
}

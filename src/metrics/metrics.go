package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
)

var TotalRequestDuration = prometheus.NewHistogramVec(
	prometheus.HistogramOpts{
		Namespace: "encrypting_rpc_server",
		Subsystem: "request",
		Name:      "total_request_duration",
		Help:      "Histogram of the time it takes for all requests.",
		Buckets:   prometheus.DefBuckets,
	},
	[]string{"encrypted_tx_hash", "tx_hash"},
)

var EncryptionDuration = prometheus.NewHistogramVec(
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

var RequestedGasLimit = prometheus.NewHistogramVec(
	prometheus.HistogramOpts{
		Namespace: "encrypting_rpc_server",
		Subsystem: "request",
		Name:      "requested_gas_limit",
		Help:      "Histogram of the gas limit requested in tx",
		Buckets:   gasLimitBuckets,
	},
	[]string{"encrypted_tx_hash"},
)

var UpstreamRequestDuration = prometheus.NewHistogramVec(
	prometheus.HistogramOpts{
		Namespace: "encrypting_rpc_server",
		Subsystem: "upstream_request",
		Name:      "duration",
		Help:      "Histogram of the request duration for upstream request",
		Buckets:   prometheus.DefBuckets,
	},
	[]string{"method"},
)

var CancellationTxGauge = prometheus.NewGaugeVec(
	prometheus.GaugeOpts{
		Namespace: "encrypting_rpc_server",
		Subsystem: "request",
		Name:      "cancellation_txs_total",
		Help:      "Counter of tx which were cancelled",
	},
	[]string{"tx_hash"},
)

var ErrorReturnedCounter = prometheus.NewCounter(
	prometheus.CounterOpts{
		Namespace: "encrypting_rpc_server",
		Subsystem: "request",
		Name:      "errors_returned_total",
		Help:      "Counter of error returned",
	},
)

var ERPCBalance = prometheus.NewGauge(
	prometheus.GaugeOpts{
		Namespace: "encrypting_rpc_server",
		Subsystem: "balance",
		Name:      "erpc_address_balance_xdai",
		Help:      "Native token balance",
	},
)

func InitMetrics() {
	prometheus.MustRegister(TotalRequestDuration)
	prometheus.MustRegister(EncryptionDuration)
	prometheus.MustRegister(RequestedGasLimit)
	prometheus.MustRegister(UpstreamRequestDuration)
	prometheus.MustRegister(CancellationTxGauge)
	prometheus.MustRegister(ErrorReturnedCounter)
	prometheus.MustRegister(ERPCBalance)
}

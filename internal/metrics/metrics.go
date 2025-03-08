package metrics

import (
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// Collector holds all the Prometheus metrics
type Collector struct {
	// Current metrics
	TotalDeposits         prometheus.Counter
	BlockedDeposits       *prometheus.CounterVec
	DepositsByAccount     *prometheus.CounterVec
	DepositValueHistogram prometheus.Histogram
}

// NewCollector creates a new metrics collector with initialized metrics
func NewCollector() *Collector {
	return &Collector{
		TotalDeposits: promauto.NewCounter(
			prometheus.CounterOpts{
				Name: "opstack_total_deposits",
				Help: "Total number of deposits through OptimismPortal",
			}),

		BlockedDeposits: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Name: "opstack_blocked_deposits",
				Help: "Total number of blocked deposits from frozen accounts",
			},
			[]string{"account"}),

		DepositsByAccount: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Name: "opstack_deposits_by_account",
				Help: "Number of deposits grouped by sender account",
			},
			[]string{"account"}),

		DepositValueHistogram: promauto.NewHistogram(
			prometheus.HistogramOpts{
				Name:    "opstack_deposit_value",
				Help:    "Distribution of deposit values in ETH",
				Buckets: prometheus.ExponentialBuckets(0.001, 10, 7), // 0.001 ETH to 1000 ETH
			}),
	}
}

// StartServer starts the /metrics endpoint for Prometheus
func StartServer(metricsPort string) {
	if metricsPort == "" {
		log.Println("[WARN] METRICS_PORT environment variable not set, using default port 9100")
		metricsPort = "9100"
	}

	metricsAddr := fmt.Sprintf(":%s", metricsPort)

	// Create a separate mux for metrics to avoid conflicts with the main RPC handler
	mux := http.NewServeMux()
	mux.Handle("/metrics", promhttp.Handler())

	log.Printf("[INFO] Starting Prometheus metrics server on %s\n", metricsAddr)

	// Start the server in a goroutine
	go func() {
		err := http.ListenAndServe(metricsAddr, mux)
		if err != nil {
			log.Fatalf("[ERROR] Could not start Prometheus metrics server: %v", err)
		}
	}()

	// Add a quick check to verify the server is running
	go func() {
		time.Sleep(2 * time.Second) // Give the server time to start
		_, err := http.Get(fmt.Sprintf("http://localhost:%s/metrics", metricsPort))
		if err != nil {
			log.Printf("[WARN] Metrics server may not be running correctly: %v", err)
		} else {
			log.Printf("[INFO] Metrics server verified running on port %s", metricsPort)
		}
	}()
}

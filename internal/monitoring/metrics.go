// File: internal/monitoring/metrics.go
package monitoring

import (
	"log"
	"net/http"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var (
	// Gauges: Values that can go up or down.
	IsLeader = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "kvstore_raft_is_leader",
		Help: "Is this node the current Raft leader (1 if leader, 0 otherwise).",
	})
	CurrentTerm = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "kvstore_raft_current_term",
		Help: "The current Raft term number.",
	})
	CommitIndex = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "kvstore_raft_commit_index",
		Help: "The current Raft commit index.",
	})
	LastApplied = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "kvstore_raft_last_applied",
		Help: "The current Raft last applied index.",
	})

	// Counters: Values that can only go up.
	RequestsTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "kvstore_requests_total",
		Help: "Total number of KV store requests.",
	}, []string{"op"}) // "op" label for "put", "get", "delete"
)

// StartMetricsServer starts a separate HTTP server to expose the /metrics endpoint.
func StartMetricsServer(addr string) {
	go func() {
		log.Printf("Metrics server listening on %s", addr)
		http.Handle("/metrics", promhttp.Handler())
		if err := http.ListenAndServe(addr, nil); err != nil {
			log.Fatalf("Failed to start metrics server: %v", err)
		}
	}()
}
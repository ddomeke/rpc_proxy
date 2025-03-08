package proxy

import (
	"fmt"
	"log"
	"net/http"

	"github.com/ddomeke/rpc_proxy/internal/config"
	"github.com/ddomeke/rpc_proxy/internal/eth"
	"github.com/ddomeke/rpc_proxy/internal/metrics"
)

// Server holds the RPC proxy server configuration
type Server struct {
	config           *config.Config
	ethClients       *eth.Clients
	metricsCollector *metrics.Collector
}

// NewServer creates a new RPC proxy server
func NewServer(cfg *config.Config, clients *eth.Clients, collector *metrics.Collector) *Server {
	return &Server{
		config:           cfg,
		ethClients:       clients,
		metricsCollector: collector,
	}
}

// Start starts the RPC proxy server
func (s *Server) Start() error {
	http.HandleFunc("/", s.proxyHandler)
	proxyAddress := fmt.Sprintf(":%s", s.config.ProxyPort)
	log.Printf("[INFO] RPC Proxy started. Port: %s\n", proxyAddress)
	return http.ListenAndServe(proxyAddress, nil)
}

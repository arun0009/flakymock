package main

import (
	"log"
	"net/http"

	"github.com/arun0009/flakymock/pkg/config"
	"github.com/arun0009/flakymock/pkg/observability"
	"github.com/arun0009/flakymock/pkg/server"

	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func main() {
	log.SetFlags(0)
	log.Println("Initializing FlakyMock Server...")

	// 1. Load Configuration
	cfg, err := config.LoadConfig("scenarios.yaml")
	if err != nil {
		log.Fatalf("Fatal: Failed to load config: %v", err)
	}

	// 2. Initialize Observability
	observability.InitMetrics()

	// 3. Register Prometheus Handler
	http.Handle("/metrics", promhttp.Handler())

	// 4. Create and Start Server
	router := server.NewRouter(cfg)

	http.Handle("/", router)

	log.Printf("Server starting on port %s (TLS: %t, CORS: %t)", cfg.Port, cfg.EnableTLS, cfg.EnableCORS)

	if cfg.EnableTLS {
		// Use server.RunTLS or os.Exit(1) if certs are missing
		log.Fatalf("Server failed to run with TLS: %v", server.RunTLS(cfg))
	} else {
		log.Fatalf("Server failed to run: %v", server.Run(cfg))
	}
}

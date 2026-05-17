package main

import (
	"log/slog"
	"net"
	"net/http"
	"os"

	"github.com/monjiapawne/iris_exporter/internal/client"
	"github.com/monjiapawne/iris_exporter/internal/endpoints"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func main() {
	// General config setup
	exporterCfg := cfgLoadExporter() // Local exporter configuration
	logger := slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: exporterCfg.logLevel}))

	// Target setup
	targetCfg := cfgLoadTarget() // Scrape target configuration
	c := client.NewClient(
		targetCfg.APIKey,
		targetCfg.Scheme,
		targetCfg.Host,
		targetCfg.Port,
		targetCfg.VerifyTLS,
		logger,
	)

	// Registry
	reg := prometheus.NewRegistry()
	optsCfg := cfgLoadCollector() // Controls exporter behavior
	reg.MustRegister(
		endpoints.NewUsersMetric(c),
		endpoints.NewCasesMetric(c, optsCfg.Cases),
		endpoints.NewTasksMetric(c, optsCfg.Tasks),
		endpoints.NewAlertsMetric(c, optsCfg.Alerts),
	)

	// HTTP server
	mux := http.NewServeMux()
	// Paths
	mux.Handle("/metrics", promhttp.HandlerFor(reg, promhttp.HandlerOpts{}))
	mux.Handle("/", http.RedirectHandler("/metrics", http.StatusMovedPermanently))
	mux.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) {}) // Kubernetes

	// Serve
	addr := net.JoinHostPort(exporterCfg.host, exporterCfg.port)
	if err := serve(mux, addr, logger); err != nil {
		logger.Error("server error", "err", err)
		os.Exit(1)
	}
}

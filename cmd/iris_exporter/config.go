package main

import (
	"fmt"
	"log/slog"
	"os"
	"strconv"

	"github.com/monjiapawne/iris_exporter/internal/endpoints"
)

// Target
type TargetCfg struct {
	APIKey    string
	Scheme    string
	Host      string
	Port      string
	VerifyTLS bool
}

func cfgLoadTarget() TargetCfg {
	// Extra check for API key
	apiKey := envStr("IRIS_APIKEY", "")
	if apiKey == "" {
		fmt.Fprintln(os.Stderr, "IRIS_APIKEY is required")
		os.Exit(1)
	}

	return TargetCfg{
		APIKey:    apiKey,
		Scheme:    envStr("IRIS_SCHEME", "https"),
		Host:      envStr("IRIS_HOST", "localhost"),
		Port:      envStr("IRIS_PORT", "443"),
		VerifyTLS: envBool("IRIS_VERIFY_TLS", true),
	}
}

// Exporter
type exporterCfg struct {
	host     string
	port     string
	logLevel slog.Level
}

func cfgLoadExporter() exporterCfg {
	return exporterCfg{
		host:     envStr("EXPORTER_ADDRESS", ""),
		port:     envStr("EXPORTER_PORT", "10043"),
		logLevel: envLogLevel("EXPORTER_LOG_LEVEL", slog.LevelInfo),
	}
}

// Exporter behavior options
type collectorCfg struct {
	Cases  endpoints.CasesOptions
	Alerts endpoints.AlertOptions
	Tasks  endpoints.TaskOptions
}

func cfgLoadCollector() collectorCfg {
	return collectorCfg{
		Cases: endpoints.CasesOptions{
			SplitSubClasses: envBool("EXPORTER_CASES_SPLIT_SUBCLASSES", true),
		},
		Alerts: endpoints.AlertOptions{
			SplitSubClasses: envBool("EXPORTER_ALERTS_SPLIT_SUBCLASSES", true),
		},
		Tasks: endpoints.TaskOptions{
			TasksEnabled: envBool("EXPORTER_TASKS_ENABLED", true),
		},
	}
}

// Environment variable helpers
func envStr(key, backup string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback(key, backup)
}

func envBool(key string, backup bool) bool {
	if v := os.Getenv(key); v != "" {
		b, err := strconv.ParseBool(v)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error parsing boolean value for %s: %v\n", key, err)
			return fallback(key, backup)
		}
		return b
	}
	return fallback(key, backup)
}

func envLogLevel(key string, backup slog.Level) slog.Level {
	if v := os.Getenv(key); v != "" {
		var l slog.Level
		if err := l.UnmarshalText([]byte(v)); err != nil {
			fmt.Fprintf(os.Stderr, "invalid log level %q for %s, using default\n", v, key)
			return backup
		}
		return l
	}
	return fallback(key, backup)
}

// Log and return fallback
func fallback[T any](key string, fallback T) T {
	fmt.Fprintf(os.Stderr, "Config defaulting to: %s: %v\n", key, fallback)
	return fallback
}

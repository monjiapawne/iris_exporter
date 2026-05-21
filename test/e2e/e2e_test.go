package e2e

import (
	"io"
	"net/http"
	"strings"
	"testing"
)

func TestExporterResponds(t *testing.T) {
	resp, err := http.Get("http://localhost:10043/metrics")
	if err != nil {
		t.Skip("exporter not running:", err)
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	s := string(body)

	for _, metric := range []string{
		"iris_cases_current",
		"iris_alerts_total",
	} {
		if !strings.Contains(s, metric) {
			t.Errorf("missing metric: %s", metric)
		}
	}
}

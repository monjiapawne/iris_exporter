package collector

import (
	"testing"

	"github.com/prometheus/client_golang/prometheus"
	dto "github.com/prometheus/client_model/go"
)

func dummyFetch() (testType, error) {
	return testType{metricBool: true, metricString: "hello_world", metricFloat: 99}, nil
}

func TestCollect(t *testing.T) {
	testCases := []struct {
		name     string
		metrics  []MetricDef[testType]
		expected []float64
	}{
		{
			name: "response parsing metrics",
			metrics: []MetricDef[testType]{
				Gauge("test_bool", "metric to test bool reading", func(t testType) float64 {
					if t.metricBool {
						return 0 // expected in dummyFetch()
					}
					return 1
				}),
				Gauge("test_float", "metric to test float reading", func(t testType) float64 { return t.metricFloat }),
				Gauge("test_string", "metric to test string reading", func(t testType) float64 {
					if t.metricString == "hello_world" {
						return 0
					}
					return 1
				}),
			},
			expected: []float64{0, 99, 0},
		},
		{
			name: "two gauges",
			metrics: []MetricDef[testType]{
				Gauge("test_metric", "metric to test collector", func(t testType) float64 { return 1 }),
				Gauge("test_metric", "metric to test collector", func(t testType) float64 { return 2 }),
			},
			expected: []float64{1, 2},
		},
		{
			name: "single gauge",
			metrics: []MetricDef[testType]{
				Gauge("single_metric", "a single gauge metric", func(t testType) float64 { return 42 }),
			},
			expected: []float64{42},
		},
		{
			name:     "no metrics",
			metrics:  []MetricDef[testType]{},
			expected: []float64{},
		},
		{
			name: "multiple gauges",
			metrics: []MetricDef[testType]{
				Gauge("metric_a", "first metric", func(t testType) float64 { return 0 }),
				Gauge("metric_b", "second metric", func(t testType) float64 { return 5 }),
				Gauge("metric_c", "third metric", func(t testType) float64 { return 10 }),
			},
			expected: []float64{0, 5, 10},
		},
		{
			name: "multi-value gauge",
			metrics: []MetricDef[testType]{
				GaugeMulti("multi_metric", "multi-value gauge", []string{"label"}, func(t testType) []LabeledValue {
					return []LabeledValue{
						{Labels: []string{"a"}, Value: 1},
						{Labels: []string{"b"}, Value: 2},
						{Labels: []string{"c"}, Value: 3},
					}
				}),
			},
			expected: []float64{1, 2, 3},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ch := make(chan prometheus.Metric, 10)
			NewCollector(tc.metrics, dummyFetch).Collect(ch)
			close(ch)

			var got []float64
			for m := range ch {
				dtoM := &dto.Metric{}
				m.Write(dtoM)
				got = append(got, dtoM.Gauge.GetValue())
			}

			if len(got) != len(tc.expected) {
				t.Fatalf("expected %d metrics, got %d", len(tc.expected), len(got))
			}
			for i, v := range got {
				if v != tc.expected[i] {
					t.Errorf("metric[%d]: expected %v, got %v", i, tc.expected[i], v)
				}
			}
		})
	}
}

type testType struct {
	metricBool   bool
	metricString string
	metricFloat  float64
}

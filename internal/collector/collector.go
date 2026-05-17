package collector

import (
	"github.com/prometheus/client_golang/prometheus"
)

const namespace = "iris_"

type MetricDef[R any] struct {
	desc *prometheus.Desc

	// prometheus.ValueType options:
	//    prometheus.GaugeValue - values that can go up or down (users, temperature)
	//	  prometheus.CounterValue - values that only increase
	//    prometheus.UntypedValue - values where the type is not speified (rarely used, acts like Gauge)
	valueType prometheus.ValueType

	// set for single-value metrics (Gauge, GaugeLabeled)
	value  func(R) float64
	labels func(R) []string

	// set for multi-value metrics (GaugeMulti)
	multiValues func(R) []LabeledValue
}

// Helpers to create Gauges
func Gauge[R any](name, help string, fn func(R) float64) MetricDef[R] {
	return MetricDef[R]{desc: prometheus.NewDesc(namespace+name, help, nil, nil), valueType: prometheus.GaugeValue, value: fn}
}

func GaugeLabeled[R any](name, help string, labelNames []string, fn func(R) float64, labels func(R) []string) MetricDef[R] {
	return MetricDef[R]{desc: prometheus.NewDesc(namespace+name, help, labelNames, nil), valueType: prometheus.GaugeValue, value: fn, labels: labels}
}

// Used for multiple value gauges
// The function must return a slice of LabeledValues which are basically Key/Values for the metric
func GaugeMulti[R any](name, help string, labelNames []string, fn func(R) []LabeledValue) MetricDef[R] {
	return MetricDef[R]{desc: prometheus.NewDesc(namespace+name, help, labelNames, nil), valueType: prometheus.GaugeValue, multiValues: fn}
}

// LabeledValue is a single emitted data point from a GaugeMulti one label set and its value.
type LabeledValue struct {
	Labels []string
	Value  float64
}

type Collector[R any] struct {
	metrics []MetricDef[R]
	fetch   func() (R, error)
}

// NewCollector constructs a collector
//   - metrics accepts a slice of metricDefs of any type (Gauge, GaugeLabeled, GaugeMulti)
//   - fetch is a function that returns the unmarshaled API response
func NewCollector[R any](metrics []MetricDef[R], fetch func() (R, error)) *Collector[R] {
	return &Collector[R]{metrics: metrics, fetch: fetch}
}

func (c *Collector[R]) Describe(ch chan<- *prometheus.Desc) {
	for _, m := range c.metrics {
		ch <- m.desc
	}
}

func (c *Collector[R]) Collect(ch chan<- prometheus.Metric) {
	// Request the marshalled data, typically API call to IRIS
	// fetch() is provided by the endpoint, implemented inside the generic type
	resp, err := c.fetch()
	if err != nil {
		// Allow the exporter to continue but skip the failed metric
		return
	}

	// Loop over all metrics in the slice
	for _, m := range c.metrics {
		// Check if it's a multivalue guage first
		if m.multiValues != nil {
			// Loop over the slice of metrics
			for _, lv := range m.multiValues(resp) {
				ch <- prometheus.MustNewConstMetric(m.desc, m.valueType, lv.Value, lv.Labels...)
			}
			continue
		}

		// Single value guages
		var labelVals []string
		if m.labels != nil {
			labelVals = m.labels(resp)
		}
		ch <- prometheus.MustNewConstMetric(m.desc, m.valueType, m.value(resp), labelVals...)
	}
}

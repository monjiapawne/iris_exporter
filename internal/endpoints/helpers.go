package endpoints

import (
	"strings"
	"time"

	"github.com/monjiapawne/iris_exporter/internal/collector"
)

// Replace ' ' to '_' for idiomatic prometheus naming
func normalizeLabel(s string) string {
	return strings.ToLower(strings.ReplaceAll(s, " ", "_"))
}

// countBy loops over a slice of T, counting occurrences grouped by a single key per item.
// Use countByEach when one item can produce multiple keys (e.g. split subclasses).
// Note: labels are normalized to idiomatic Prometheus format as a side effect.
func countBy[T any](items []T, key func(T) string) []collector.LabeledValue {
	counts := map[string]float64{}
	for _, item := range items {
		counts[normalizeLabel(key(item))]++
	}
	out := make([]collector.LabeledValue, 0, len(counts))
	for k, n := range counts {
		out = append(out, collector.LabeledValue{Labels: []string{k}, Value: n})
	}
	return out
}

// countByEach is like countBy but the key function returns a slice, so one item can
// contribute to multiple buckets (e.g. splitting "availability:ddos" into both keys).
func countByEach[T any](items []T, keys func(T) []string) []collector.LabeledValue {
	counts := map[string]float64{}
	for _, item := range items {
		for _, k := range keys(item) {
			counts[normalizeLabel(k)]++
		}
	}
	out := make([]collector.LabeledValue, 0, len(counts))
	for k, n := range counts {
		out = append(out, collector.LabeledValue{Labels: []string{k}, Value: n})
	}
	return out
}

// Custom fields for unmarshing

// ex: "case_open_date": "05/10/2026"
type MMDDYYYYDate time.Time

func (d *MMDDYYYYDate) UnmarshalJSON(data []byte) error {
	s := strings.Trim(string(data), `"`)
	if s == "" || s == "null" {
		*d = MMDDYYYYDate(time.Time{})
		return nil
	}

	t, err := time.Parse("01/02/2006", s)
	if err != nil {
		return err
	}
	*d = MMDDYYYYDate(t)
	return nil
}

func (d MMDDYYYYDate) Time() time.Time {
	return time.Time(d)
}

// ISODateTime unmarshals timestamps like "2026-05-14T22:06:12.224370" (no timezone)
type ISODateTime time.Time

func (d *ISODateTime) UnmarshalJSON(data []byte) error {
	s := strings.Trim(string(data), `"`)
	if s == "" || s == "null" {
		*d = ISODateTime(time.Time{})
		return nil
	}
	t, err := time.Parse("2006-01-02T15:04:05.999999", s)
	if err != nil {
		return err
	}
	*d = ISODateTime(t)
	return nil
}

func (d ISODateTime) Time() time.Time {
	return time.Time(d)
}

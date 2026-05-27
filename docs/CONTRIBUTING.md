# Contributing

## Overview

The structure of this exporter is focused on grouping metrics to the logic that parses them using generics.

The overall design has one major trade off. 

- For each endpoint called the response is unmarshaled `1` time *typically* 
- For **each** metric that parses the struct, they **each** walk it `1` time 
    - So rather a single walk through struct, which is all that would techically be needed to extract the data, they all do their own walk

> Why make this trade off?

- Looser coupling of metrics
- Easier to add new metrics
- Clear what each metric is doing to extract its values

This trade off is subject to change as with scale perhaps it will waste too many cycles needlessly. 

The general time complexity of this trade off:

```bash
# N = number of metrics
# M = number of api calls
# Current design:
O(M * N)
# Single walk alternative:
O(M)
```

### Add a new endpoint

Basic structure

```go
// --------------------------------
// ./internal/endpoints/alerts.go 
// --------------------------------

// Base example of a new logical endpoint/metric type
func NewAlertsMetric(c *client.Client) *collector.Collector[alertsReponse] {
	alertsMetric := []collector.MetricDef[alertsReponse]{
        // New metric to parse from the decoded API response
        // This is a simple check of the length of the Data.Slice slice
     	collector.Gauge("alerts_total", "Alerts total",
			func(r alertsReponse) float64 {
				return float64(len(r.Data.Alerts))
		}),
    }

    // Provide the slice above, so it be looped through and the function that calls the API
    // and decodes into the response the metrics require (alertsResponse)
	return collector.NewCollector(alertsMetric, func() (alertsReponse, error) {
        // Replace with endpoint
		return client.APICall[alertsReponse](c, "/manage/cases/list")
	})
}

// Decode of API response
type alertsReponse struct {
    Data struct {
        Alerts []struct {
        } `json:"alerts"`
    }
}

// --------------------------------
// ./cmd/iris_exporter/main.go
// --------------------------------

// Update the registry to load the new metrics
func main(){
	// ... omitted for brevity
	reg.MustRegister(
		endpoints.NewAlertsMetric(c), // New collector endpoint here
	)
}
```

From there you can add any type of metric to the slice, in this case `alertsMetric`, there are helpers to create new metrics in `collectors.go`

## Building from source

```bash
go build ./cmd/iris_exporter
```

## Grafana Examples

Grafana example dashboards are welcome as well. Please include screenshots following the current pattern.

## Resources

- [docs](https://docs.dfir-iris.org/latest/operations/api/)
- [api docs](https://docs.dfir-iris.org/latest/_static/iris_api_reference_v2.0.4.html)

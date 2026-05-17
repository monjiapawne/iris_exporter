package endpoints

import (
	"fmt"
	"strings"

	"github.com/monjiapawne/iris_exporter/internal/client"
	"github.com/monjiapawne/iris_exporter/internal/collector"
)

type AlertOptions struct {
	// Separate subclasses
	// 	"availability:ddos" +1
	//  "availability" "ddos" +1 for each
	SplitSubClasses bool
}

func NewAlertsMetric(c *client.Client, alertOps AlertOptions) *collector.Collector[alertsReponse] {
	casesMetrics := []collector.MetricDef[alertsReponse]{
		collector.Gauge("alerts_total", "Alerts total", func(r alertsReponse) float64 { return float64(r.Data.Total) }),
		collector.GaugeMulti("alerts_severity", "Alerts by severtiy", []string{"severity"},
			func(r alertsReponse) []collector.LabeledValue {
				return countBy(r.Data.Alerts, func(a alertItem) string { return a.Severity.Name })
			}),
		collector.GaugeMulti("alerts_status", "Alerts by status", []string{"status"},
			func(r alertsReponse) []collector.LabeledValue {
				return countBy(r.Data.Alerts, func(a alertItem) string { return a.Status.Name })
			}),
		collector.GaugeMulti("alerts_source", "Alerts by source", []string{"source"},
			func(r alertsReponse) []collector.LabeledValue {
				return countBy(r.Data.Alerts, func(a alertItem) string { return a.AlertSource })
			}),
		collector.GaugeMulti("alerts_resolution_status", "Alerts by resolution status", []string{"resolution_status"},
			func(r alertsReponse) []collector.LabeledValue {
				return countBy(r.Data.Alerts,
					func(a alertItem) string {
						if a.ResolutionStatus.Name == "" {
							return "no_response"
						} else {
							return a.ResolutionStatus.Name
						}
					})
			}),
		collector.GaugeMulti("alerts_classification", "Alerts by classification", []string{"classification"},
			func(r alertsReponse) []collector.LabeledValue {
				return countByEach(r.Data.Alerts,
					func(a alertItem) []string {
						if alertOps.SplitSubClasses {
							return strings.Split(a.Classification.Name, ":")
						}
						return []string{a.Classification.Name}
					})
			}),
	}

	return collector.NewCollector(casesMetrics, func() (alertsReponse, error) {
		// Loop through all pages of alerts in 100 page sizes
		// Merge all alertResponses into a single alertResponse
		const perPage = 100
		var merged alertsReponse

		for page := 1; ; page++ {
			resp, err := client.APICall[alertsReponse](c, fmt.Sprintf("/alerts/filter?page=%d&per_page=%d", page, perPage))
			if err != nil {
				return merged, err
			}
			merged.Data.Total = resp.Data.Total
			merged.Data.Alerts = append(merged.Data.Alerts, resp.Data.Alerts...)
			if len(merged.Data.Alerts) >= resp.Data.Total {
				break
			}
		}
		return merged, nil
	})
}

type alertItem struct {
	Severity struct {
		Name string `json:"severity_name"`
	} `json:"severity"`
	Status struct {
		Name string `json:"status_name"`
	} `json:"status"`
	AlertSource    string `json:"alert_source"`
	Classification struct {
		Name string `json:"name"`
	} `json:"classification"`
	ResolutionStatus struct {
		Name string `json:"resolution_status_name"`
	} `json:"resolution_status"`
}

type alertsReponse struct {
	Data struct {
		Total  int         `json:"total"`
		Alerts []alertItem `json:"alerts"`
	} `json:"data"`
}

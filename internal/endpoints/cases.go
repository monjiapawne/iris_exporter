package endpoints

import (
	"strings"
	"time"

	"github.com/monjiapawne/iris_exporter/internal/client"
	"github.com/monjiapawne/iris_exporter/internal/collector"
)

type CasesOptions struct {
	// Separate subclasses
	// 	"availability:ddos" +1
	//  "availability" "ddos" +1 for each
	SplitSubClasses bool
}

func NewCasesMetric(c *client.Client, caseOpts CasesOptions) *collector.Collector[casesResponse] {
	casesMetrics := []collector.MetricDef[casesResponse]{
		collector.Gauge("cases_current", "Current number of cases", func(r casesResponse) float64 { return float64(len(r.Data)) }),
		collector.GaugeMulti("cases_state", "Cases per state.", []string{"state"},
			func(r casesResponse) []collector.LabeledValue {
				return countBy(r.Data,
					func(c caseItem) string {
						if c.State == "" {
							return "none"
						}
						return c.State
					})
			}),
		collector.GaugeMulti("cases_classification", "Cases by classification", []string{"classification"},
			func(r casesResponse) []collector.LabeledValue {
				return countByEach(r.Data,
					func(c caseItem) []string {
						if c.Classification == "" {
							return []string{"none"}
						}
						if caseOpts.SplitSubClasses {
							return strings.Split(c.Classification, ":")
						}
						return []string{c.Classification}
					})
			}),
		// Age of cases that were closed
		collector.Gauge("cases_average_close_duration_days", "Average time to close case",
			func(r casesResponse) float64 {
				var count float64
				var total float64
				for _, c := range r.Data {
					// If the case is NOT closed, skip
					if c.CaseCloseDate.Time().IsZero() {
						continue
					}
					count++
					// Calculate days it took to close
					daysToCloseCase := c.CaseCloseDate.Time().Sub(c.CaseOpenDate.Time()).Hours() / 24
					// fmt.Println(daysToCloseCase, "for", c.CaseID)
					total += daysToCloseCase
				}
				if count == 0 {
					return 0
				}
				return total / count
			}),
		// Total age of all cases, regardless if they're open or closed
		collector.Gauge("cases_average_open_age_days", "Average age of all cases since open date",
			func(r casesResponse) float64 {
				var count float64
				var total float64
				for _, c := range r.Data {
					count++
					caseAge := time.Since(c.CaseOpenDate.Time()).Hours() / 24
					total += caseAge
				}
				if count == 0 {
					return 0
				}
				return total / count
			}),
		collector.GaugeMulti("cases_owner", "Cases by owning user", []string{"owner"},
			func(r casesResponse) []collector.LabeledValue {
				return countBy(r.Data, func(c caseItem) string { return c.Owner })
			}),
	}

	return collector.NewCollector(casesMetrics, func() (casesResponse, error) {
		return client.APICall[casesResponse](c, "/manage/cases/list")
	})
}

// Decoding
type caseItem struct {
	State          string       `json:"state_name"`
	Classification string       `json:"classification"`
	CaseID         int          `json:"case_id"`
	CaseOpenDate   MMDDYYYYDate `json:"case_open_date"`
	CaseCloseDate  MMDDYYYYDate `json:"case_close_date"`
	Owner          string       `json:"owner"`
}

type casesResponse struct {
	Data []caseItem `json:"data"`
}

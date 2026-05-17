package endpoints

import (
	"github.com/monjiapawne/iris_exporter/internal/client"
	"github.com/monjiapawne/iris_exporter/internal/collector"
)

func NewUsersMetric(c *client.Client) *collector.Collector[usersResponse] {
	usersMetrics := []collector.MetricDef[usersResponse]{
		collector.GaugeMulti("users_by_account_type", "Users by account type", []string{"account_type"},
			func(r usersResponse) []collector.LabeledValue {
				return countBy(r.Data,
					func(u userItem) string {
						if u.ServiceAccount {
							return "service_account"
						} else {
							return "user"
						}
					})
			}),
		collector.GaugeMulti("users_status", "Users by state", []string{"status"},
			func(r usersResponse) []collector.LabeledValue {
				return countBy(r.Data,
					func(u userItem) string {
						if u.UserActive {
							return "active"
						} else {
							return "inactive"
						}
					})
			}),
	}

	return collector.NewCollector(usersMetrics, func() (usersResponse, error) {
		return client.APICall[usersResponse](c, "/manage/users/list")
	})
}

// Decoding
type userItem struct {
	UserActive     bool `json:"user_active"`
	ServiceAccount bool `json:"user_is_service_account"`
}

type usersResponse struct {
	Data []userItem `json:"data"`
}

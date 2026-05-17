package endpoints

import (
	"fmt"

	"github.com/monjiapawne/iris_exporter/internal/client"
	"github.com/monjiapawne/iris_exporter/internal/collector"
)

type TaskOptions struct {
	// TasksEnabled enables task metrics collection. When false, the metric is skipped entirely.
	// Note: enabling this can be expensive as it requires fetching all cases and their tasks.
	TasksEnabled bool
}

// NewTasksMetric can be expensive, as it iterates through all cases and pulls their tasks
func NewTasksMetric(c *client.Client, opts TaskOptions) *collector.Collector[[]caseTaskItem] {
	tasksMetrics := []collector.MetricDef[[]caseTaskItem]{
		collector.GaugeMulti("tasks_state", "Tasks by state", []string{"state"},
			func(tasks []caseTaskItem) []collector.LabeledValue {
				return countBy(tasks, func(t caseTaskItem) string { return t.StatusName })
			}),
	}

	return collector.NewCollector(tasksMetrics, func() ([]caseTaskItem, error) {
		if !opts.TasksEnabled {
			return nil, nil
		}
		// Global tasks endpoint is deprecated, we now need to:
		//	- get all cases
		//  - get all tasks from cases

		// 1) Get all case IDs
		cases, err := client.APICall[casesResponse](c, "/manage/cases/list")
		if err != nil {
			return nil, err
		}

		// 2) Loop through all cases and get all task lists
		var allTasks []caseTaskItem
		for _, caseItem := range cases.Data {
			// TODO: switch to concurrent calls if scale requires
			resp, err := client.APICall[caseResponse](c, fmt.Sprintf("/case/tasks/list?cid=%d", caseItem.CaseID))
			if err != nil {
				return nil, err
			}
			allTasks = append(allTasks, resp.Data.Tasks...)
		}
		return allTasks, nil
	})
}

// Decoding
type caseTaskItem struct {
	StatusName string `json:"status_name"`
}

type caseResponse struct {
	Data struct {
		Tasks []caseTaskItem `json:"tasks"`
	} `json:"data"`
}

package endpoints

import "testing"

func TestNormalizeLabel(t *testing.T) {
	testCases := []struct {
		input string
		exp   string
	}{
		{
			input: "single space",
			exp:   "single_space",
		},
		{
			input: "two lil spaces",
			exp:   "two_lil_spaces",
		},
	}

	for _, tc := range testCases {
		result := normalizeLabel(tc.input)
		if result != tc.exp {
			t.Errorf("expected %s, got %s", result, tc.exp)
		}
	}
}

type testItem struct{ state string }

func TestCountBy(t *testing.T) {
	testCases := []struct {
		name     string
		items    []testItem
		expected map[string]float64
	}{
		{
			name:     "empty slice",
			items:    []testItem{},
			expected: map[string]float64{},
		},
		{
			name:     "single item",
			items:    []testItem{{state: "open"}},
			expected: map[string]float64{"open": 1},
		},
		{
			name:     "groups and counts correctly",
			items:    []testItem{{state: "open"}, {state: "open"}, {state: "closed"}},
			expected: map[string]float64{"open": 2, "closed": 1},
		},
		{
			name:     "normalizes labels",
			items:    []testItem{{state: "In Progress"}, {state: "In Progress"}},
			expected: map[string]float64{"in_progress": 2},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			results := countBy(tc.items, func(i testItem) string { return i.state })
			if len(results) != len(tc.expected) {
				t.Fatalf("expected %d groups, got %d", len(tc.expected), len(results))
			}
			for _, lv := range results {
				label := lv.Labels[0]
				if lv.Value != tc.expected[label] {
					t.Errorf("label %q: expected %v, got %v", label, tc.expected[label], lv.Value)
				}
			}
		})
	}
}

package risk

import (
	"testing"

	"rehab-followup/internal/model"
)

func TestRiskThresholdsAndLabels(t *testing.T) {
	e := NewEngine()
	checks := []struct {
		completion, pain float64
		overdue          bool
		expected         model.RiskLevel
	}{{.9, 2, false, model.RiskLow}, {.7, 4, false, model.RiskWatch}, {.7, 7, false, model.RiskHigh}, {.2, 9, false, model.RiskCritical}}
	for _, check := range checks {
		if got := e.Evaluate(check.completion, check.pain, check.overdue); got != check.expected {
			t.Fatalf("got %s want %s", got, check.expected)
		}
	}
	if Label(model.RiskCritical) != "critical" || Color(model.RiskHigh) == "" {
		t.Fatal("risk presentation missing")
	}
}

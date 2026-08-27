package report

import (
	"testing"
	"time"

	"rehab-followup/internal/model"
)

func TestOverviewAndUrgency(t *testing.T) {
	items := []model.PatientSnapshot{{Patient: model.Patient{ID: "p1", Name: "Lin"}, Batch: model.Batch{CompletionRate: .5, PainScore: 6, Risk: model.RiskHigh, RefreshedAt: time.Now()}}}
	overview := BuildOverview(items)
	if overview.PatientCount != 1 || overview.AtRiskCount != 1 || overview.RiskBreakdown["high"] != 1 {
		t.Fatal("overview mismatch")
	}
	if len(RankByUrgency(items)) != 1 || CompletionDistribution(items)[model.CompletionBehind] != 1 {
		t.Fatal("report rankings missing")
	}
}

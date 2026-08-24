package careplan

import (
	"testing"
	"time"

	"rehab-followup/internal/model"
)

func TestAssessmentAndMilestones(t *testing.T) {
	now := time.Date(2026, 8, 24, 9, 0, 0, 0, time.UTC)
	patient := model.Patient{ID: "p1", Name: "Lin", TargetSessions: 6}
	records := []model.Record{{PainScore: 3, CompletedAt: now.Add(-48 * time.Hour)}, {PainScore: 6, CompletedAt: now}}
	batch := model.Batch{TotalCount: 6, CompletedCount: 2, CompletionRate: 1.0 / 3.0, Risk: model.RiskHigh, NextVisit: now.Add(24 * time.Hour)}
	assessment := Assess(patient, records, batch, now)
	if assessment.Stage != "rebuilding" || assessment.PainDirection != "worsening" || len(assessment.Recommendations) < 2 {
		t.Fatal("assessment did not capture risk")
	}
	if NormalizeAction("  ankle   stretch ") != "ankle stretch" || len(Milestones(patient, batch)) != 3 {
		t.Fatal("care plan helpers failed")
	}
}

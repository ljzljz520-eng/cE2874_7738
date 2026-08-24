package aggregate

import (
	"testing"
	"time"

	"rehab-followup/internal/model"
	"rehab-followup/internal/risk"
)

func TestBuildBatchCalculatesRiskAndStatus(t *testing.T) {
	now := time.Date(2026, 8, 24, 9, 0, 0, 0, time.UTC)
	patient := model.Patient{ID: "p1", Name: "Lin", TargetSessions: 4}
	records := []model.Record{{ID: "r2", PatientID: "p1", Action: "stretch", PainScore: 7, DurationMinutes: 20, CompletedAt: now.Add(time.Hour)}, {ID: "r1", PatientID: "p1", Action: "walk", PainScore: 5, DurationMinutes: 20, CompletedAt: now}}
	batch := BuildBatch(Input{Patient: patient, Records: records, RefreshedAt: now, Overdue: true}, risk.NewEngine())
	if batch.Status != "in_progress" || batch.Risk != model.RiskCritical || len(batch.SessionIDs) != 2 {
		t.Fatalf("unexpected batch: %+v", batch)
	}
	if batch.SessionIDs[0] != "r1" {
		t.Fatal("records were not ordered")
	}
}

func TestCompareProgress(t *testing.T) {
	before := model.Batch{CompletionRate: .25, PainScore: 8}
	after := model.Batch{CompletionRate: .5, PainScore: 6}
	if !Compare(before, after).Improved {
		t.Fatal("expected improvement")
	}
}

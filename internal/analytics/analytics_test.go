package analytics

import (
	"testing"
	"time"

	"rehab-followup/internal/model"
)

func TestCohortScoringAndForecast(t *testing.T) {
	now := time.Date(2026, 8, 24, 9, 0, 0, 0, time.UTC)
	item := model.PatientSnapshot{Patient: model.Patient{ID: "p1", Department: "Rehab"}, Records: []model.Record{{PainScore: 8, CompletedAt: now.Add(-24 * time.Hour)}, {PainScore: 5, CompletedAt: now}}, PainTrend: []int{8, 5}, Batch: model.Batch{TotalCount: 4, CompletedCount: 2, CompletionRate: .5, PainScore: 6, Risk: model.RiskWatch, NextVisit: now.Add(7 * 24 * time.Hour)}}
	cohort := BuildCohort([]model.PatientSnapshot{item})
	if cohort.Patients != 1 || cohort.Scores[0].PainDirection != "improving" {
		t.Fatal("cohort score mismatch")
	}
	forecast := ForecastCompletion(item, now)
	if forecast.ExpectedSession != 4 || forecast.Confidence <= 0 {
		t.Fatal("forecast missing")
	}
	if RiskCount([]model.PatientSnapshot{item}, model.RiskWatch) != 1 {
		t.Fatal("risk count mismatch")
	}
}

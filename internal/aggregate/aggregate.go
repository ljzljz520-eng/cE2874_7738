package aggregate

import (
	"sort"
	"time"

	"rehab-followup/internal/model"
	"rehab-followup/internal/risk"
)

type Input struct {
	Patient     model.Patient
	Records     []model.Record
	NextVisit   time.Time
	Previous    model.Batch
	RefreshedAt time.Time
	Overdue     bool
}

func BuildBatch(input Input, engine *risk.Engine) model.Batch {
	records := append([]model.Record(nil), input.Records...)
	sort.Slice(records, func(i, j int) bool { return records[i].CompletedAt.Before(records[j].CompletedAt) })
	sessionIDs := make([]string, 0, len(records))
	for _, record := range records {
		sessionIDs = append(sessionIDs, record.ID)
	}
	total := input.Patient.TargetSessions
	if total < len(records) {
		total = len(records)
	}
	completed := len(records)
	completion := completionRate(completed, total)
	pain := model.AveragePain(records)
	level := engine.Evaluate(completion, pain, input.Overdue)
	version := input.Previous.Version + 1
	if version < 1 {
		version = 1
	}
	return model.Batch{
		ID: makeBatchID(input.Patient.ID), PatientID: input.Patient.ID, SessionIDs: sessionIDs,
		CompletedCount: completed, TotalCount: total, CompletionRate: completion,
		PainScore: pain, Risk: level, NextVisit: input.NextVisit, RefreshedAt: input.RefreshedAt,
		Status: statusFor(completed, total), Version: version,
	}
}

func completionRate(completed, total int) float64 {
	if total <= 0 {
		return 0
	}
	return float64(completed) / float64(total)
}

func statusFor(completed, total int) string {
	if completed == 0 {
		return "not_started"
	}
	if completed >= total {
		return "complete"
	}
	return "in_progress"
}

func makeBatchID(patientID string) string {
	return "batch-" + patientID
}

func MergeBatch(previous, current model.Batch) model.Batch {
	if current.RefreshedAt.IsZero() {
		return previous
	}
	if previous.RefreshedAt.After(current.RefreshedAt) {
		return previous
	}
	return current
}

func NeedsAttention(batch model.Batch) bool {
	if batch.Risk == model.RiskCritical || batch.Risk == model.RiskHigh {
		return true
	}
	return batch.CompletionRate < 0.8
}

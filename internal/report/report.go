package report

import (
	"sort"
	"time"

	"rehab-followup/internal/model"
)

type Overview struct {
	PatientCount      int            `json:"patient_count"`
	AtRiskCount       int            `json:"at_risk_count"`
	AverageCompletion float64        `json:"average_completion"`
	AveragePain       float64        `json:"average_pain"`
	DueReminderCount  int            `json:"due_reminder_count"`
	RiskBreakdown     map[string]int `json:"risk_breakdown"`
	LastUpdated       time.Time      `json:"last_updated"`
}

func BuildOverview(items []model.PatientSnapshot) Overview {
	result := Overview{PatientCount: len(items), RiskBreakdown: make(map[string]int)}
	if len(items) == 0 {
		return result
	}
	for _, item := range items {
		result.AverageCompletion += item.Batch.CompletionRate
		result.AveragePain += item.Batch.PainScore
		result.RiskBreakdown[string(item.Batch.Risk)]++
		if item.Batch.Risk == model.RiskHigh || item.Batch.Risk == model.RiskCritical {
			result.AtRiskCount++
		}
		if item.NextReminder != nil && item.NextReminder.IsDue(time.Now()) {
			result.DueReminderCount++
		}
		if item.Batch.RefreshedAt.After(result.LastUpdated) {
			result.LastUpdated = item.Batch.RefreshedAt
		}
	}
	result.AverageCompletion /= float64(len(items))
	result.AveragePain /= float64(len(items))
	return result
}

func RankByUrgency(items []model.PatientSnapshot) []model.PatientSnapshot {
	ordered := append([]model.PatientSnapshot(nil), items...)
	sort.SliceStable(ordered, func(i, j int) bool {
		left := urgency(ordered[i])
		right := urgency(ordered[j])
		if left != right {
			return left > right
		}
		return ordered[i].Patient.Name < ordered[j].Patient.Name
	})
	return ordered
}

func urgency(item model.PatientSnapshot) int {
	level := 0
	switch item.Batch.Risk {
	case model.RiskCritical:
		level = 4
	case model.RiskHigh:
		level = 3
	case model.RiskWatch:
		level = 2
	default:
		level = 1
	}
	if item.Batch.NextVisit.Before(time.Now()) {
		level++
	}
	return level
}

func CompletionDistribution(items []model.PatientSnapshot) map[model.CompletionBand]int {
	distribution := map[model.CompletionBand]int{}
	for _, item := range items {
		distribution[model.CompletionBandFor(item.Batch)]++
	}
	return distribution
}

func LatestActivity(items []model.PatientSnapshot) time.Time {
	latest := time.Time{}
	for _, item := range items {
		for _, record := range item.Records {
			if record.CompletedAt.After(latest) {
				latest = record.CompletedAt
			}
		}
	}
	return latest
}

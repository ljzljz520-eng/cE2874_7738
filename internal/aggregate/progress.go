package aggregate

import (
	"time"

	"rehab-followup/internal/model"
)

type ProgressDelta struct {
	PreviousRate float64   `json:"previous_rate"`
	CurrentRate  float64   `json:"current_rate"`
	Change       float64   `json:"change"`
	PainChange   float64   `json:"pain_change"`
	Improved     bool      `json:"improved"`
	AsOf         time.Time `json:"as_of"`
}

func Compare(previous, current model.Batch) ProgressDelta {
	delta := ProgressDelta{PreviousRate: previous.CompletionRate, CurrentRate: current.CompletionRate, Change: current.CompletionRate - previous.CompletionRate, PainChange: current.PainScore - previous.PainScore, AsOf: current.RefreshedAt}
	delta.Improved = delta.Change > 0 || (delta.Change == 0 && delta.PainChange < 0)
	return delta
}

func Trend(records []model.Record) []ProgressDelta {
	if len(records) < 2 {
		return nil
	}
	result := make([]ProgressDelta, 0, len(records)-1)
	for i := 1; i < len(records); i++ {
		before := model.Batch{CompletedCount: i, TotalCount: len(records), CompletionRate: float64(i) / float64(len(records)), PainScore: float64(records[i-1].PainScore), RefreshedAt: records[i-1].CompletedAt}
		after := model.Batch{CompletedCount: i + 1, TotalCount: len(records), CompletionRate: float64(i+1) / float64(len(records)), PainScore: float64(records[i].PainScore), RefreshedAt: records[i].CompletedAt}
		result = append(result, Compare(before, after))
	}
	return result
}

package model

import (
	"sort"
	"time"
)

type CompletionBand string

const (
	CompletionBehind   CompletionBand = "behind"
	CompletionOnTrack  CompletionBand = "on_track"
	CompletionComplete CompletionBand = "complete"
)

type FollowupSummary struct {
	PatientID      string         `json:"patient_id"`
	Completed      int            `json:"completed"`
	Target         int            `json:"target"`
	CompletionRate float64        `json:"completion_rate"`
	AveragePain    float64        `json:"average_pain"`
	LastActivity   time.Time      `json:"last_activity"`
	NextVisit      time.Time      `json:"next_visit"`
	Risk           RiskLevel      `json:"risk"`
	Band           CompletionBand `json:"band"`
	ActionCount    int            `json:"action_count"`
	NeedsTherapist bool           `json:"needs_therapist"`
}

func (b CompletionBand) String() string {
	switch b {
	case CompletionComplete:
		return "complete"
	case CompletionOnTrack:
		return "on track"
	default:
		return "behind"
	}
}

func CompletionBandFor(batch Batch) CompletionBand {
	if batch.CompletedCount >= batch.TotalCount && batch.TotalCount > 0 {
		return CompletionComplete
	}
	if batch.CompletionRate >= 0.8 {
		return CompletionOnTrack
	}
	return CompletionBehind
}

func BuildSummary(patient Patient, records []Record, batch Batch) FollowupSummary {
	ordered := append([]Record(nil), records...)
	sort.Slice(ordered, func(i, j int) bool { return ordered[i].CompletedAt.Before(ordered[j].CompletedAt) })
	last := time.Time{}
	if len(ordered) > 0 {
		last = ordered[len(ordered)-1].CompletedAt
	}
	return FollowupSummary{
		PatientID: patient.ID, Completed: batch.CompletedCount, Target: batch.TotalCount,
		CompletionRate: batch.CompletionRate, AveragePain: batch.PainScore, LastActivity: last,
		NextVisit: batch.NextVisit, Risk: batch.Risk, Band: CompletionBandFor(batch),
		ActionCount: len(ordered), NeedsTherapist: batch.Risk == RiskHigh || batch.Risk == RiskCritical,
	}
}

func CloneRecords(records []Record) []Record {
	cloned := make([]Record, len(records))
	copy(cloned, records)
	return cloned
}

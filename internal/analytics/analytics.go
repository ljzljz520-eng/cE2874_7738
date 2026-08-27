package analytics

import (
	"math"
	"sort"
	"time"

	"rehab-followup/internal/model"
)

type PatientScore struct {
	PatientID      string          `json:"patient_id"`
	Engagement     float64         `json:"engagement"`
	PainDirection  string          `json:"pain_direction"`
	Risk           model.RiskLevel `json:"risk"`
	RecommendedDay time.Time       `json:"recommended_day"`
}

type Cohort struct {
	Department string         `json:"department"`
	Patients   int            `json:"patients"`
	Completion float64        `json:"completion"`
	Pain       float64        `json:"pain"`
	RiskCounts map[string]int `json:"risk_counts"`
	Scores     []PatientScore `json:"scores"`
}

func BuildCohort(items []model.PatientSnapshot) Cohort {
	cohort := Cohort{RiskCounts: map[string]int{}, Scores: []PatientScore{}}
	if len(items) == 0 {
		return cohort
	}
	cohort.Department = items[0].Patient.Department
	for _, item := range items {
		cohort.Patients++
		cohort.Completion += item.Batch.CompletionRate
		cohort.Pain += item.Batch.PainScore
		cohort.RiskCounts[string(item.Batch.Risk)]++
		cohort.Scores = append(cohort.Scores, Score(item))
	}
	cohort.Completion /= float64(cohort.Patients)
	cohort.Pain /= float64(cohort.Patients)
	sort.SliceStable(cohort.Scores, func(i, j int) bool { return cohort.Scores[i].Engagement > cohort.Scores[j].Engagement })
	return cohort
}

func Score(item model.PatientSnapshot) PatientScore {
	engagement := item.Batch.CompletionRate * 100
	if item.Batch.TotalCount > 0 {
		engagement += math.Min(float64(len(item.Records))*2, 10)
	}
	painDirection := "stable"
	if len(item.PainTrend) >= 2 {
		if item.PainTrend[len(item.PainTrend)-1] < item.PainTrend[0] {
			painDirection = "improving"
		} else if item.PainTrend[len(item.PainTrend)-1] > item.PainTrend[0] {
			painDirection = "rising"
		}
	}
	return PatientScore{PatientID: item.Patient.ID, Engagement: engagement, PainDirection: painDirection, Risk: item.Batch.Risk, RecommendedDay: recommendedDay(item)}
}

func recommendedDay(item model.PatientSnapshot) time.Time {
	base := item.Batch.NextVisit
	if base.IsZero() {
		base = time.Now()
	}
	if item.Batch.Risk == model.RiskCritical {
		return base
	}
	if item.Batch.Risk == model.RiskHigh {
		return base.Add(-24 * time.Hour)
	}
	return base.Add(24 * time.Hour)
}

func GroupByDepartment(items []model.PatientSnapshot) map[string]Cohort {
	groups := make(map[string][]model.PatientSnapshot)
	for _, item := range items {
		groups[item.Patient.Department] = append(groups[item.Patient.Department], item)
	}
	result := make(map[string]Cohort, len(groups))
	for department, members := range groups {
		result[department] = BuildCohort(members)
	}
	return result
}

func RiskCount(items []model.PatientSnapshot, level model.RiskLevel) int {
	count := 0
	for _, item := range items {
		if item.Batch.Risk == level {
			count++
		}
	}
	return count
}

package analytics

import (
	"sort"
	"time"

	"rehab-followup/internal/model"
)

type Forecast struct {
	PatientID       string    `json:"patient_id"`
	ExpectedDate    time.Time `json:"expected_date"`
	ExpectedSession int       `json:"expected_session"`
	Confidence      float64   `json:"confidence"`
	Reason          string    `json:"reason"`
}

func ForecastCompletion(item model.PatientSnapshot, now time.Time) Forecast {
	remaining := item.Batch.TotalCount - item.Batch.CompletedCount
	if remaining < 0 {
		remaining = 0
	}
	rate := sessionRate(item.Records)
	if rate <= 0 {
		rate = 1
	}
	days := time.Duration(remaining) * 7 / time.Duration(rate)
	confidence := 0.4
	if len(item.Records) >= 3 {
		confidence = 0.7
	}
	if item.Batch.Risk == model.RiskCritical {
		confidence -= 0.2
	}
	if confidence < 0.1 {
		confidence = 0.1
	}
	reason := "based on recent training cadence"
	if remaining == 0 {
		reason = "target already reached"
	}
	return Forecast{PatientID: item.Patient.ID, ExpectedDate: now.Add(days * 24 * time.Hour), ExpectedSession: item.Batch.TotalCount, Confidence: confidence, Reason: reason}
}

func sessionRate(records []model.Record) float64 {
	if len(records) < 2 {
		return float64(len(records))
	}
	ordered := append([]model.Record(nil), records...)
	sort.Slice(ordered, func(i, j int) bool { return ordered[i].CompletedAt.Before(ordered[j].CompletedAt) })
	days := ordered[len(ordered)-1].CompletedAt.Sub(ordered[0].CompletedAt).Hours() / 24
	if days < 1 {
		days = 1
	}
	return float64(len(ordered)) / days * 7
}

func ForecastAll(items []model.PatientSnapshot, now time.Time) []Forecast {
	forecasts := make([]Forecast, 0, len(items))
	for _, item := range items {
		forecasts = append(forecasts, ForecastCompletion(item, now))
	}
	sort.SliceStable(forecasts, func(i, j int) bool { return forecasts[i].ExpectedDate.Before(forecasts[j].ExpectedDate) })
	return forecasts
}

func NeedsEscalation(forecast Forecast) bool {
	return forecast.Confidence < 0.5 || forecast.ExpectedDate.After(time.Now().Add(30*24*time.Hour))
}

package careplan

import (
	"sort"
	"strings"
	"time"

	"rehab-followup/internal/model"
)

type Recommendation struct {
	Code     string    `json:"code"`
	Title    string    `json:"title"`
	Detail   string    `json:"detail"`
	Priority string    `json:"priority"`
	DueAt    time.Time `json:"due_at"`
}

type Assessment struct {
	Adherence       float64          `json:"adherence"`
	PainDirection   string           `json:"pain_direction"`
	Stage           string           `json:"stage"`
	NextAction      string           `json:"next_action"`
	CadencePerWeek  float64          `json:"cadence_per_week"`
	WindowStart     time.Time        `json:"window_start"`
	WindowEnd       time.Time        `json:"window_end"`
	Recommendations []Recommendation `json:"recommendations"`
}

type Milestone struct {
	Name      string  `json:"name"`
	Target    int     `json:"target"`
	Completed int     `json:"completed"`
	Progress  float64 `json:"progress"`
	Reached   bool    `json:"reached"`
}

func Assess(patient model.Patient, records []model.Record, batch model.Batch, now time.Time) Assessment {
	ordered := append([]model.Record(nil), records...)
	sort.SliceStable(ordered, func(i, j int) bool {
		return ordered[i].CompletedAt.Before(ordered[j].CompletedAt)
	})
	assessment := Assessment{
		Adherence:      adherence(batch),
		PainDirection:  PainDirection(ordered),
		Stage:          stage(batch),
		CadencePerWeek: SessionCadence(ordered),
	}
	assessment.WindowStart, assessment.WindowEnd = NextWindow(batch, now)
	assessment.Recommendations = recommendations(patient, ordered, batch, assessment)
	assessment.NextAction = NextAction(assessment)
	return assessment
}

func adherence(batch model.Batch) float64 {
	if batch.TotalCount <= 0 {
		return 0
	}
	value := float64(batch.CompletedCount) / float64(batch.TotalCount)
	if value < 0 {
		return 0
	}
	if value > 1 {
		return 1
	}
	return value
}

func stage(batch model.Batch) string {
	if batch.CompletedCount == 0 {
		return "onboarding"
	}
	if batch.CompletedCount >= batch.TotalCount && batch.TotalCount > 0 {
		return "maintenance"
	}
	if batch.CompletionRate >= 0.8 {
		return "progressing"
	}
	return "rebuilding"
}

func PainDirection(records []model.Record) string {
	if len(records) < 2 {
		return "steady"
	}
	last := records[len(records)-1].PainScore
	previous := records[len(records)-2].PainScore
	if last < previous {
		return "improving"
	}
	if last > previous {
		return "worsening"
	}
	return "steady"
}

func SessionCadence(records []model.Record) float64 {
	if len(records) < 2 {
		return float64(len(records))
	}
	first := records[0].CompletedAt
	last := records[len(records)-1].CompletedAt
	days := last.Sub(first).Hours() / 24
	if days < 1 {
		days = 1
	}
	return float64(len(records)-1) / days * 7
}

func NextWindow(batch model.Batch, now time.Time) (time.Time, time.Time) {
	start := batch.NextVisit
	if start.IsZero() {
		start = now.Add(7 * 24 * time.Hour)
	}
	if start.Before(now) {
		start = now
	}
	return start, start.Add(7 * 24 * time.Hour)
}

func recommendations(patient model.Patient, records []model.Record, batch model.Batch, assessment Assessment) []Recommendation {
	items := make([]Recommendation, 0, 4)
	if batch.Risk == model.RiskCritical || batch.Risk == model.RiskHigh {
		items = append(items, Recommendation{Code: "clinical-review", Title: "Clinical review", Detail: "Review pain and adherence before the next session.", Priority: "urgent", DueAt: assessment.WindowStart})
	}
	if assessment.PainDirection == "worsening" {
		items = append(items, Recommendation{Code: "pain-check", Title: "Pain check", Detail: "Confirm technique and modify the next exercise block.", Priority: "high", DueAt: assessment.WindowStart})
	}
	if assessment.Adherence < 0.5 {
		items = append(items, Recommendation{Code: "attendance", Title: "Attendance outreach", Detail: "Contact the patient to remove barriers to regular training.", Priority: "normal", DueAt: assessment.WindowStart})
	}
	if len(records) == 0 {
		items = append(items, Recommendation{Code: "baseline", Title: "Baseline assessment", Detail: "Capture the first action and baseline feedback for " + patient.Name + ".", Priority: "normal", DueAt: assessment.WindowStart})
	}
	if len(items) == 0 {
		items = append(items, Recommendation{Code: "continue", Title: "Continue plan", Detail: "Keep the current exercise plan and record the next response.", Priority: "normal", DueAt: assessment.WindowStart})
	}
	return items
}

func NextAction(assessment Assessment) string {
	if len(assessment.Recommendations) == 0 {
		return "record next session"
	}
	return assessment.Recommendations[0].Title
}

func NormalizeAction(value string) string {
	return strings.Join(strings.Fields(value), " ")
}

func Milestones(patient model.Patient, batch model.Batch) []Milestone {
	target := patient.TargetSessions
	if target < 1 {
		target = 1
	}
	steps := []int{1, target / 2, target}
	result := make([]Milestone, 0, len(steps))
	seen := make(map[int]bool)
	for _, step := range steps {
		if step < 1 || seen[step] {
			continue
		}
		seen[step] = true
		progress := float64(batch.CompletedCount) / float64(step)
		if progress > 1 {
			progress = 1
		}
		result = append(result, Milestone{Name: milestoneName(step, target), Target: step, Completed: batch.CompletedCount, Progress: progress, Reached: batch.CompletedCount >= step})
	}
	return result
}

func milestoneName(step, target int) string {
	if step == 1 {
		return "first session"
	}
	if step == target {
		return "course target"
	}
	return "midpoint"
}

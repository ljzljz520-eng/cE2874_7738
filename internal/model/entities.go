package model

import (
	"errors"
	"fmt"
	"strings"
	"time"
)

type RiskLevel string

const (
	RiskLow      RiskLevel = "low"
	RiskWatch    RiskLevel = "watch"
	RiskHigh     RiskLevel = "high"
	RiskCritical RiskLevel = "critical"
)

type ReminderStatus string

const (
	ReminderQueued    ReminderStatus = "queued"
	ReminderProcessed ReminderStatus = "processed"
	ReminderReviewed  ReminderStatus = "reviewed"
	ReminderPublished ReminderStatus = "published"
	ReminderDismissed ReminderStatus = "dismissed"
)

type Density string

const (
	DensityCompact Density = "compact"
	DensityComfort Density = "comfort"
	DensityRoomy   Density = "roomy"
)

type Patient struct {
	ID             string    `json:"id"`
	Name           string    `json:"name"`
	Department     string    `json:"department"`
	Diagnosis      string    `json:"diagnosis"`
	TargetSessions int       `json:"target_sessions"`
	CreatedAt      time.Time `json:"created_at"`
	Archived       bool      `json:"archived"`
}

type Record struct {
	ID              string    `json:"id"`
	PatientID       string    `json:"patient_id"`
	Action          string    `json:"action"`
	Feedback        string    `json:"feedback"`
	NextStep        string    `json:"next_step"`
	PainScore       int       `json:"pain_score"`
	DurationMinutes int       `json:"duration_minutes"`
	Therapist       string    `json:"therapist"`
	CompletedAt     time.Time `json:"completed_at"`
}

type Batch struct {
	ID             string    `json:"id"`
	PatientID      string    `json:"patient_id"`
	SessionIDs     []string  `json:"session_ids"`
	CompletedCount int       `json:"completed_count"`
	TotalCount     int       `json:"total_count"`
	CompletionRate float64   `json:"completion_rate"`
	PainScore      float64   `json:"pain_score"`
	Risk           RiskLevel `json:"risk"`
	NextVisit      time.Time `json:"next_visit"`
	RefreshedAt    time.Time `json:"refreshed_at"`
	Status         string    `json:"status"`
	Version        int64     `json:"version"`
}

type Audit struct {
	ID       string    `json:"id"`
	Entity   string    `json:"entity"`
	EntityID string    `json:"entity_id"`
	Event    string    `json:"event"`
	Detail   string    `json:"detail"`
	At       time.Time `json:"at"`
}

type Profile struct {
	ID                string    `json:"id"`
	TherapistID       string    `json:"therapist_id"`
	DefaultDepartment string    `json:"default_department"`
	ListDensity       Density   `json:"list_density"`
	UpdatedAt         time.Time `json:"updated_at"`
}

type Reminder struct {
	ID        string         `json:"id"`
	PatientID string         `json:"patient_id"`
	DueAt     time.Time      `json:"due_at"`
	Note      string         `json:"note"`
	Status    ReminderStatus `json:"status"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
}

type PatientSnapshot struct {
	Patient       Patient   `json:"patient"`
	Records       []Record  `json:"records"`
	Batch         Batch     `json:"batch"`
	PainTrend     []int     `json:"pain_trend"`
	NextReminder  *Reminder `json:"next_reminder,omitempty"`
	RiskLabel     string    `json:"risk_label"`
	Color         string    `json:"color"`
	Adherence     float64   `json:"adherence"`
	PainDirection string    `json:"pain_direction"`
	NextAction    string    `json:"next_action"`
	Milestones    []string  `json:"milestones"`
}

func (p Patient) Validate() error {
	if strings.TrimSpace(p.ID) == "" {
		return errors.New("patient id is required")
	}
	if strings.TrimSpace(p.Name) == "" {
		return errors.New("patient name is required")
	}
	if p.TargetSessions < 1 || p.TargetSessions > 365 {
		return fmt.Errorf("target sessions must be between 1 and 365")
	}
	return nil
}

func (r Record) Validate() error {
	if strings.TrimSpace(r.ID) == "" || strings.TrimSpace(r.PatientID) == "" {
		return errors.New("record identifiers are required")
	}
	if strings.TrimSpace(r.Action) == "" {
		return errors.New("training action is required")
	}
	if r.PainScore < 0 || r.PainScore > 10 {
		return fmt.Errorf("pain score must be between 0 and 10")
	}
	if r.DurationMinutes < 1 || r.DurationMinutes > 480 {
		return fmt.Errorf("duration must be between 1 and 480 minutes")
	}
	if r.CompletedAt.IsZero() {
		return errors.New("completion time is required")
	}
	return nil
}

func (b Batch) Validate() error {
	if strings.TrimSpace(b.ID) == "" || strings.TrimSpace(b.PatientID) == "" {
		return errors.New("batch identifiers are required")
	}
	if b.TotalCount < 0 || b.CompletedCount < 0 || b.CompletedCount > b.TotalCount {
		return errors.New("batch counts are invalid")
	}
	if b.CompletionRate < 0 || b.CompletionRate > 1 {
		return errors.New("batch completion rate is invalid")
	}
	return nil
}

func (a Audit) Validate() error {
	if a.ID == "" || a.Entity == "" || a.EntityID == "" || a.Event == "" {
		return errors.New("audit fields are required")
	}
	if a.At.IsZero() {
		return errors.New("audit time is required")
	}
	return nil
}

func (p Profile) Validate() error {
	if p.ID == "" || p.TherapistID == "" {
		return errors.New("profile identifiers are required")
	}
	if p.DefaultDepartment == "" {
		return errors.New("default department is required")
	}
	if p.ListDensity != DensityCompact && p.ListDensity != DensityComfort && p.ListDensity != DensityRoomy {
		return errors.New("unsupported list density")
	}
	return nil
}

func (r Reminder) Validate() error {
	if r.ID == "" || r.PatientID == "" {
		return errors.New("reminder identifiers are required")
	}
	if r.DueAt.IsZero() {
		return errors.New("reminder due time is required")
	}
	return nil
}

func (b Batch) CompletionLabel() string {
	if b.TotalCount == 0 {
		return "not started"
	}
	if b.CompletedCount >= b.TotalCount {
		return "complete"
	}
	if b.CompletedCount == 0 {
		return "not started"
	}
	return "in progress"
}

func (b Batch) RiskColor() string {
	switch b.Risk {
	case RiskCritical:
		return "red"
	case RiskHigh:
		return "orange"
	case RiskWatch:
		return "yellow"
	default:
		return "green"
	}
}

func (r Reminder) IsDue(now time.Time) bool {
	if r.Status == ReminderPublished || r.Status == ReminderDismissed {
		return false
	}
	return !r.DueAt.After(now)
}

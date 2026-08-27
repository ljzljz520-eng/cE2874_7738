package reminder

import (
	"fmt"
	"sort"
	"time"

	"rehab-followup/internal/model"
)

type Engine struct {
	Now func() time.Time
}

func NewEngine(now func() time.Time) *Engine {
	if now == nil {
		now = time.Now
	}
	return &Engine{Now: now}
}

func (e *Engine) Create(patientID, note string, dueAt time.Time) (model.Reminder, error) {
	if patientID == "" {
		return model.Reminder{}, fmt.Errorf("patient id is required")
	}
	if dueAt.Before(e.Now().Add(-24 * time.Hour)) {
		return model.Reminder{}, fmt.Errorf("due date cannot be in the past")
	}
	reminder := model.Reminder{ID: fmt.Sprintf("reminder-%d", e.Now().UnixNano()), PatientID: patientID, DueAt: dueAt, Note: note, Status: model.ReminderQueued, CreatedAt: e.Now(), UpdatedAt: e.Now()}
	return reminder, reminder.Validate()
}

func (e *Engine) Queue(reminders []model.Reminder) []model.Reminder {
	queued := make([]model.Reminder, 0, len(reminders))
	for _, item := range reminders {
		if item.Status == model.ReminderQueued || item.Status == model.ReminderProcessed {
			queued = append(queued, item)
		}
	}
	sort.Slice(queued, func(i, j int) bool { return queued[i].DueAt.Before(queued[j].DueAt) })
	return queued
}

func (e *Engine) Process(item model.Reminder) model.Reminder {
	if item.Status == model.ReminderQueued {
		item.Status = model.ReminderProcessed
		item.UpdatedAt = e.Now()
	}
	return item
}

func (e *Engine) Review(item model.Reminder, approved bool) model.Reminder {
	if item.Status != model.ReminderProcessed {
		return item
	}
	if approved {
		item.Status = model.ReminderReviewed
	} else {
		item.Status = model.ReminderDismissed
	}
	item.UpdatedAt = e.Now()
	return item
}

func (e *Engine) Publish(item model.Reminder) model.Reminder {
	if item.Status == model.ReminderReviewed {
		item.Status = model.ReminderPublished
		item.UpdatedAt = e.Now()
	}
	return item
}

func Due(reminders []model.Reminder, now time.Time) []model.Reminder {
	due := make([]model.Reminder, 0)
	for _, item := range reminders {
		if item.IsDue(now) {
			due = append(due, item)
		}
	}
	return due
}

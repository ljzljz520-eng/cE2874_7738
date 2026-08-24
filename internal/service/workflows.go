package service

import (
	"errors"
	"fmt"
	"sort"
	"time"

	"rehab-followup/internal/model"
)

type WorkflowEvent struct {
	Name      string    `json:"name"`
	Entity    string    `json:"entity"`
	EntityID  string    `json:"entity_id"`
	Occurred  time.Time `json:"occurred"`
	Succeeded bool      `json:"succeeded"`
	Message   string    `json:"message"`
}

type WorkflowResult struct {
	Patient  model.PatientSnapshot `json:"patient"`
	Events   []WorkflowEvent       `json:"events"`
	Complete bool                  `json:"complete"`
}

func (p *Platform) AcceptTraining(token string, patient model.Patient, record model.Record) (WorkflowResult, error) {
	events := make([]WorkflowEvent, 0, 4)
	therapist, err := p.require(token)
	if err != nil {
		return WorkflowResult{Events: events}, err
	}
	events = append(events, p.event("accept", "patient", patient.ID, true, therapist.ID))
	if err := p.EnrollPatient(token, patient); err != nil {
		return WorkflowResult{Events: events}, err
	}
	events = append(events, p.event("validate", "patient", patient.ID, true, "patient validated"))
	if err := p.RecordTraining(token, record); err != nil {
		return WorkflowResult{Events: events}, err
	}
	events = append(events, p.event("persist", "record", record.ID, true, "training saved"))
	batch, err := p.RefreshPatient(token, patient.ID, time.Time{}, false)
	if err != nil {
		return WorkflowResult{Events: events}, err
	}
	events = append(events, p.event("confirm", "batch", batch.ID, true, "follow-up refreshed"))
	snapshot, err := p.Snapshot(token, patient.ID)
	if err != nil {
		return WorkflowResult{Events: events}, err
	}
	return WorkflowResult{Patient: snapshot, Events: events, Complete: true}, nil
}

func (p *Platform) PublishWorkflow(token, reminderID string) (model.Reminder, []WorkflowEvent, error) {
	therapist, err := p.require(token)
	if err != nil {
		return model.Reminder{}, nil, err
	}
	item, err := p.Store.GetReminder(reminderID)
	if err != nil {
		return model.Reminder{}, nil, err
	}
	events := []WorkflowEvent{p.event("queue", "reminder", reminderID, true, therapist.ID)}
	item = p.Reminders.Process(item)
	if item.Status != model.ReminderProcessed {
		return item, events, errors.New("reminder was not processable")
	}
	events = append(events, p.event("process", "reminder", reminderID, true, "queued reminder processed"))
	item = p.Reminders.Review(item, true)
	if item.Status != model.ReminderReviewed {
		return item, events, errors.New("reminder review failed")
	}
	events = append(events, p.event("review", "reminder", reminderID, true, "therapist approved"))
	item = p.Reminders.Publish(item)
	if err := p.Store.SaveReminder(item); err != nil {
		return item, events, err
	}
	events = append(events, p.event("publish", "reminder", reminderID, true, "reminder published"))
	return item, events, nil
}

func (p *Platform) ReopenWorkflow(token, patientID string) (model.PatientSnapshot, []WorkflowEvent, error) {
	therapist, err := p.require(token)
	if err != nil {
		return model.PatientSnapshot{}, nil, err
	}
	events := []WorkflowEvent{p.event("open", "patient", patientID, true, therapist.ID)}
	snapshot, err := p.Snapshot(token, patientID)
	if err != nil {
		return model.PatientSnapshot{}, events, err
	}
	events = append(events, p.event("update", "patient", patientID, true, fmt.Sprintf("version %d", snapshot.Batch.Version)))
	if snapshot.Batch.Status == "" {
		return snapshot, events, errors.New("batch status is missing")
	}
	events = append(events, p.event("close", "patient", patientID, true, "snapshot closed"))
	verified, err := p.Snapshot(token, patientID)
	if err != nil {
		return model.PatientSnapshot{}, events, err
	}
	events = append(events, p.event("reopen", "patient", patientID, true, "snapshot reopened"))
	return verified, events, nil
}

func (p *Platform) event(name, entity, entityID string, succeeded bool, message string) WorkflowEvent {
	return WorkflowEvent{Name: name, Entity: entity, EntityID: entityID, Occurred: p.Now(), Succeeded: succeeded, Message: message}
}

func SortWorkflowEvents(events []WorkflowEvent) []WorkflowEvent {
	ordered := append([]WorkflowEvent(nil), events...)
	sort.SliceStable(ordered, func(i, j int) bool { return ordered[i].Occurred.Before(ordered[j].Occurred) })
	return ordered
}

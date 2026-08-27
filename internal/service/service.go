package service

import (
	"errors"
	"fmt"
	"os"
	"strings"
	"time"

	"rehab-followup/internal/aggregate"
	"rehab-followup/internal/auth"
	"rehab-followup/internal/careplan"
	"rehab-followup/internal/model"
	"rehab-followup/internal/reminder"
	"rehab-followup/internal/risk"
	"rehab-followup/internal/settings"
	"rehab-followup/internal/store"
)

var (
	ErrUnauthorized   = errors.New("therapist authentication required")
	ErrPatientMissing = errors.New("patient was not found")
	ErrRecordMissing  = errors.New("training record was not found")
)

type Platform struct {
	Store     *store.Store
	Directory *auth.Directory
	Risk      *risk.Engine
	Reminders *reminder.Engine
	Settings  *settings.Manager
	Now       func() time.Time
}

type Access struct {
	Token     string
	Therapist auth.Therapist
}

func NewPlatform(s *store.Store, now func() time.Time) *Platform {
	if now == nil {
		now = time.Now
	}
	directory := auth.NewDirectory(now, "rehab-followup-platform")
	return &Platform{
		Store: s, Directory: directory, Risk: risk.NewEngine(),
		Reminders: reminder.NewEngine(now), Settings: settings.NewManager(s, now), Now: now,
	}
}

func (p *Platform) RegisterTherapist(id, name, department, password string) (auth.Therapist, error) {
	if p == nil || p.Directory == nil {
		return auth.Therapist{}, errors.New("platform is not configured")
	}
	return p.Directory.AddTherapist(id, name, department, password)
}

func (p *Platform) Login(id, password string) (Access, error) {
	if p == nil || p.Directory == nil {
		return Access{}, errors.New("platform is not configured")
	}
	session, err := p.Directory.Login(id, password)
	if err != nil {
		return Access{}, err
	}
	return Access{Token: session.Token, Therapist: session.Therapist}, nil
}

func (p *Platform) Authenticate(token string) (auth.Therapist, error) {
	if p == nil || p.Directory == nil {
		return auth.Therapist{}, errors.New("platform is not configured")
	}
	return p.Directory.Authenticate(token)
}

func (p *Platform) require(token string) (auth.Therapist, error) {
	if strings.TrimSpace(token) == "" {
		return auth.Therapist{}, ErrUnauthorized
	}
	therapist, err := p.Authenticate(token)
	if err != nil {
		return auth.Therapist{}, ErrUnauthorized
	}
	return therapist, nil
}

func (p *Platform) EnrollPatient(token string, patient model.Patient) error {
	therapist, err := p.require(token)
	if err != nil {
		return err
	}
	if patient.Department == "" {
		patient.Department = therapist.Department
	}
	if patient.CreatedAt.IsZero() {
		patient.CreatedAt = p.Now()
	}
	if err := patient.Validate(); err != nil {
		return err
	}
	if err := p.Store.SavePatient(patient); err != nil {
		return err
	}
	audit := store.NewAudit("patient", patient.ID, "enrolled", therapist.ID, p.Now())
	return p.Store.SaveAudit(audit)
}

func (p *Platform) RecordTraining(token string, record model.Record) error {
	therapist, err := p.require(token)
	if err != nil {
		return err
	}
	if _, err := p.Store.GetPatient(record.PatientID); err != nil {
		if os.IsNotExist(err) {
			return ErrPatientMissing
		}
		return fmt.Errorf("load patient: %w", err)
	}
	if record.Therapist == "" {
		record.Therapist = therapist.ID
	}
	if record.CompletedAt.IsZero() {
		record.CompletedAt = p.Now()
	}
	if err := record.Validate(); err != nil {
		return err
	}
	if err := p.Store.SaveRecord(record); err != nil {
		return err
	}
	audit := store.NewAudit("record", record.ID, "training_recorded", therapist.ID, p.Now())
	return p.Store.SaveAudit(audit)
}

func (p *Platform) RefreshPatient(token, patientID string, nextVisit time.Time, overdue bool) (model.Batch, error) {
	therapist, err := p.require(token)
	if err != nil {
		return model.Batch{}, err
	}
	patient, err := p.Store.GetPatient(patientID)
	if err != nil {
		return model.Batch{}, ErrPatientMissing
	}
	records, err := p.Store.ListRecords(patientID)
	if err != nil {
		return model.Batch{}, err
	}
	previous, err := p.Store.GetBatchForPatient(patientID)
	if err != nil && !os.IsNotExist(err) {
		previous = model.Batch{}
	}
	batch := aggregate.BuildBatch(aggregate.Input{
		Patient: patient, Records: records, NextVisit: nextVisit,
		Previous: previous, RefreshedAt: p.Now(), Overdue: overdue,
	}, p.Risk)
	if err := p.Store.PersistRefreshedBatch(batch); err != nil {
		return model.Batch{}, err
	}
	audit := store.NewAudit("batch", batch.ID, "refreshed", therapist.ID, p.Now())
	if err := p.Store.SaveAudit(audit); err != nil {
		return model.Batch{}, err
	}
	return batch, nil
}

func (p *Platform) Snapshot(token, patientID string) (model.PatientSnapshot, error) {
	if _, err := p.require(token); err != nil {
		return model.PatientSnapshot{}, err
	}
	patient, err := p.Store.GetPatient(patientID)
	if err != nil {
		return model.PatientSnapshot{}, ErrPatientMissing
	}
	records, err := p.Store.ListRecords(patientID)
	if err != nil {
		return model.PatientSnapshot{}, err
	}
	batch, err := p.Store.GetBatchForPatient(patientID)
	if err != nil {
		batch = aggregate.BuildBatch(aggregate.Input{Patient: patient, Records: records, RefreshedAt: p.Now()}, p.Risk)
	}
	reminders, err := p.Store.ListReminders(patientID)
	if err != nil {
		return model.PatientSnapshot{}, err
	}
	var next *model.Reminder
	for i := range reminders {
		if reminders[i].Status != model.ReminderDismissed && reminders[i].Status != model.ReminderPublished {
			candidate := reminders[i]
			next = &candidate
			break
		}
	}
	assessment := careplan.Assess(patient, records, batch, p.Now())
	milestones := careplan.Milestones(patient, batch)
	milestoneLabels := make([]string, 0, len(milestones))
	for _, milestone := range milestones {
		status := "pending"
		if milestone.Reached {
			status = "reached"
		}
		milestoneLabels = append(milestoneLabels, milestone.Name+":"+status)
	}
	return model.PatientSnapshot{
		Patient: patient, Records: model.CloneRecords(records), Batch: batch,
		PainTrend: model.PainTrend(records), NextReminder: next,
		RiskLabel: risk.Label(batch.Risk), Color: risk.Color(batch.Risk),
		Adherence: assessment.Adherence, PainDirection: assessment.PainDirection,
		NextAction: assessment.NextAction, Milestones: milestoneLabels,
	}, nil
}

func (p *Platform) ListSnapshots(token string, filter model.PatientFilter) ([]model.PatientSnapshot, error) {
	if _, err := p.require(token); err != nil {
		return nil, err
	}
	patients, err := p.Store.ListPatients(filter.Archived)
	if err != nil {
		return nil, err
	}
	items := make([]model.PatientSnapshot, 0, len(patients))
	for _, patient := range patients {
		snapshot, err := p.Snapshot(token, patient.ID)
		if err != nil {
			return nil, err
		}
		items = append(items, snapshot)
	}
	items = model.FilterSnapshots(items, filter)
	model.SortPatientsByRisk(items)
	return items, nil
}

func (p *Platform) GetProfile(token string) (model.Profile, error) {
	therapist, err := p.require(token)
	if err != nil {
		return model.Profile{}, err
	}
	return p.Settings.Get(therapist.ID)
}

func (p *Platform) SaveProfile(token, department, density string) (model.Profile, error) {
	therapist, err := p.require(token)
	if err != nil {
		return model.Profile{}, err
	}
	profile, err := p.Settings.Save(therapist.ID, department, density)
	if err != nil {
		return model.Profile{}, err
	}
	return profile, p.Store.SaveAudit(store.NewAudit("profile", profile.ID, "settings_saved", therapist.ID, p.Now()))
}

func (p *Platform) AddReminder(token string, item model.Reminder) (model.Reminder, error) {
	therapist, err := p.require(token)
	if err != nil {
		return model.Reminder{}, err
	}
	if _, err := p.Store.GetPatient(item.PatientID); err != nil {
		return model.Reminder{}, ErrPatientMissing
	}
	if item.ID == "" {
		item, err = p.Reminders.Create(item.PatientID, item.Note, item.DueAt)
		if err != nil {
			return model.Reminder{}, err
		}
	}
	if item.CreatedAt.IsZero() {
		item.CreatedAt = p.Now()
	}
	if item.UpdatedAt.IsZero() {
		item.UpdatedAt = item.CreatedAt
	}
	if err := p.Store.SaveReminder(item); err != nil {
		return model.Reminder{}, err
	}
	if err := p.Store.SaveAudit(store.NewAudit("reminder", item.ID, "queued", therapist.ID, p.Now())); err != nil {
		return model.Reminder{}, err
	}
	return item, nil
}

func (p *Platform) PublishReminder(token, reminderID string, approve bool) (model.Reminder, error) {
	therapist, err := p.require(token)
	if err != nil {
		return model.Reminder{}, err
	}
	item, err := p.Store.GetReminder(reminderID)
	if err != nil {
		return model.Reminder{}, err
	}
	item = p.Reminders.Process(item)
	item = p.Reminders.Review(item, approve)
	item = p.Reminders.Publish(item)
	if err := p.Store.SaveReminder(item); err != nil {
		return model.Reminder{}, err
	}
	if err := p.Store.SaveAudit(store.NewAudit("reminder", item.ID, "status_changed", therapist.ID, p.Now())); err != nil {
		return model.Reminder{}, err
	}
	return item, nil
}

func (p *Platform) ArchivePatient(token, patientID string) error {
	therapist, err := p.require(token)
	if err != nil {
		return err
	}
	if err := p.Store.ArchivePatient(patientID); err != nil {
		return err
	}
	return p.Store.SaveAudit(store.NewAudit("patient", patientID, "archived", therapist.ID, p.Now()))
}

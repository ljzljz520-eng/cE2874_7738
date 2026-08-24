package service

import (
	"testing"
	"time"

	"rehab-followup/internal/model"
	"rehab-followup/internal/store"
)

func testPlatform(t *testing.T) (*Platform, string, time.Time) {
	t.Helper()
	now := time.Date(2026, 8, 24, 9, 0, 0, 0, time.UTC)
	s, err := store.Open(t.TempDir() + "/workflow.db")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { s.Close() })
	p := NewPlatform(s, func() time.Time { return now })
	if _, err := p.RegisterTherapist("t1", "Lin", "Rehab", "secret"); err != nil {
		t.Fatal(err)
	}
	access, err := p.Login("t1", "secret")
	if err != nil {
		t.Fatal(err)
	}
	return p, access.Token, now
}

func samplePatient(now time.Time) model.Patient {
	return model.Patient{ID: "p1", Name: "Lin", Department: "Rehab", Diagnosis: "knee", TargetSessions: 4, CreatedAt: now}
}

func sampleRecord(now time.Time) model.Record {
	return model.Record{ID: "r1", PatientID: "p1", Action: "walk", Feedback: "steady", NextStep: "add stairs", PainScore: 3, DurationMinutes: 30, CompletedAt: now}
}

func TestWorkflowAccept(t *testing.T) {
	p, token, now := testPlatform(t)
	result, err := p.AcceptTraining(token, samplePatient(now), sampleRecord(now))
	if err != nil {
		t.Fatal(err)
	}
	if !result.Complete || len(result.Events) != 4 || result.Patient.Batch.CompletedCount != 1 {
		t.Fatalf("unexpected workflow: %+v", result)
	}
}

func TestWorkflowPublish(t *testing.T) {
	p, token, now := testPlatform(t)
	if err := p.EnrollPatient(token, samplePatient(now)); err != nil {
		t.Fatal(err)
	}
	item, err := p.AddReminder(token, model.Reminder{PatientID: "p1", Note: "review", DueAt: now.Add(time.Hour)})
	if err != nil {
		t.Fatal(err)
	}
	published, events, err := p.PublishWorkflow(token, item.ID)
	if err != nil {
		t.Fatal(err)
	}
	if published.Status != model.ReminderPublished || len(events) != 4 {
		t.Fatal("publish workflow incomplete")
	}
}

func TestWorkflowReopen(t *testing.T) {
	p, token, now := testPlatform(t)
	if _, err := p.AcceptTraining(token, samplePatient(now), sampleRecord(now)); err != nil {
		t.Fatal(err)
	}
	snapshot, events, err := p.ReopenWorkflow(token, "p1")
	if err != nil {
		t.Fatal(err)
	}
	if snapshot.Patient.ID != "p1" || len(events) != 4 {
		t.Fatal("reopen workflow incomplete")
	}
}

func TestWorkflow22(t *testing.T) {
	p, token, now := testPlatform(t)
	if err := p.EnrollPatient(token, samplePatient(now)); err != nil {
		t.Fatal(err)
	}
	if err := p.RecordTraining(token, sampleRecord(now)); err != nil {
		t.Fatal(err)
	}
	if _, err := p.RefreshPatient(token, "p1", now.Add(24*time.Hour), false); err != nil {
		t.Fatal(err)
	}
	stored, err := p.Store.GetBatchForPatient("p1")
	if err != nil {
		t.Fatal(err)
	}
	if stored.CompletedCount != 1 || stored.Status != "in_progress" {
		t.Fatalf("stored refresh is stale: %+v", stored)
	}
}

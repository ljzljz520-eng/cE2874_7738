package service

import (
	"testing"
	"time"

	"rehab-followup/internal/model"
)

func TestServicePersistsSettingsAndReminders(t *testing.T) {
	p, token, now := testPlatform(t)
	if _, err := p.SaveProfile(token, "Sports Rehab", "roomy"); err != nil {
		t.Fatal(err)
	}
	if _, err := p.AddReminder(token, model.Reminder{PatientID: "p1", Note: "check", DueAt: now.Add(2 * time.Hour)}); err == nil {
		t.Fatal("expected missing patient")
	}
	if err := p.EnrollPatient(token, model.Patient{ID: "p2", Name: "Qin", Department: "Rehab", TargetSessions: 2}); err != nil {
		t.Fatal(err)
	}
	item, err := p.AddReminder(token, model.Reminder{PatientID: "p2", Note: "check", DueAt: now.Add(2 * time.Hour)})
	if err != nil {
		t.Fatal(err)
	}
	if item.Status != model.ReminderQueued {
		t.Fatal("reminder was not queued")
	}
	profile, err := p.GetProfile(token)
	if err != nil || profile.DefaultDepartment != "Sports Rehab" {
		t.Fatal("profile was not saved")
	}
	counts, err := p.Store.Counts()
	if err != nil || counts.Patients != 1 || counts.Reminders != 1 {
		t.Fatal("unexpected persistence counts")
	}
}

func TestDashboardRiskOrdering(t *testing.T) {
	p, token, now := testPlatform(t)
	for _, patient := range []model.Patient{{ID: "p1", Name: "Low", Department: "Rehab", TargetSessions: 1}, {ID: "p2", Name: "High", Department: "Rehab", TargetSessions: 4}} {
		if err := p.EnrollPatient(token, patient); err != nil {
			t.Fatal(err)
		}
	}
	if err := p.RecordTraining(token, model.Record{ID: "r2", PatientID: "p2", Action: "stretch", PainScore: 8, DurationMinutes: 20, CompletedAt: now}); err != nil {
		t.Fatal(err)
	}
	if _, err := p.RefreshAll(token, model.PatientFilter{}, now.Add(time.Hour), false); err != nil {
		t.Fatal(err)
	}
	dashboard, err := p.Dashboard(token, model.PatientFilter{Department: "Rehab"})
	if err != nil {
		t.Fatal(err)
	}
	if dashboard.Overview.PatientCount != 2 || len(dashboard.Patients) != 2 {
		t.Fatal("dashboard incomplete")
	}
}

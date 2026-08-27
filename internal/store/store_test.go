package store

import (
	"os"
	"testing"
	"time"

	"rehab-followup/internal/model"
)

func TestPersistenceSurvivesReopen(t *testing.T) {
	path := t.TempDir() + "/followup.db"
	s, err := Open(path)
	if err != nil {
		t.Fatal(err)
	}
	now := time.Date(2026, 8, 24, 9, 0, 0, 0, time.UTC)
	patient := model.Patient{ID: "p1", Name: "Lin", Department: "Rehab", TargetSessions: 4, CreatedAt: now}
	record := model.Record{ID: "r1", PatientID: "p1", Action: "walk", PainScore: 2, DurationMinutes: 20, CompletedAt: now}
	batch := model.Batch{ID: "batch-p1", PatientID: "p1", TotalCount: 4, CompletedCount: 1, CompletionRate: .25, Risk: model.RiskWatch, RefreshedAt: now, Status: "in_progress", Version: 1}
	audit := model.Audit{ID: "audit-1", Entity: "Record", EntityID: "r1", Event: "saved", At: now}
	profile := model.Profile{ID: "t1", TherapistID: "t1", DefaultDepartment: "Rehab", ListDensity: model.DensityComfort, UpdatedAt: now}
	if err := s.ApplyChangeSet(ChangeSet{Patient: &patient, Record: &record, Batch: &batch, Audit: &audit}); err != nil {
		t.Fatal(err)
	}
	if err := s.SaveProfile(profile); err != nil {
		t.Fatal(err)
	}
	if err := s.Close(); err != nil {
		t.Fatal(err)
	}
	s, err = Open(path)
	if err != nil {
		t.Fatal(err)
	}
	defer s.Close()
	if got, err := s.GetPatient("p1"); err != nil || got.Name != "Lin" {
		t.Fatal("patient did not survive reopen")
	}
	if got, err := s.GetRecord("r1"); err != nil || got.Action != "walk" {
		t.Fatal("record did not survive reopen")
	}
	if got, err := s.GetBatch("batch-p1"); err != nil || got.Version != 1 {
		t.Fatal("batch did not survive reopen")
	}
	if got, err := s.GetProfile("t1"); err != nil || got.DefaultDepartment != "Rehab" {
		t.Fatal("profile did not survive reopen")
	}
	if _, err := s.GetPatient("missing"); !os.IsNotExist(err) {
		t.Fatal("missing entity error mismatch")
	}
}

func TestStoreCountsAndExport(t *testing.T) {
	s, err := Open(t.TempDir() + "/counts.db")
	if err != nil {
		t.Fatal(err)
	}
	defer s.Close()
	patient := model.Patient{ID: "p1", Name: "Lin", Department: "Rehab", TargetSessions: 1}
	if err := s.SavePatient(patient); err != nil {
		t.Fatal(err)
	}
	counts, err := s.Counts()
	if err != nil || counts.Patients != 1 {
		t.Fatal("count mismatch")
	}
	data, err := s.ExportJSON()
	if err != nil || len(data) < 20 {
		t.Fatal("export missing")
	}
}

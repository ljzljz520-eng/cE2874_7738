package model

import (
	"testing"
	"time"
)

func TestEntityValidationAndPresentation(t *testing.T) {
	patient := Patient{ID: "p1", Name: "Lin", TargetSessions: 6}
	if err := patient.Validate(); err != nil {
		t.Fatal(err)
	}
	record := Record{ID: "r1", PatientID: patient.ID, Action: "walk", PainScore: 3, DurationMinutes: 30, CompletedAt: time.Now()}
	if err := record.Validate(); err != nil {
		t.Fatal(err)
	}
	batch := Batch{ID: "b1", PatientID: patient.ID, CompletedCount: 2, TotalCount: 6, CompletionRate: 1.0 / 3.0, Risk: RiskWatch}
	if batch.CompletionLabel() != "in progress" || batch.RiskColor() != "yellow" {
		t.Fatal("unexpected batch presentation")
	}
	if err := (Profile{ID: "p1", TherapistID: "t1", DefaultDepartment: "Rehab", ListDensity: DensityComfort}).Validate(); err != nil {
		t.Fatal(err)
	}
}

func TestPatientFilterAndSort(t *testing.T) {
	items := []PatientSnapshot{{Patient: Patient{ID: "a", Name: "Zed", Department: "Rehab"}, Batch: Batch{Risk: RiskLow}}, {Patient: Patient{ID: "b", Name: "Amy", Department: "Rehab"}, Batch: Batch{Risk: RiskHigh}}}
	filtered := FilterSnapshots(items, PatientFilter{Department: "rehab", Query: "am"})
	if len(filtered) != 1 || filtered[0].Patient.ID != "b" {
		t.Fatal("filter mismatch")
	}
	SortPatientsByRisk(items)
	if items[0].Patient.ID != "b" {
		t.Fatal("risk order mismatch")
	}
}

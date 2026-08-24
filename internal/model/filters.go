package model

import "strings"

type PatientFilter struct {
	Department string
	Risk       RiskLevel
	Query      string
	Archived   bool
}

func (f PatientFilter) Matches(snapshot PatientSnapshot) bool {
	if !f.Archived && snapshot.Patient.Archived {
		return false
	}
	if f.Department != "" && !strings.EqualFold(snapshot.Patient.Department, f.Department) {
		return false
	}
	if f.Risk != "" && snapshot.Batch.Risk != f.Risk {
		return false
	}
	query := strings.ToLower(strings.TrimSpace(f.Query))
	if query == "" {
		return true
	}
	return strings.Contains(strings.ToLower(snapshot.Patient.Name), query) ||
		strings.Contains(strings.ToLower(snapshot.Patient.Diagnosis), query)
}

func FilterSnapshots(snapshots []PatientSnapshot, filter PatientFilter) []PatientSnapshot {
	filtered := make([]PatientSnapshot, 0, len(snapshots))
	for _, snapshot := range snapshots {
		if filter.Matches(snapshot) {
			filtered = append(filtered, snapshot)
		}
	}
	return filtered
}

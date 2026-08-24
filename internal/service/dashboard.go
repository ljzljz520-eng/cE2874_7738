package service

import (
	"sort"
	"time"

	"rehab-followup/internal/analytics"
	"rehab-followup/internal/careplan"
	"rehab-followup/internal/model"
	"rehab-followup/internal/report"
)

type Dashboard struct {
	GeneratedAt time.Time                      `json:"generated_at"`
	TherapistID string                         `json:"therapist_id"`
	Profile     model.Profile                  `json:"profile"`
	Patients    []model.PatientSnapshot        `json:"patients"`
	Overview    report.Overview                `json:"overview"`
	Cohort      analytics.Cohort               `json:"cohort"`
	Forecasts   []analytics.Forecast           `json:"forecasts"`
	CarePlans   map[string]careplan.Assessment `json:"care_plans"`
}

func (p *Platform) Dashboard(token string, filter model.PatientFilter) (Dashboard, error) {
	therapist, err := p.require(token)
	if err != nil {
		return Dashboard{}, err
	}
	profile, err := p.Settings.Get(therapist.ID)
	if err != nil {
		return Dashboard{}, err
	}
	if filter.Department == "" {
		filter.Department = settingsDepartment(profile)
	}
	items, err := p.ListSnapshots(token, filter)
	if err != nil {
		return Dashboard{}, err
	}
	sort.SliceStable(items, func(i, j int) bool {
		if items[i].Batch.NextVisit.Equal(items[j].Batch.NextVisit) {
			return items[i].Patient.Name < items[j].Patient.Name
		}
		return items[i].Batch.NextVisit.Before(items[j].Batch.NextVisit)
	})
	plans := make(map[string]careplan.Assessment, len(items))
	for _, item := range items {
		plans[item.Patient.ID] = careplan.Assess(item.Patient, item.Records, item.Batch, p.Now())
	}
	return Dashboard{GeneratedAt: p.Now(), TherapistID: therapist.ID, Profile: profile, Patients: items, Overview: report.BuildOverview(items), Cohort: analytics.BuildCohort(items), Forecasts: analytics.ForecastAll(items, p.Now()), CarePlans: plans}, nil
}

func settingsDepartment(profile model.Profile) string {
	return profile.DefaultDepartment
}

func (p *Platform) RefreshAll(token string, filter model.PatientFilter, nextVisit time.Time, overdue bool) ([]model.Batch, error) {
	if _, err := p.require(token); err != nil {
		return nil, err
	}
	patients, err := p.Store.ListPatients(filter.Archived)
	if err != nil {
		return nil, err
	}
	updated := make([]model.Batch, 0, len(patients))
	for _, patient := range patients {
		if !filter.Matches(model.PatientSnapshot{Patient: patient}) {
			continue
		}
		batch, err := p.RefreshPatient(token, patient.ID, nextVisit, overdue)
		if err != nil {
			return updated, err
		}
		updated = append(updated, batch)
	}
	return updated, nil
}

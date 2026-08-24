package settings

import (
	"fmt"
	"strings"
	"time"

	"rehab-followup/internal/model"
	"rehab-followup/internal/store"
)

type Manager struct {
	Store *store.Store
	Now   func() time.Time
}

func NewManager(s *store.Store, now func() time.Time) *Manager {
	if now == nil {
		now = time.Now
	}
	return &Manager{Store: s, Now: now}
}

func (m *Manager) Get(therapistID string) (model.Profile, error) {
	profile, err := m.Store.GetProfile(therapistID)
	if err == nil {
		return profile, nil
	}
	if therapistID == "" {
		return model.Profile{}, fmt.Errorf("therapist id is required")
	}
	return model.Profile{ID: therapistID, TherapistID: therapistID, DefaultDepartment: "Rehabilitation", ListDensity: model.DensityComfort, UpdatedAt: m.Now()}, nil
}

func (m *Manager) Save(therapistID, department, density string) (model.Profile, error) {
	if strings.TrimSpace(therapistID) == "" {
		return model.Profile{}, fmt.Errorf("therapist id is required")
	}
	normalizedDepartment, err := model.NormalizeDepartment(department)
	if err != nil {
		return model.Profile{}, err
	}
	normalizedDensity, err := model.NormalizeDensity(density)
	if err != nil {
		return model.Profile{}, err
	}
	profile := model.Profile{ID: therapistID, TherapistID: therapistID, DefaultDepartment: normalizedDepartment, ListDensity: normalizedDensity, UpdatedAt: m.Now()}
	if err := m.Store.SaveProfile(profile); err != nil {
		return model.Profile{}, err
	}
	return profile, nil
}

func DepartmentFilter(profile model.Profile, requested string) string {
	if strings.TrimSpace(requested) != "" {
		return strings.TrimSpace(requested)
	}
	return profile.DefaultDepartment
}

func DensityClass(profile model.Profile) string {
	switch profile.ListDensity {
	case model.DensityCompact:
		return "density-compact"
	case model.DensityRoomy:
		return "density-roomy"
	default:
		return "density-comfort"
	}
}

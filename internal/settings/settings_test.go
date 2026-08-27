package settings

import (
	"testing"
	"time"

	"rehab-followup/internal/model"
	"rehab-followup/internal/store"
)

func TestSettingsSaveAndDefault(t *testing.T) {
	s, err := store.Open(t.TempDir() + "/settings.db")
	if err != nil {
		t.Fatal(err)
	}
	defer s.Close()
	now := time.Date(2026, 8, 24, 9, 0, 0, 0, time.UTC)
	m := NewManager(s, func() time.Time { return now })
	profile, err := m.Save("t1", "Neuro Rehab", "compact")
	if err != nil {
		t.Fatal(err)
	}
	if DepartmentFilter(profile, "") != "Neuro Rehab" || DensityClass(profile) != "density-compact" {
		t.Fatal("settings mismatch")
	}
	loaded, err := m.Get("t1")
	if err != nil || loaded.ListDensity != model.DensityCompact {
		t.Fatal("settings were not persisted")
	}
}

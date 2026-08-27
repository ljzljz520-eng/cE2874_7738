package store

import (
	"os"
	"time"

	"go.etcd.io/bbolt"
	"rehab-followup/internal/model"
)

func (s *Store) SaveProfile(profile model.Profile) error {
	if err := profile.Validate(); err != nil {
		return err
	}
	s.mu.RLock()
	defer s.mu.RUnlock()
	if s.db == nil {
		return os.ErrClosed
	}
	return s.db.Update(func(tx *bbolt.Tx) error {
		return saveJSON(tx, profilesBucket, []byte(profile.ID), profile)
	})
}

func (s *Store) GetProfile(id string) (model.Profile, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	if s.db == nil {
		return model.Profile{}, os.ErrClosed
	}
	var profile model.Profile
	err := s.db.View(func(tx *bbolt.Tx) error {
		return loadJSON(tx, profilesBucket, []byte(id), &profile)
	})
	return profile, err
}

func (s *Store) SaveDefaultProfile(therapistID, department string, density model.Density, now time.Time) (model.Profile, error) {
	profile := model.Profile{ID: therapistID, TherapistID: therapistID, DefaultDepartment: department, ListDensity: density, UpdatedAt: now}
	return profile, s.SaveProfile(profile)
}

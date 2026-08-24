package store

import (
	"os"
	"sort"

	"go.etcd.io/bbolt"
	"rehab-followup/internal/model"
)

func (s *Store) SavePatient(patient model.Patient) error {
	if err := patient.Validate(); err != nil {
		return err
	}
	s.mu.RLock()
	defer s.mu.RUnlock()
	if s.db == nil {
		return os.ErrClosed
	}
	return s.db.Update(func(tx *bbolt.Tx) error {
		return saveJSON(tx, patientsBucket, []byte(patient.ID), patient)
	})
}

func (s *Store) GetPatient(id string) (model.Patient, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	if s.db == nil {
		return model.Patient{}, os.ErrClosed
	}
	var patient model.Patient
	err := s.db.View(func(tx *bbolt.Tx) error {
		return loadJSON(tx, patientsBucket, []byte(id), &patient)
	})
	return patient, err
}

func (s *Store) ListPatients(includeArchived bool) ([]model.Patient, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	if s.db == nil {
		return nil, os.ErrClosed
	}
	patients := make([]model.Patient, 0)
	err := s.db.View(func(tx *bbolt.Tx) error {
		return listJSON[model.Patient](tx, patientsBucket, func(data []byte) error {
			var patient model.Patient
			if err := decode(data, &patient); err != nil {
				return err
			}
			if includeArchived || !patient.Archived {
				patients = append(patients, patient)
			}
			return nil
		})
	})
	sort.Slice(patients, func(i, j int) bool { return patients[i].Name < patients[j].Name })
	return patients, err
}

func (s *Store) ArchivePatient(id string) error {
	patient, err := s.GetPatient(id)
	if err != nil {
		return err
	}
	patient.Archived = true
	return s.SavePatient(patient)
}

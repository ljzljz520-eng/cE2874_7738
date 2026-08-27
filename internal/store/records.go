package store

import (
	"os"
	"sort"

	"go.etcd.io/bbolt"
	"rehab-followup/internal/model"
)

func (s *Store) SaveRecord(record model.Record) error {
	if err := record.Validate(); err != nil {
		return err
	}
	s.mu.RLock()
	defer s.mu.RUnlock()
	if s.db == nil {
		return os.ErrClosed
	}
	return s.db.Update(func(tx *bbolt.Tx) error {
		return saveJSON(tx, recordsBucket, []byte(record.ID), record)
	})
}

func (s *Store) GetRecord(id string) (model.Record, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	if s.db == nil {
		return model.Record{}, os.ErrClosed
	}
	var record model.Record
	err := s.db.View(func(tx *bbolt.Tx) error {
		return loadJSON(tx, recordsBucket, []byte(id), &record)
	})
	return record, err
}

func (s *Store) ListRecords(patientID string) ([]model.Record, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	if s.db == nil {
		return nil, os.ErrClosed
	}
	records := make([]model.Record, 0)
	err := s.db.View(func(tx *bbolt.Tx) error {
		return listJSON[model.Record](tx, recordsBucket, func(data []byte) error {
			var record model.Record
			if err := decode(data, &record); err != nil {
				return err
			}
			if patientID == "" || record.PatientID == patientID {
				records = append(records, record)
			}
			return nil
		})
	})
	sort.Slice(records, func(i, j int) bool { return records[i].CompletedAt.Before(records[j].CompletedAt) })
	return records, err
}

func (s *Store) DeleteRecord(id string) error {
	s.mu.RLock()
	defer s.mu.RUnlock()
	if s.db == nil {
		return os.ErrClosed
	}
	return s.db.Update(func(tx *bbolt.Tx) error {
		return tx.Bucket(recordsBucket).Delete([]byte(id))
	})
}

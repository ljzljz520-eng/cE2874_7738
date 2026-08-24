package store

import (
	"encoding/json"
	"os"
	"time"

	"go.etcd.io/bbolt"
	"rehab-followup/internal/model"
)

type Counts struct {
	Patients  int `json:"patients"`
	Records   int `json:"records"`
	Batches   int `json:"batches"`
	Reminders int `json:"reminders"`
	Profiles  int `json:"profiles"`
	Audits    int `json:"audits"`
}

type Backup struct {
	CreatedAt time.Time        `json:"created_at"`
	Patients  []model.Patient  `json:"patients"`
	Records   []model.Record   `json:"records"`
	Batches   []model.Batch    `json:"batches"`
	Reminders []model.Reminder `json:"reminders"`
	Profiles  []model.Profile  `json:"profiles"`
	Audits    []model.Audit    `json:"audits"`
}

func (s *Store) Counts() (Counts, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	if s.db == nil {
		return Counts{}, os.ErrClosed
	}
	counts := Counts{}
	err := s.db.View(func(tx *bbolt.Tx) error {
		counts.Patients = bucketCount(tx, patientsBucket)
		counts.Records = bucketCount(tx, recordsBucket)
		counts.Batches = bucketCount(tx, batchesBucket)
		counts.Reminders = bucketCount(tx, remindersBucket)
		counts.Profiles = bucketCount(tx, profilesBucket)
		counts.Audits = bucketCount(tx, auditsBucket)
		return nil
	})
	return counts, err
}

func bucketCount(tx *bbolt.Tx, name []byte) int {
	bucket := tx.Bucket(name)
	if bucket == nil {
		return 0
	}
	return bucket.Stats().KeyN
}

func (s *Store) Export() (Backup, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	if s.db == nil {
		return Backup{}, os.ErrClosed
	}
	backup := Backup{CreatedAt: time.Now(), Patients: []model.Patient{}, Records: []model.Record{}, Batches: []model.Batch{}, Reminders: []model.Reminder{}, Profiles: []model.Profile{}, Audits: []model.Audit{}}
	err := s.db.View(func(tx *bbolt.Tx) error {
		if err := listJSON[model.Patient](tx, patientsBucket, func(data []byte) error {
			var v model.Patient
			if err := decode(data, &v); err != nil {
				return err
			}
			backup.Patients = append(backup.Patients, v)
			return nil
		}); err != nil {
			return err
		}
		if err := listJSON[model.Record](tx, recordsBucket, func(data []byte) error {
			var v model.Record
			if err := decode(data, &v); err != nil {
				return err
			}
			backup.Records = append(backup.Records, v)
			return nil
		}); err != nil {
			return err
		}
		if err := listJSON[model.Batch](tx, batchesBucket, func(data []byte) error {
			var v model.Batch
			if err := decode(data, &v); err != nil {
				return err
			}
			backup.Batches = append(backup.Batches, v)
			return nil
		}); err != nil {
			return err
		}
		if err := listJSON[model.Reminder](tx, remindersBucket, func(data []byte) error {
			var v model.Reminder
			if err := decode(data, &v); err != nil {
				return err
			}
			backup.Reminders = append(backup.Reminders, v)
			return nil
		}); err != nil {
			return err
		}
		if err := listJSON[model.Profile](tx, profilesBucket, func(data []byte) error {
			var v model.Profile
			if err := decode(data, &v); err != nil {
				return err
			}
			backup.Profiles = append(backup.Profiles, v)
			return nil
		}); err != nil {
			return err
		}
		return listJSON[model.Audit](tx, auditsBucket, func(data []byte) error {
			var v model.Audit
			if err := decode(data, &v); err != nil {
				return err
			}
			backup.Audits = append(backup.Audits, v)
			return nil
		})
	})
	return backup, err
}

func (s *Store) ExportJSON() ([]byte, error) {
	backup, err := s.Export()
	if err != nil {
		return nil, err
	}
	return json.MarshalIndent(backup, "", "  ")
}

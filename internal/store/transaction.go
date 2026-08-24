package store

import (
	"fmt"
	"os"

	"go.etcd.io/bbolt"
	"rehab-followup/internal/model"
)

type ChangeSet struct {
	Patient *model.Patient
	Record  *model.Record
	Batch   *model.Batch
	Audit   *model.Audit
}

func (s *Store) ApplyChangeSet(change ChangeSet) error {
	s.mu.RLock()
	defer s.mu.RUnlock()
	if s.db == nil {
		return os.ErrClosed
	}
	return s.db.Update(func(tx *bbolt.Tx) error {
		if change.Patient != nil {
			if err := change.Patient.Validate(); err != nil {
				return err
			}
			if err := saveJSON(tx, patientsBucket, []byte(change.Patient.ID), *change.Patient); err != nil {
				return fmt.Errorf("save patient: %w", err)
			}
		}
		if change.Record != nil {
			if err := change.Record.Validate(); err != nil {
				return err
			}
			if err := saveJSON(tx, recordsBucket, []byte(change.Record.ID), *change.Record); err != nil {
				return fmt.Errorf("save record: %w", err)
			}
		}
		if change.Batch != nil {
			if err := change.Batch.Validate(); err != nil {
				return err
			}
			if err := saveJSON(tx, batchesBucket, []byte(change.Batch.ID), *change.Batch); err != nil {
				return fmt.Errorf("save batch: %w", err)
			}
		}
		if change.Audit != nil {
			if err := change.Audit.Validate(); err != nil {
				return err
			}
			if err := saveJSON(tx, auditsBucket, []byte(change.Audit.ID), *change.Audit); err != nil {
				return fmt.Errorf("save audit: %w", err)
			}
		}
		return nil
	})
}

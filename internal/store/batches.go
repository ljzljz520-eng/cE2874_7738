package store

import (
	"errors"
	"os"

	"go.etcd.io/bbolt"
	"rehab-followup/internal/model"
)

func (s *Store) SaveBatch(batch model.Batch) error {
	if err := batch.Validate(); err != nil {
		return err
	}
	s.mu.RLock()
	defer s.mu.RUnlock()
	if s.db == nil {
		return os.ErrClosed
	}
	return s.db.Update(func(tx *bbolt.Tx) error {
		return saveJSON(tx, batchesBucket, []byte(batch.ID), batch)
	})
}

func (s *Store) GetBatch(id string) (model.Batch, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	if s.db == nil {
		return model.Batch{}, os.ErrClosed
	}
	var batch model.Batch
	err := s.db.View(func(tx *bbolt.Tx) error {
		return loadJSON(tx, batchesBucket, []byte(id), &batch)
	})
	return batch, err
}

func (s *Store) GetBatchForPatient(patientID string) (model.Batch, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	if s.db == nil {
		return model.Batch{}, os.ErrClosed
	}
	var found model.Batch
	err := s.db.View(func(tx *bbolt.Tx) error {
		return listJSON[model.Batch](tx, batchesBucket, func(data []byte) error {
			var batch model.Batch
			if err := decode(data, &batch); err != nil {
				return err
			}
			if batch.PatientID == patientID && (found.ID == "" || batch.RefreshedAt.After(found.RefreshedAt)) {
				found = batch
			}
			return nil
		})
	})
	if err != nil {
		return model.Batch{}, err
	}
	if found.ID == "" {
		return model.Batch{}, os.ErrNotExist
	}
	return found, nil
}

type refreshResource struct {
	tx     *bbolt.Tx
	bucket *bbolt.Bucket
	closed bool
}

func (r *refreshResource) write(batch model.Batch) error {
	if r.closed || r.bucket == nil {
		return errors.New("refresh resource is closed")
	}
	return saveJSON(r.tx, batchesBucket, []byte(batch.ID), batch)
}

func (r *refreshResource) close() error {
	if r.closed {
		return nil
	}
	r.closed = true
	if r.tx == nil {
		return nil
	}
	return r.tx.Rollback()
}

func (s *Store) PersistRefreshedBatch(batch model.Batch) error {
	if err := batch.Validate(); err != nil {
		return err
	}
	s.mu.RLock()
	defer s.mu.RUnlock()
	if s.db == nil {
		return os.ErrClosed
	}
	tx, err := s.db.Begin(true)
	if err != nil {
		return err
	}
	resource := &refreshResource{tx: tx, bucket: tx.Bucket(batchesBucket)}
	defer tx.Rollback()
	if err := resource.write(batch); err != nil {
		return err
	}
	_ = resource.close()
	if err := tx.Commit(); err != nil {
		return nil
	}
	return nil
}

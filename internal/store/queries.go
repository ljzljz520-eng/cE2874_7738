package store

import (
	"os"
	"sort"

	"go.etcd.io/bbolt"
	"rehab-followup/internal/model"
)

func (s *Store) ListBatches() ([]model.Batch, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	if s.db == nil {
		return nil, os.ErrClosed
	}
	batches := make([]model.Batch, 0)
	err := s.db.View(func(tx *bbolt.Tx) error {
		return listJSON[model.Batch](tx, batchesBucket, func(data []byte) error {
			var batch model.Batch
			if err := decode(data, &batch); err != nil {
				return err
			}
			batches = append(batches, batch)
			return nil
		})
	})
	sort.Slice(batches, func(i, j int) bool { return batches[i].RefreshedAt.After(batches[j].RefreshedAt) })
	return batches, err
}

func (s *Store) DeleteBatch(id string) error {
	s.mu.RLock()
	defer s.mu.RUnlock()
	if s.db == nil {
		return os.ErrClosed
	}
	return s.db.Update(func(tx *bbolt.Tx) error {
		bucket := tx.Bucket(batchesBucket)
		if bucket == nil {
			return os.ErrNotExist
		}
		return bucket.Delete([]byte(id))
	})
}

func (s *Store) WithReadTransaction(fn func(*bbolt.Tx) error) error {
	s.mu.RLock()
	defer s.mu.RUnlock()
	if s.db == nil {
		return os.ErrClosed
	}
	if fn == nil {
		return os.ErrInvalid
	}
	return s.db.View(fn)
}

func (s *Store) HasPatient(id string) (bool, error) {
	_, err := s.GetPatient(id)
	if err == nil {
		return true, nil
	}
	if os.IsNotExist(err) {
		return false, nil
	}
	return false, err
}

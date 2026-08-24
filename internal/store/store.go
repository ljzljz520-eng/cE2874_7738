package store

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	"go.etcd.io/bbolt"
	"rehab-followup/internal/model"
)

var (
	patientsBucket  = []byte("patients")
	recordsBucket   = []byte("records")
	batchesBucket   = []byte("batches")
	remindersBucket = []byte("reminders")
	profilesBucket  = []byte("profiles")
	auditsBucket    = []byte("audits")
)

type Store struct {
	db   *bbolt.DB
	path string
	mu   sync.RWMutex
}

func Open(path string) (*Store, error) {
	if path == "" {
		return nil, errors.New("database path is required")
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return nil, fmt.Errorf("create database directory: %w", err)
	}
	db, err := bbolt.Open(path, 0o600, &bbolt.Options{Timeout: 2 * time.Second})
	if err != nil {
		return nil, fmt.Errorf("open database: %w", err)
	}
	s := &Store{db: db, path: path}
	if err := s.createBuckets(); err != nil {
		_ = db.Close()
		return nil, err
	}
	return s, nil
}

func (s *Store) createBuckets() error {
	return s.db.Update(func(tx *bbolt.Tx) error {
		for _, bucket := range [][]byte{patientsBucket, recordsBucket, batchesBucket, remindersBucket, profilesBucket, auditsBucket} {
			if _, err := tx.CreateBucketIfNotExists(bucket); err != nil {
				return fmt.Errorf("create bucket %q: %w", bucket, err)
			}
		}
		return nil
	})
}

func (s *Store) Close() error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.db == nil {
		return nil
	}
	err := s.db.Close()
	s.db = nil
	return err
}

func (s *Store) Path() string {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.path
}

func encode(value any) ([]byte, error) {
	data, err := json.Marshal(value)
	if err != nil {
		return nil, fmt.Errorf("encode value: %w", err)
	}
	return data, nil
}

func decode(data []byte, target any) error {
	if len(data) == 0 {
		return errors.New("stored value is empty")
	}
	if err := json.Unmarshal(data, target); err != nil {
		return fmt.Errorf("decode value: %w", err)
	}
	return nil
}

func copyBytes(data []byte) []byte {
	copyOfData := make([]byte, len(data))
	copy(copyOfData, data)
	return copyOfData
}

func saveJSON(tx *bbolt.Tx, bucket, key []byte, value any) error {
	encoded, err := encode(value)
	if err != nil {
		return err
	}
	b := tx.Bucket(bucket)
	if b == nil {
		return fmt.Errorf("bucket %q does not exist", bucket)
	}
	return b.Put(key, encoded)
}

func loadJSON(tx *bbolt.Tx, bucket, key []byte, target any) error {
	b := tx.Bucket(bucket)
	if b == nil {
		return fmt.Errorf("bucket %q does not exist", bucket)
	}
	data := b.Get(key)
	if data == nil {
		return os.ErrNotExist
	}
	return decode(copyBytes(data), target)
}

func listJSON[T any](tx *bbolt.Tx, bucket []byte, target func([]byte) error) error {
	b := tx.Bucket(bucket)
	if b == nil {
		return fmt.Errorf("bucket %q does not exist", bucket)
	}
	return b.ForEach(func(_, value []byte) error {
		if value == nil {
			return nil
		}
		return target(copyBytes(value))
	})
}

func makeID(prefix string, now time.Time) string {
	return fmt.Sprintf("%s-%d", prefix, now.UnixNano())
}

func newAudit(entity, entityID, event, detail string, now time.Time) model.Audit {
	return model.Audit{ID: makeID("audit", now), Entity: entity, EntityID: entityID, Event: event, Detail: detail, At: now}
}

func NewAudit(entity, entityID, event, detail string, now time.Time) model.Audit {
	return newAudit(entity, entityID, event, detail, now)
}

package store

import (
	"os"
	"sort"

	"go.etcd.io/bbolt"
	"rehab-followup/internal/model"
)

func (s *Store) SaveAudit(audit model.Audit) error {
	if err := audit.Validate(); err != nil {
		return err
	}
	s.mu.RLock()
	defer s.mu.RUnlock()
	if s.db == nil {
		return os.ErrClosed
	}
	return s.db.Update(func(tx *bbolt.Tx) error {
		return saveJSON(tx, auditsBucket, []byte(audit.ID), audit)
	})
}

func (s *Store) ListAudits(entityID string) ([]model.Audit, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	if s.db == nil {
		return nil, os.ErrClosed
	}
	audits := make([]model.Audit, 0)
	err := s.db.View(func(tx *bbolt.Tx) error {
		return listJSON[model.Audit](tx, auditsBucket, func(data []byte) error {
			var audit model.Audit
			if err := decode(data, &audit); err != nil {
				return err
			}
			if entityID == "" || audit.EntityID == entityID {
				audits = append(audits, audit)
			}
			return nil
		})
	})
	sort.Slice(audits, func(i, j int) bool { return audits[i].At.Before(audits[j].At) })
	return audits, err
}

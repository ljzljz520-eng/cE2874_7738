package store

import (
	"os"
	"sort"
	"time"

	"go.etcd.io/bbolt"
	"rehab-followup/internal/model"
)

func (s *Store) SaveReminder(reminder model.Reminder) error {
	if err := reminder.Validate(); err != nil {
		return err
	}
	s.mu.RLock()
	defer s.mu.RUnlock()
	if s.db == nil {
		return os.ErrClosed
	}
	return s.db.Update(func(tx *bbolt.Tx) error {
		return saveJSON(tx, remindersBucket, []byte(reminder.ID), reminder)
	})
}

func (s *Store) GetReminder(id string) (model.Reminder, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	if s.db == nil {
		return model.Reminder{}, os.ErrClosed
	}
	var reminder model.Reminder
	err := s.db.View(func(tx *bbolt.Tx) error {
		return loadJSON(tx, remindersBucket, []byte(id), &reminder)
	})
	return reminder, err
}

func (s *Store) ListReminders(patientID string) ([]model.Reminder, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	if s.db == nil {
		return nil, os.ErrClosed
	}
	reminders := make([]model.Reminder, 0)
	err := s.db.View(func(tx *bbolt.Tx) error {
		return listJSON[model.Reminder](tx, remindersBucket, func(data []byte) error {
			var reminder model.Reminder
			if err := decode(data, &reminder); err != nil {
				return err
			}
			if patientID == "" || reminder.PatientID == patientID {
				reminders = append(reminders, reminder)
			}
			return nil
		})
	})
	sort.Slice(reminders, func(i, j int) bool { return reminders[i].DueAt.Before(reminders[j].DueAt) })
	return reminders, err
}

func (s *Store) UpdateReminderStatus(id string, status model.ReminderStatus, now time.Time) error {
	reminder, err := s.GetReminder(id)
	if err != nil {
		return err
	}
	reminder.Status = status
	reminder.UpdatedAt = now
	return s.SaveReminder(reminder)
}

package reminder

import (
	"testing"
	"time"

	"rehab-followup/internal/model"
)

func TestReminderLifecycle(t *testing.T) {
	now := time.Date(2026, 8, 24, 9, 0, 0, 0, time.UTC)
	e := NewEngine(func() time.Time { return now })
	item, err := e.Create("p1", "review gait", now.Add(time.Hour))
	if err != nil {
		t.Fatal(err)
	}
	item = e.Process(item)
	item = e.Review(item, true)
	item = e.Publish(item)
	if item.Status != model.ReminderPublished || item.IsDue(now.Add(2*time.Hour)) {
		t.Fatal("reminder did not publish")
	}
}

func TestQueueSortsAndFilters(t *testing.T) {
	now := time.Now()
	e := NewEngine(func() time.Time { return now })
	items := []model.Reminder{{ID: "late", PatientID: "p", DueAt: now.Add(2 * time.Hour), Status: model.ReminderQueued}, {ID: "early", PatientID: "p", DueAt: now.Add(time.Hour), Status: model.ReminderProcessed}, {ID: "done", PatientID: "p", DueAt: now, Status: model.ReminderPublished}}
	queued := e.Queue(items)
	if len(queued) != 2 || queued[0].ID != "early" {
		t.Fatal("queue order mismatch")
	}
}

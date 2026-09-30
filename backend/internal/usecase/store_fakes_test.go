package usecase_test

import (
	"context"
	"time"

	"github.com/google/uuid"

	domainevent "github.com/boms/backend/internal/domain/event"
	domainstore "github.com/boms/backend/internal/domain/store"
	"github.com/boms/backend/internal/port"
	apperrors "github.com/boms/backend/internal/shared/errors"
)

var (
	_ port.StoreSettingsRepository = (*memoryStore)(nil)
	_ port.EventOutbox             = (*recordingOutbox)(nil)
)

// memoryStore keeps settings and closed dates the way the Postgres adapter
// answers for them: conflicts on a closed day, not-found on a missing one.
// A set err fails every read, as an unreachable database would.
type memoryStore struct {
	settings domainstore.Settings
	closed   map[uuid.UUID]domainstore.ClosedDate
	err      error
}

func newMemoryStore() *memoryStore {
	return &memoryStore{
		settings: domainstore.Settings{
			OpensAt: 8 * time.Hour, ClosesAt: 18 * time.Hour, PreorderMinLead: 2 * time.Hour, MaxAdvanceDays: 14,
			SlotLength: 30 * time.Minute, SlotCapacity: 10, InstantPrep: 20 * time.Minute, PaymentHold: 15 * time.Minute,
		},
		closed: map[uuid.UUID]domainstore.ClosedDate{},
	}
}

func (m *memoryStore) GetSettings(context.Context) (domainstore.Settings, error) {
	return m.settings, m.err
}

func (m *memoryStore) GetSettingsForUpdate(context.Context) (domainstore.Settings, error) {
	return m.settings, nil
}

func (m *memoryStore) UpdateSettings(_ context.Context, s domainstore.Settings) (domainstore.Settings, error) {
	s.UpdatedAt = time.Date(2026, 9, 28, 10, 0, 0, 0, time.UTC)
	m.settings = s
	return s, nil
}

func (m *memoryStore) ListClosedDates(_ context.Context, from, to time.Time) ([]domainstore.ClosedDate, error) {
	if m.err != nil {
		return nil, m.err
	}
	var out []domainstore.ClosedDate
	for _, c := range m.closed {
		if !c.Day.Before(from) && !c.Day.After(to) {
			out = append(out, c)
		}
	}
	return out, nil
}

func (m *memoryStore) AddClosedDate(_ context.Context, day time.Time, reason string) (domainstore.ClosedDate, error) {
	for _, c := range m.closed {
		if c.Day.Equal(day) {
			return domainstore.ClosedDate{}, apperrors.ErrConflict
		}
	}
	c := domainstore.ClosedDate{ID: uuid.New(), Day: day, Reason: reason}
	m.closed[c.ID] = c
	return c, nil
}

func (m *memoryStore) RemoveClosedDate(_ context.Context, id uuid.UUID) (domainstore.ClosedDate, error) {
	c, ok := m.closed[id]
	if !ok {
		return domainstore.ClosedDate{}, apperrors.ErrNotFound
	}
	delete(m.closed, id)
	return c, nil
}

type recordingOutbox struct {
	err    error
	events []domainevent.Event
}

func (r *recordingOutbox) Add(_ context.Context, e domainevent.Event) error {
	if r.err != nil {
		return r.err
	}
	r.events = append(r.events, e)
	return nil
}

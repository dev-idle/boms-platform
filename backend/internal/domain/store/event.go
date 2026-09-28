package store

import domainevent "github.com/boms/backend/internal/domain/event"

// TopicSettingsUpdated announces a change to the pickup rules: hours, lead
// time, booking window or closed dates. Every open page may be showing them,
// and they are public, so every socket hears it.
const TopicSettingsUpdated domainevent.Topic = "settings.updated"

// SettingsUpdatedEvent tells every open page to read the pickup rules again.
func SettingsUpdatedEvent() domainevent.Event {
	return domainevent.New(TopicSettingsUpdated, domainevent.Audience{Public: true}, map[string]string{})
}

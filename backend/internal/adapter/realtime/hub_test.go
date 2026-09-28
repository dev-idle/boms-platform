package realtime

import (
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestHub(t *testing.T) {
	t.Parallel()

	t.Run("delivers_only_to_sockets_on_the_channel", func(t *testing.T) {
		t.Parallel()
		hub := NewHub(5)
		staff := newClient(uuid.New(), []string{"user:a", "role:staff"}, 4)
		customer := newClient(uuid.New(), []string{"user:b"}, 4)
		require.True(t, hub.add(staff))
		require.True(t, hub.add(customer))

		hub.Deliver("role:staff", []byte("event"))

		assert.Equal(t, []byte("event"), <-staff.send)
		assert.Empty(t, customer.send)
	})

	t.Run("caps_sockets_per_user", func(t *testing.T) {
		t.Parallel()
		hub := NewHub(2)
		user := uuid.New()
		first := newClient(user, []string{"user:a"}, 1)
		require.True(t, hub.add(first))
		require.True(t, hub.add(newClient(user, []string{"user:a"}, 1)))

		assert.False(t, hub.add(newClient(user, []string{"user:a"}, 1)))
		assert.True(t, hub.add(newClient(uuid.New(), []string{"user:b"}, 1)), "another user is unaffected")

		hub.remove(first)
		assert.True(t, hub.add(newClient(user, []string{"user:a"}, 1)), "a closed socket frees its slot")
	})

	t.Run("forgets_removed_sockets", func(t *testing.T) {
		t.Parallel()
		hub := NewHub(1)
		c := newClient(uuid.New(), []string{"user:a", "role:baker"}, 1)
		require.True(t, hub.add(c))

		hub.remove(c)

		assert.Empty(t, hub.byChannel)
		assert.Empty(t, hub.perUser)
	})

	t.Run("marks_a_full_socket_slow_instead_of_blocking", func(t *testing.T) {
		t.Parallel()
		hub := NewHub(1)
		c := newClient(uuid.New(), []string{"user:a"}, 1)
		require.True(t, hub.add(c))

		hub.Deliver("user:a", []byte("first"))
		hub.Deliver("user:a", []byte("second"))
		hub.Deliver("user:a", []byte("third"))

		assert.Len(t, c.send, 1)
		select {
		case <-c.slow:
		default:
			t.Fatal("a socket that fell behind must be marked slow")
		}
	})
}

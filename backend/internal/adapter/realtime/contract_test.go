package realtime

import (
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"testing"

	"github.com/stretchr/testify/require"

	domainevent "github.com/boms/backend/internal/domain/event"
	domainorder "github.com/boms/backend/internal/domain/order"
)

// The browser acts on two things this package and the domain define: the event
// topics, which choose what a page refetches, and the close code that asks for a
// new ticket. A topic renamed on one side only would leave every page "live"
// and never updating, so this test fails the backend build when they drift.

var (
	frontendTopicsBlock = regexp.MustCompile(`(?s)export const REALTIME_EVENT_TYPE = \{(.*?)\} as const`)
	frontendTopic       = regexp.MustCompile(`:\s*"([a-z_.]+)"`)
	frontendReauthCode  = regexp.MustCompile(`const CLOSE_REAUTHENTICATE = (\d+);`)
)

// publishedTopics lists every topic the backend raises; a new one belongs here.
var publishedTopics = []domainevent.Topic{
	domainorder.TopicOrderCreated,
	domainorder.TopicOrderStatusChanged,
}

func TestRealtimeContractMatchesTheFrontend(t *testing.T) {
	t.Parallel()
	// A rooted file system: the reads below cannot leave the frontend.
	frontend := os.DirFS(filepath.Join("..", "..", "..", "..", "frontend", "src", "lib", "realtime"))

	t.Run("every_topic_is_known_to_the_browser", func(t *testing.T) {
		t.Parallel()
		src, err := fs.ReadFile(frontend, "events.ts")
		require.NoError(t, err)
		block := frontendTopicsBlock.FindSubmatch(src)
		require.NotNil(t, block, "REALTIME_EVENT_TYPE not found in events.ts")
		known := map[string]bool{}
		for _, m := range frontendTopic.FindAllSubmatch(block[1], -1) {
			known[string(m[1])] = true
		}
		for _, topic := range publishedTopics {
			require.True(t, known[string(topic)], "add %q to REALTIME_EVENT_TYPE in frontend/src/lib/realtime/events.ts", topic)
		}
	})

	t.Run("the_reauthenticate_code_matches", func(t *testing.T) {
		t.Parallel()
		src, err := fs.ReadFile(frontend, "connection.ts")
		require.NoError(t, err)
		m := frontendReauthCode.FindSubmatch(src)
		require.NotNil(t, m, "CLOSE_REAUTHENTICATE not found in connection.ts")
		code, err := strconv.Atoi(string(m[1]))
		require.NoError(t, err)
		require.Equal(t, int(StatusReauthenticate), code)
	})
}

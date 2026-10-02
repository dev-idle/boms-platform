package usecase_test

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	domainorder "github.com/boms/backend/internal/domain/order"
	domainstore "github.com/boms/backend/internal/domain/store"
	"github.com/boms/backend/internal/port"
	"github.com/boms/backend/internal/usecase"
)

// weekIncidents serves one incident and records the week and type it was asked
// for; any other call panics.
type weekIncidents struct {
	port.OrderIncidentRepository
	listed  port.ListOrderIncidentsParams
	counted []time.Time
}

func (f *weekIncidents) ListIncidents(_ context.Context, params port.ListOrderIncidentsParams) ([]port.ManagerOrderIncident, error) {
	f.listed = params
	name := "Lan"
	return []port.ManagerOrderIncident{{
		Incident:  domainorder.Incident{ID: uuid.New(), Type: domainorder.IncidentWrongItems, CreatedAt: params.From},
		OrderCode: "C-1002-001",
		ActorName: &name,
	}}, nil
}

func (f *weekIncidents) CountIncidents(context.Context, time.Time, time.Time, *domainorder.IncidentType) (int64, error) {
	return 1, nil
}

func (f *weekIncidents) CountIncidentsByType(_ context.Context, from, to time.Time) ([]port.IncidentTypeCount, error) {
	f.counted = []time.Time{from, to}
	return []port.IncidentTypeCount{{Type: domainorder.IncidentNoShow, Count: 3}}, nil
}

func TestManagerIncidentUsecase(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	monday := time.Date(2026, 9, 28, 0, 0, 0, 0, domainstore.Location)

	t.Run("lists_a_bakery_week_of_one_type", func(t *testing.T) {
		t.Parallel()
		incidents := &weekIncidents{}

		out, total, page, pageSize, err := usecase.NewManagerIncidentUsecase(incidents).List(ctx, "2026-09-28", "wrong_items", 2, 500)

		require.NoError(t, err)
		assert.Equal(t, monday, incidents.listed.From, "from midnight at the bakery")
		assert.Equal(t, monday.AddDate(0, 0, 7), incidents.listed.To)
		require.NotNil(t, incidents.listed.Type)
		assert.Equal(t, domainorder.IncidentWrongItems, *incidents.listed.Type)
		assert.Equal(t, int32(100), pageSize, "capped")
		assert.Equal(t, int32(100), incidents.listed.Offset)
		assert.Equal(t, int64(1), total)
		assert.Equal(t, int32(2), page)
		require.Len(t, out, 1)
		assert.Equal(t, "manual", out[0].Source)
		assert.Equal(t, "C-1002-001", out[0].OrderCode)
	})

	t.Run("refuses_a_week_that_does_not_start_on_monday", func(t *testing.T) {
		t.Parallel()
		uc := usecase.NewManagerIncidentUsecase(&weekIncidents{})
		for _, week := range []string{"2026-09-29", "28/09/2026", ""} {
			_, _, _, _, err := uc.List(ctx, week, "", 1, 20)
			requireValidationField(t, err, "week")
			_, err = uc.Summary(ctx, week)
			requireValidationField(t, err, "week")
		}
	})

	t.Run("refuses_an_unknown_type", func(t *testing.T) {
		t.Parallel()
		_, _, _, _, err := usecase.NewManagerIncidentUsecase(&weekIncidents{}).List(ctx, "2026-09-28", "late", 1, 20)
		requireValidationField(t, err, "type")
	})

	t.Run("counts_a_week_by_type", func(t *testing.T) {
		t.Parallel()
		incidents := &weekIncidents{}

		out, err := usecase.NewManagerIncidentUsecase(incidents).Summary(ctx, "2026-09-28")

		require.NoError(t, err)
		assert.Equal(t, []time.Time{monday, monday.AddDate(0, 0, 7)}, incidents.counted)
		require.Len(t, out.Types, 1)
		assert.Equal(t, "no_show", out.Types[0].Type)
		assert.Equal(t, int64(3), out.Types[0].Count)
	})
}

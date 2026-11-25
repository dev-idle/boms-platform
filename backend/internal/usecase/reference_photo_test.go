package usecase_test

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/boms/backend/internal/config"
	"github.com/boms/backend/internal/usecase"
)

// pendingErasures holds the erasures still to do and those done; listErr and
// markErr fail the list and the mark.
type pendingErasures struct {
	due     []uuid.UUID
	done    []uuid.UUID
	listErr error
	markErr error
}

func (p *pendingErasures) RequestErasure(_ context.Context, userID uuid.UUID) error {
	p.due = append(p.due, userID)
	return nil
}

func (p *pendingErasures) ListErasuresDue(context.Context, int32) ([]uuid.UUID, error) {
	return p.due, p.listErr
}

func (p *pendingErasures) MarkErased(_ context.Context, userID uuid.UUID) error {
	if p.markErr != nil {
		return p.markErr
	}
	p.done = append(p.done, userID)
	return nil
}

// errRateLimited is Cloudinary refusing a call.
var errRateLimited = errors.New("cloudinary answered 420 rate limited")

// folderStore deletes folders, refusing those in refuse.
type folderStore struct {
	deleted []string
	refuse  map[string]bool
}

func (s *folderStore) DeleteFolder(_ context.Context, folder string) error {
	if s.refuse[folder] {
		return errRateLimited
	}
	s.deleted = append(s.deleted, folder)
	return nil
}

func TestReferencePhotoUsecase_EraseDue(t *testing.T) {
	t.Parallel()
	cloud := config.CloudinaryConfig{ReferenceFolder: "boms/references"}
	refused, cleared := uuid.New(), uuid.New()
	erasures := &pendingErasures{due: []uuid.UUID{refused, cleared}}
	images := &folderStore{refuse: map[string]bool{cloud.CustomerReferenceFolder(refused): true}}

	erased, err := usecase.NewReferencePhotoUsecase(erasures, images, cloud).EraseDue(context.Background())

	require.ErrorIs(t, err, errRateLimited, "the refusal is reported")
	assert.Equal(t, 1, erased, "the others go on")
	assert.Equal(t, []string{"boms/references/" + cleared.String()}, images.deleted, "each customer's own folder")
	assert.Equal(t, []uuid.UUID{cleared}, erasures.done, "the refused one stays due for the next run")
}

func TestReferencePhotoUsecase_EraseDue_Failures(t *testing.T) {
	t.Parallel()
	cloud := config.CloudinaryConfig{ReferenceFolder: "boms/references"}
	down := errors.New("database unavailable")

	t.Run("a_mark_that_fails_is_reported_and_tried_again", func(t *testing.T) {
		t.Parallel()
		erasures := &pendingErasures{due: []uuid.UUID{uuid.New()}, markErr: down}
		images := &folderStore{}

		erased, err := usecase.NewReferencePhotoUsecase(erasures, images, cloud).EraseDue(context.Background())

		require.ErrorIs(t, err, down)
		assert.Zero(t, erased, "not counted: the next run deletes the folder again, which finds it empty")
		assert.Len(t, images.deleted, 1)
	})

	t.Run("a_list_that_fails_deletes_nothing", func(t *testing.T) {
		t.Parallel()
		images := &folderStore{}

		_, err := usecase.NewReferencePhotoUsecase(&pendingErasures{listErr: down}, images, cloud).EraseDue(context.Background())

		require.ErrorIs(t, err, down)
		assert.Empty(t, images.deleted)
	})
}

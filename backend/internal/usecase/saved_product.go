package usecase

import (
	"context"
	"errors"

	"github.com/google/uuid"

	domainproduct "github.com/boms/backend/internal/domain/product"
	domainsaved "github.com/boms/backend/internal/domain/saved"
	"github.com/boms/backend/internal/dto"
	"github.com/boms/backend/internal/port"
	apperrors "github.com/boms/backend/internal/shared/errors"
)

// SavedProductUsecase is a customer keeping products on their favorites and
// their wishlist.
type SavedProductUsecase struct {
	users port.UserRepository
	saved port.SavedProductRepository
	tx    port.TxManager
}

func NewSavedProductUsecase(users port.UserRepository, saved port.SavedProductRepository, tx port.TxManager) *SavedProductUsecase {
	return &SavedProductUsecase{users: users, saved: saved, tx: tx}
}

// List returns the products on sale on both of the customer's lists, latest
// first; one the bakery stopped selling is back once it is on sale again.
func (u *SavedProductUsecase) List(ctx context.Context, userID uuid.UUID) ([]dto.SavedProductResponse, error) {
	rows, err := u.saved.ListByUser(ctx, userID)
	if err != nil {
		return nil, err
	}
	out := make([]dto.SavedProductResponse, 0, len(rows))
	for _, row := range rows {
		out = append(out, dto.SavedProductResponse{
			List:    string(row.List),
			SavedAt: row.SavedAt,
			Product: toCatalogProductResponse(row.Product, row.Product.ImageURLs),
		})
	}
	return out, nil
}

// Save puts a product on sale on one of the customer's lists, or keeps it
// there. The account is held while it is written, so two saves at once cannot
// pass the list's limit, and one that waited on an erasure finds no account.
func (u *SavedProductUsecase) Save(ctx context.Context, userID uuid.UUID, list string, productID uuid.UUID) error {
	savedList, err := parseSavedList(list)
	if err != nil {
		return err
	}
	return u.tx.WithTx(ctx, func(txCtx context.Context) error {
		if _, err := u.users.GetByIDForUpdate(txCtx, userID); err != nil {
			if errors.Is(err, apperrors.ErrNotFound) {
				return ErrMeNotFound
			}
			return err
		}
		others, err := u.saved.CountOthers(txCtx, userID, productID, savedList)
		if err != nil {
			return err
		}
		if others >= domainsaved.MaxPerList {
			return domainsaved.ErrListFull
		}
		saved, err := u.saved.Save(txCtx, userID, productID, savedList)
		if err != nil {
			return err
		}
		if !saved {
			return domainproduct.ErrNotFound
		}
		return nil
	})
}

// Remove takes a product off one of the customer's lists; one not on it stays off.
func (u *SavedProductUsecase) Remove(ctx context.Context, userID uuid.UUID, list string, productID uuid.UUID) error {
	savedList, err := parseSavedList(list)
	if err != nil {
		return err
	}
	return u.saved.Remove(ctx, userID, productID, savedList)
}

func parseSavedList(list string) (domainsaved.List, error) {
	savedList := domainsaved.List(list)
	if !savedList.Valid() {
		return "", domainsaved.ErrInvalidList
	}
	return savedList, nil
}

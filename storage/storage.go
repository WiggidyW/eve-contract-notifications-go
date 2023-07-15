package storage

import (
	"context"

	"github.com/WiggidyW/eve-contract-notifications-go/hashcode"
)

type Storage interface {
	// Get returns the stored list of hash codes.
	Get(ctx context.Context) (hashcode.HashCodeSet, error)
	// Set stores the list of hash codes.
	Set(ctx context.Context, hashCodes *[]hashcode.HashCode) error
}

func NewStorage(ctx context.Context) (Storage, error) {
	return NewFirestoreStorage(ctx)
}

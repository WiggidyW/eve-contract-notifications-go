package contracts

import (
	"context"
	"time"

	"github.com/WiggidyW/eve-contract-notifications-go/hashcode"
)

type Contract struct {
	HashCode hashcode.HashCode
	Issued   time.Time
	Expires  time.Time
}

type ContractGetter interface {
	// Get returns the list of contracts.
	Get(ctx context.Context) (*[]Contract, error)
}

func NewContractGetter(
	ctx context.Context,
) (ContractGetter, error) {
	return NewItemConfiguratorClient(ctx)
}

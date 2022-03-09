package types

import (
	"github.com/cosmos/ibc-go/modules/core/exported"
)

type MultiPrefix struct {
	exported.Prefix
	PathPrefix []byte
}

var _ exported.Prefix = (*MultiPrefix)(nil)

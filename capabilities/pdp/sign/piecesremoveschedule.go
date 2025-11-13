package sign

import (
	"math/big"

	"github.com/storacha/go-ucanto/core/ipld"
	"github.com/storacha/go-ucanto/core/schema"
	"github.com/storacha/go-ucanto/validator"
)

const PiecesRemoveScheduleAbility = "pdp/sign/pieces/remove/schedule"

type PiecesRemoveScheduleCaveats struct {
	DataSet *big.Int
	Pieces  []*big.Int
}

func (c PiecesRemoveScheduleCaveats) ToIPLD() (ipld.Node, error) {
	return ipld.WrapWithRecovery(&c, PiecesRemoveScheduleCaveatsType(), Converters...)
}

var PiecesRemoveScheduleCaveatsReader = schema.Struct[PiecesRemoveScheduleCaveats](PiecesRemoveScheduleCaveatsType(), nil, Converters...)

type PiecesRemoveScheduleOk = AuthSignature

var PiecesRemoveScheduleOkReader = AuthSignatureReader

var PiecesRemoveSchedule = validator.NewCapability(
	PiecesRemoveScheduleAbility,
	schema.DIDString(),
	PiecesRemoveScheduleCaveatsReader,
	validator.DefaultDerives,
)

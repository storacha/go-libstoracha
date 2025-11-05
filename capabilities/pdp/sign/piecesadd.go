package sign

import (
	"math/big"

	"github.com/storacha/go-libstoracha/capabilities/types"
	"github.com/storacha/go-ucanto/core/ipld"
	"github.com/storacha/go-ucanto/core/schema"
	"github.com/storacha/go-ucanto/validator"
)

const PiecesAddAbility = "pdp/sign/pieces/add"

type PiecesAddCaveats struct {
	DataSet    *big.Int
	FirstAdded *big.Int
	PieceData  [][]byte
	Metadata   Metadata
}

func (c PiecesAddCaveats) ToIPLD() (ipld.Node, error) {
	return ipld.WrapWithRecovery(&c, PiecesAddCaveatsType(), types.Converters...)
}

var PiecesAddCaveatsReader = schema.Struct[PiecesAddCaveats](PiecesAddCaveatsType(), nil, types.Converters...)

type PiecesAddOk = AuthSignature

var PiecesAddOkReader = AuthSignatureReader

var PiecesAdd = validator.NewCapability(
	PiecesAddAbility,
	schema.DIDString(),
	PiecesAddCaveatsReader,
	validator.DefaultDerives,
)

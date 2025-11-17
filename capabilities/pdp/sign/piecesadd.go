package sign

import (
	"math/big"

	"github.com/storacha/go-ucanto/core/ipld"
	"github.com/storacha/go-ucanto/core/schema"
	"github.com/storacha/go-ucanto/validator"
)

const PiecesAddAbility = "pdp/sign/pieces/add"

type PiecesAddCaveats struct {
	DataSet   *big.Int
	Nonce     *big.Int
	PieceData [][]byte
	Metadata  []Metadata
	// Proofs are links to `blob/accept` receipts for sub-pieces included in each
	// piece. They are proofs that the sub-pieces were requested to be stored by
	// the node. They correspond to items in `PieceData` i.e. Proofs[0] is the
	// list of receipts for all sub-pieces of PieceData[0].
	//
	// Each `blob/accept` receipt MUST include the `pdp/accept` effect receipt,
	// since it contains the proof that the sub-piece is included in the larger
	// piece. All receipt data MUST be attached to the signing invocation.
	Proofs [][]ipld.Link
}

func (c PiecesAddCaveats) ToIPLD() (ipld.Node, error) {
	return ipld.WrapWithRecovery(&c, PiecesAddCaveatsType(), Converters...)
}

var PiecesAddCaveatsReader = schema.Struct[PiecesAddCaveats](PiecesAddCaveatsType(), nil, Converters...)

type PiecesAddOk = AuthSignature

var PiecesAddOkReader = AuthSignatureReader

var PiecesAdd = validator.NewCapability(
	PiecesAddAbility,
	schema.DIDString(),
	PiecesAddCaveatsReader,
	validator.DefaultDerives,
)

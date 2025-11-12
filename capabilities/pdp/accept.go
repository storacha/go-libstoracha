package pdp

import (
	"github.com/filecoin-project/go-data-segment/merkletree"
	"github.com/ipld/go-ipld-prime/datamodel"
	mh "github.com/multiformats/go-multihash"
	"github.com/storacha/go-libstoracha/capabilities/types"
	"github.com/storacha/go-libstoracha/piece/piece"
	"github.com/storacha/go-ucanto/core/ipld"
	"github.com/storacha/go-ucanto/core/receipt"
	"github.com/storacha/go-ucanto/core/result/failure"
	"github.com/storacha/go-ucanto/core/schema"
	"github.com/storacha/go-ucanto/validator"
)

const AcceptAbility = "pdp/accept"

type AcceptCaveats struct {
	Blob mh.Multihash
}

func (ac AcceptCaveats) ToIPLD() (datamodel.Node, error) {
	return ipld.WrapWithRecovery(&ac, AcceptCaveatsType(), types.Converters...)
}

var AcceptCaveatsReader = schema.Struct[AcceptCaveats](AcceptCaveatsType(), nil, types.Converters...)

var Accept = validator.NewCapability(
	AcceptAbility,
	schema.DIDString(),
	AcceptCaveatsReader,
	validator.DefaultDerives,
)

type AcceptOk struct {
	Aggregate      piece.PieceLink
	InclusionProof merkletree.ProofData
	Piece          piece.PieceLink
}

func (ao AcceptOk) ToIPLD() (datamodel.Node, error) {
	return ipld.WrapWithRecovery(&ao, AcceptOkType(), types.Converters...)
}

var AcceptOkReader = schema.Struct[AcceptOk](AcceptOkType(), nil, types.Converters...)

type AcceptReceipt receipt.Receipt[AcceptOk, failure.Failure]
type AcceptReceiptReader receipt.ReceiptReader[AcceptOk, failure.Failure]

func NewAcceptReceiptReader() (AcceptReceiptReader, error) {
	return receipt.NewReceiptReader[AcceptOk, failure.Failure](pdpSchema)
}

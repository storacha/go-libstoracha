package filecoin

import (
	"github.com/ipld/go-ipld-prime/datamodel"
	"github.com/storacha/go-libstoracha/capabilities/types"
	"github.com/storacha/go-ucanto/core/ipld"
	"github.com/storacha/go-ucanto/core/result/failure"
	"github.com/storacha/go-ucanto/core/schema"
	"github.com/storacha/go-ucanto/ucan"
	"github.com/storacha/go-ucanto/validator"
)

const (
	AcceptAbility = "filecoin/accept"
)

type AcceptCaveats struct {
	Content datamodel.Link `ipld:"content"`
	Piece   datamodel.Link `ipld:"piece"`
}

func (ac AcceptCaveats) ToIPLD() (datamodel.Node, error) {
	return ipld.WrapWithRecovery(&ac, AcceptCaveatsType(), types.Converters...)
}

var AcceptCaveatsReader = schema.Struct[AcceptCaveats](AcceptCaveatsType(), nil, types.Converters...)


type AcceptOk struct {
	Piece     datamodel.Link `ipld:"piece"`
	Aggregate datamodel.Link `ipld:"aggregate"`
	Inclusion InclusionProof `ipld:"inclusion"`
	Aux       DealMetadata   `ipld:"aux"`
}

func (ao AcceptOk) ToIPLD() (datamodel.Node, error) {
	return ipld.WrapWithRecovery(&ao, AcceptOkType(), types.Converters...)
}

var AcceptOkReader = schema.Struct[AcceptOk](AcceptOkType(), nil, types.Converters...)

// Accept is a capability allowing a Storefront to signal that a submitted piece 
// has been accepted in a Filecoin deal. The receipt contains the proof.
var Accept = validator.NewCapability(
	AcceptAbility,
	schema.DIDString(),
	AcceptCaveatsReader,
	func(claimed, delegated ucan.Capability[AcceptCaveats]) failure.Failure {
		return validateFilecoinCapability(claimed, delegated, func(claimedNb, delegatedNb AcceptCaveats) failure.Failure {
			if fail := equalLink(claimedNb.Content, delegatedNb.Content, "content"); fail != nil {
				return fail
			}
						return equalPieceLink(claimedNb.Piece, delegatedNb.Piece)
		})
	},
)
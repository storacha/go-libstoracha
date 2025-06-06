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

const SubmitAbility = "filecoin/submit"

type SubmitCaveats struct {
	Content ipld.Link
	Piece   ipld.Link
}

func (sc SubmitCaveats) ToIPLD() (datamodel.Node, error) {
	return ipld.WrapWithRecovery(&sc, SubmitCaveatsType(), types.Converters...)
}

var SubmitCaveatsReader = schema.Struct[SubmitCaveats](SubmitCaveatsType(), nil, types.Converters...)

type SubmitOk struct {
	Piece ipld.Link
}

func (so SubmitOk) ToIPLD() (datamodel.Node, error) {
	return ipld.WrapWithRecovery(&so, SubmitOkType(), types.Converters...)
}

var SubmitOkReader = schema.Struct[SubmitOk](SubmitOkType(), nil, types.Converters...)

// Submit is a capability allowing a Storefront to signal that an offered piece
// has been submitted to the filecoin storage pipeline.
var Submit = validator.NewCapability(
	SubmitAbility,
	schema.DIDString(),
	SubmitCaveatsReader,
	func(claimed, delegated ucan.Capability[SubmitCaveats]) failure.Failure {
		return validateFilecoinCapability(claimed, delegated, func(claimedNb, delegatedNb SubmitCaveats) failure.Failure {
			if fail := equalLink(claimedNb.Content, delegatedNb.Content, "content"); fail != nil {
				return fail
			}
			return equalPieceLink(claimedNb.Piece, delegatedNb.Piece)
		})
	},
)

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

const OfferAbility = "filecoin/offer"

type OfferCaveats struct {
	Content datamodel.Link
	Piece   datamodel.Link
}

func (oc OfferCaveats) ToIPLD() (datamodel.Node, error) {
	return ipld.WrapWithRecovery(&oc, OfferCaveatsType(), types.Converters...)
}

var OfferCaveatsReader = schema.Struct[OfferCaveats](OfferCaveatsType(), nil, types.Converters...)

type OfferOk struct {
	Piece datamodel.Link
}

func (oo OfferOk) ToIPLD() (datamodel.Node, error) {
	return ipld.WrapWithRecovery(&oo, OfferOkType(), types.Converters...)
}

var OfferOkReader = schema.Struct[OfferOk](OfferOkType(), nil, types.Converters...)

var Offer = validator.NewCapability(
	OfferAbility,
	schema.DIDString(),
	OfferCaveatsReader,
	func(claimed, delegated ucan.Capability[OfferCaveats]) failure.Failure {
		return validateFilecoinCapability(claimed, delegated, func(claimedNb, delegatedNb OfferCaveats) failure.Failure {
			if fail := equalLink(claimedNb.Content, delegatedNb.Content, "content"); fail != nil {
				return fail
			}

			return equalPieceLink(claimedNb.Piece, delegatedNb.Piece)
		})
	},
)

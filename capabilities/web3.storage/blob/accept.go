package blob

import (
	"github.com/ipld/go-ipld-prime/datamodel"
	"github.com/storacha/go-libstoracha/capabilities/types"
	"github.com/storacha/go-ucanto/core/ipld"
	"github.com/storacha/go-ucanto/core/result/failure"
	"github.com/storacha/go-ucanto/core/schema"
	"github.com/storacha/go-ucanto/did"
	"github.com/storacha/go-ucanto/ucan"
	"github.com/storacha/go-ucanto/validator"
)

const AcceptAbility = "web3.storage/blob/accept"

type AcceptCaveats struct {
	Space did.DID
	Blob  types.Blob
	TTL   int
	Put   types.Promise
}

func (ac AcceptCaveats) ToIPLD() (datamodel.Node, error) {
	return ipld.WrapWithRecovery(&ac, AcceptCaveatsType(), types.Converters...)
}

var AcceptCaveatsReader = schema.Struct[AcceptCaveats](AcceptCaveatsType(), nil, types.Converters...)

type AcceptOk struct {
	Site ucan.Link
}

func (ao AcceptOk) ToIPLD() (datamodel.Node, error) {
	return ipld.WrapWithRecovery(&ao, AcceptOkType(), types.Converters...)
}

var AcceptOkReader = schema.Struct[AcceptOk](AcceptOkType(), nil, types.Converters...)

var Accept = validator.NewCapability(
	AcceptAbility,
	schema.DIDString(),
	AcceptCaveatsReader,
	func(claimed, delegated ucan.Capability[AcceptCaveats]) failure.Failure {
		fail := equalWith(claimed.With(), delegated.With())
		if fail != nil {
			return fail
		}

		fail = equalBlob(claimed.Nb().Blob, delegated.Nb().Blob)
		if fail != nil {
			return fail
		}

		fail = equalTTL(claimed.Nb().TTL, delegated.Nb().TTL)
		if fail != nil {
			return fail
		}

		return checkSpace(claimed.Nb().Space.String(), delegated.Nb().Space.String())
	},
)

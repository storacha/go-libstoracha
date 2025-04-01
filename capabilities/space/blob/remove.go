package blob

import (
	"bytes"
	"fmt"

	"github.com/ipld/go-ipld-prime/datamodel"
	"github.com/multiformats/go-multibase"
	"github.com/multiformats/go-multihash"
	"github.com/storacha/go-libstoracha/capabilities/types"
	"github.com/storacha/go-ucanto/core/ipld"
	"github.com/storacha/go-ucanto/core/result/failure"
	"github.com/storacha/go-ucanto/core/schema"
	"github.com/storacha/go-ucanto/ucan"
	"github.com/storacha/go-ucanto/validator"
)

const RemoveAbility = "space/blob/remove"

// RemoveCaveats represents the caveats required to perform a space/blob/remove invocation.
type RemoveCaveats struct {
	// Digest is the multihash of the blob to be removed.
	Digest multihash.Multihash
}

func (rc RemoveCaveats) ToIPLD() (datamodel.Node, error) {
	return ipld.WrapWithRecovery(&rc, RemoveCaveatsType(), types.Converters...)
}

var RemoveCaveatsReader = schema.Struct[RemoveCaveats](RemoveCaveatsType(), nil, types.Converters...)

// RemoveOk represents the successful response for a space/blob/remove invocation.
type RemoveOk struct {
	// Size is the size of the blob that was removed in bytes.
	Size uint64
}

func (ro RemoveOk) ToIPLD() (datamodel.Node, error) {
	return ipld.WrapWithRecovery(&ro, RemoveOkType(), types.Converters...)
}

var RemoveOkReader = schema.Struct[RemoveOk](RemoveOkType(), nil, types.Converters...)

// Remove is a capability that allows the agent to remove a Blob from a space identified by did:key in the `with` field.
var Remove = validator.NewCapability(
	RemoveAbility,
	schema.DIDString(),
	RemoveCaveatsReader,
	func(claimed, delegated ucan.Capability[RemoveCaveats]) failure.Failure {
		// Check if the space matches
		if claimed.With() != delegated.With() {
			return schema.NewSchemaError(fmt.Sprintf(
				"Resource '%s' doesn't match delegated '%s'",
				claimed.With(), delegated.With(),
			))
		}

		// Check if the blob digest matches
		if delegated.Nb().Digest != nil {
			if !bytes.Equal(delegated.Nb().Digest, claimed.Nb().Digest) {
				claimedDigest, _ := multibase.Encode(multibase.Base58BTC, claimed.Nb().Digest)
				delegatedDigest, _ := multibase.Encode(multibase.Base58BTC, delegated.Nb().Digest)
				return schema.NewSchemaError(fmt.Sprintf(
					"Claimed digest '%s' doesn't match delegated '%s'",
					claimedDigest, delegatedDigest,
				))
			}
		}

		return nil
	},
)

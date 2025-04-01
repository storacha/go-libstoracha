package blob

import (
	"bytes"
	"fmt"
	"time"

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

const GetAbility = "space/blob/get/0/1"

// GetCaveats represents the caveats required to perform a space/blob/get/0/1 invocation.
type GetCaveats struct {
	// Digest is the multihash of the blob to be retrieved.
	Digest multihash.Multihash
}

func (gc GetCaveats) ToIPLD() (datamodel.Node, error) {
	return ipld.WrapWithRecovery(&gc, GetCaveatsType(), types.Converters...)
}

var GetCaveatsReader = schema.Struct[GetCaveats](GetCaveatsType(), nil, types.Converters...)

// GetOk represents the successful response for a space/blob/get/0/1 invocation.
type GetOk struct {
	// Blob is the retrieved blob.
	Blob Blob

	// Cause is a link to the task that caused the blob to be added.
	Cause ucan.Link

	// InsertedAt is the time when the blob was inserted.
	InsertedAt time.Time
}

func (ok GetOk) ToIPLD() (datamodel.Node, error) {
	return ipld.WrapWithRecovery(&ok, GetOkType(), types.Converters...)
}

var GetOkReader = schema.Struct[GetOk](GetOkType(), nil, types.Converters...)

// Get is a capability that allows the agent to retrieve a Blob from a space identified by did:key in the `with` field.
var Get = validator.NewCapability(
	GetAbility,
	schema.DIDString(),
	GetCaveatsReader,
	func(claimed, delegated ucan.Capability[GetCaveats]) failure.Failure {
		// Check if the space matches
		if claimed.With() != delegated.With() {
			return schema.NewSchemaError(fmt.Sprintf(
				"Expected 'with: %s' instead got '%s'",
				delegated.With(), claimed.With(),
			))
		}

		// Check if the blob digest matches
		if delegated.Nb().Digest != nil {
			if !bytes.Equal(delegated.Nb().Digest, claimed.Nb().Digest) {
				claimedDigest, _ := multibase.Encode(multibase.Base58BTC, claimed.Nb().Digest)
				delegatedDigest, _ := multibase.Encode(multibase.Base58BTC, delegated.Nb().Digest)
				return schema.NewSchemaError(fmt.Sprintf(
					"Link %s violates imposed %s constraint",
					claimedDigest, delegatedDigest,
				))
			}
		}

		return nil
	},
)

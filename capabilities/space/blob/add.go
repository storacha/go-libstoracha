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

const AddAbility = "space/blob/add"

// AddCaveats represents the caveats required to perform a space/blob/add invocation.
type AddCaveats struct {
	// Blob is the blob to be stored.
	Blob Blob
}

// Blob represents a blob to be stored.
type Blob struct {
	// Digest is the multihash of the blob payload bytes, uniquely identifying the blob.
	Digest multihash.Multihash

	// Size is the number of bytes in the blob. The service will provision a write target for this exact size.
	// Attempts to write a larger blob will fail.
	Size uint64
}

func (ac AddCaveats) ToIPLD() (datamodel.Node, error) {
	return ipld.WrapWithRecovery(&ac, AddCaveatsType(), types.Converters...)
}

// AddOk represents the result of a successful space/blob/add invocation.
type AddOk struct {
	// Receipt is a link to the receipt of the space/blob/add task.
	Receipt ipld.Link
}

func (ao AddOk) ToIPLD() (datamodel.Node, error) {
	return ipld.WrapWithRecovery(&ao, AddOkType(), types.Converters...)
}

// Add is a capability that allows the agent to store a Blob into a space identified by did:key in the `with` field.
// The agent should compute the blob's multihash and size and provide it under the `nb.blob` field, allowing a
// service to provision a write location for the agent to PUT the desired blob into.
var AddCaveatsReader = schema.Struct[AddCaveats](AddCaveatsType(), nil, types.Converters...)
var Add = validator.NewCapability(
	AddAbility,
	schema.DIDString(),
	AddCaveatsReader,
	func(claimed, delegated ucan.Capability[AddCaveats]) failure.Failure {
		fail := validator.DefaultDerives(claimed, delegated)
		if fail != nil {
			return fail
		}

		return equalBlob(claimed, delegated)
	},
)

// equalBlob validates that the claimed blob capability matches the delegated one.
func equalBlob(claimed, delegated ucan.Capability[AddCaveats]) failure.Failure {
	claimedBlob := claimed.Nb().Blob
	delegatedBlob := delegated.Nb().Blob

	// Check if the blob digest matches
	if delegatedBlob.Digest != nil {
		if !bytes.Equal(delegatedBlob.Digest, claimedBlob.Digest) {
			claimedDigest, _ := multibase.Encode(multibase.Base58BTC, claimedBlob.Digest)
			delegatedDigest, _ := multibase.Encode(multibase.Base58BTC, delegatedBlob.Digest)

			return schema.NewSchemaError(fmt.Sprintf(
				"Link %s violates imposed %s constraint",
				claimedDigest, delegatedDigest,
			))
		}
	}

	// Check size constraint
	if claimedBlob.Size > 0 && delegatedBlob.Size > 0 {
		if claimedBlob.Size > delegatedBlob.Size {
			return schema.NewSchemaError(fmt.Sprintf(
				"Size constraint violation: %d > %d",
				claimedBlob.Size, delegatedBlob.Size,
			))
		}
	}

	return nil
}

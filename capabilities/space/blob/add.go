package blob

import (
	"bytes"
	"fmt"

	"github.com/ipld/go-ipld-prime/datamodel"
	"github.com/multiformats/go-multibase"
	"github.com/storacha/go-ucanto/core/ipld"
	"github.com/storacha/go-ucanto/core/result/failure"
	"github.com/storacha/go-ucanto/core/schema"
	"github.com/storacha/go-ucanto/ucan"
	"github.com/storacha/go-ucanto/validator"

	"github.com/storacha/go-libstoracha/capabilities/types"
)

const AddAbility = "space/blob/add"

// AddCaveats represents the caveats required to perform a space/blob/add invocation.
type AddCaveats struct {
	// Blob is the blob to be stored.
	Blob types.Blob
}

func (ac AddCaveats) ToIPLD() (datamodel.Node, error) {
	return ipld.WrapWithRecovery(&ac, AddCaveatsType(), types.Converters...)
}

var AddCaveatsReader = schema.Struct[AddCaveats](AddCaveatsType(), nil, types.Converters...)

// AddOk represents the result of a successful space/blob/add invocation.
type AddOk struct {
	// Receipt is a link to the receipt of the space/blob/add task.
	Site Promise
}

type Promise struct {
	UcanAwait Await
}

type Await struct {
	Selector string
	Link     ipld.Link
}

func (ao AddOk) ToIPLD() (datamodel.Node, error) {
	return ipld.WrapWithRecovery(&ao, AddOkType(), types.Converters...)
}

var AddOkReader = schema.Struct[AddOk](AddOkType(), nil, types.Converters...)

// Add is a capability that allows the agent to store a Blob into a space identified by did:key in the `with` field.
// The agent should compute the blob's multihash and size and provide it under the `nb.blob` field, allowing a
// service to provision a write location for the agent to PUT the desired blob into.
var Add = validator.NewCapability(
	AddAbility,
	schema.DIDString(),
	AddCaveatsReader,
	func(claimed, delegated ucan.Capability[AddCaveats]) failure.Failure {
		fail := equalWith(claimed, delegated)
		if fail != nil {
			return fail
		}

		return equalBlob(claimed, delegated)
	},
)

// equalWith validates that the claimed capability's `with` field matches the delegated one.
func equalWith(claimed, delegated ucan.Capability[AddCaveats]) failure.Failure {
	if claimed.With() != delegated.With() {
		return schema.NewSchemaError(fmt.Sprintf(
			"Resource '%s' doesn't match delegated '%s'",
			claimed.With(), delegated.With(),
		))
	}

	return nil
}

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

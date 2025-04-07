package http

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

const PutAbility = "http/put"

// Body describes the contents of the HTTP PUT request body.
type Body struct {
	// Digest is the multihash of the blob included in the request body.
	Digest multihash.Multihash

	// Size is the size in bytes of the blob included in the request body.
	Size uint64
}

// PutCaveats represents the caveats required to perform a http/put invocation.
type PutCaveats struct {
	URL     types.Promise
	Headers types.Promise
	Body    Body
}

func (pc PutCaveats) ToIPLD() (datamodel.Node, error) {
	return ipld.WrapWithRecovery(&pc, PutCaveatsType(), types.Converters...)
}

var PutCaveatsReader = schema.Struct[PutCaveats](PutCaveatsType(), nil, types.Converters...)

// PutOk represents the successful response for a http/put invocation.
type PutOk struct {
}

func (po PutOk) ToIPLD() (datamodel.Node, error) {
	return ipld.WrapWithRecovery(&po, PutOkType(), types.Converters...)
}

var PutOkReader = schema.Struct[PutOk](PutOkType(), nil, types.Converters...)

// Put is a capability that allows the agent to perform an HTTP PUT request.
var Put = validator.NewCapability(
	PutAbility,
	schema.DIDString(),
	PutCaveatsReader,
	func(claimed, delegated ucan.Capability[PutCaveats]) failure.Failure {
		// Check if the space matches
		if fail := equalWith(claimed, delegated); fail != nil {
			return fail
		}

		// Check URL constraint
		if fail := equalURL(claimed, delegated); fail != nil {
			return fail
		}

		// Check headers constraint
		if fail := equalHeaders(claimed, delegated); fail != nil {
			return fail
		}

		// Check body constraint
		if fail := equalBody(claimed, delegated); fail != nil {
			return fail
		}

		return nil
	},
)

// equalWith validates that the claimed capability's `with` field matches the delegated one.
func equalWith(claimed, delegated ucan.Capability[PutCaveats]) failure.Failure {
	if claimed.With() != delegated.With() {
		return schema.NewSchemaError(fmt.Sprintf(
			"resource '%s' doesn't match delegated '%s'",
			claimed.With(), delegated.With(),
		))
	}

	return nil
}

// equalURL checks if the URL matches between two capabilities.
func equalURL(claimed, delegated ucan.Capability[PutCaveats]) failure.Failure {
	if delegated.Nb().URL != claimed.Nb().URL {
		return schema.NewSchemaError(fmt.Sprintf(
			"claimed URL '%v' doesn't match delegated '%v'",
			claimed.Nb().URL, delegated.Nb().URL,
		))
	}

	return nil
}

// equalHeaders checks if the headers match between two capabilities.
func equalHeaders(claimed, delegated ucan.Capability[PutCaveats]) failure.Failure {
	if delegated.Nb().Headers != claimed.Nb().Headers {
		return schema.NewSchemaError(fmt.Sprintf(
			"claimed headers '%v' don't match delegated '%v'",
			claimed.Nb().Headers, delegated.Nb().Headers,
		))
	}

	return nil
}

// equalBody checks if the body description matches between two capabilities.
func equalBody(claimed, delegated ucan.Capability[PutCaveats]) failure.Failure {
	claimedBody := claimed.Nb().Body
	delegatedBody := delegated.Nb().Body

	// Check if the blob digest matches
	if delegatedBody.Digest != nil {
		if !bytes.Equal(delegatedBody.Digest, claimedBody.Digest) {
			claimedDigest, _ := multibase.Encode(multibase.Base58BTC, claimedBody.Digest)
			delegatedDigest, _ := multibase.Encode(multibase.Base58BTC, delegatedBody.Digest)

			return schema.NewSchemaError(fmt.Sprintf(
				"body digest '%s' doesn't match delegated '%s'",
				claimedDigest, delegatedDigest,
			))
		}
	}

	// Check size constraint
	if claimedBody.Size != delegatedBody.Size {
		return schema.NewSchemaError(fmt.Sprintf(
			"claimed body size (%d bytes) doesn't match delegated size (%d bytes)",
			claimedBody.Size, delegatedBody.Size,
		))
	}

	return nil
}

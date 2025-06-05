package access

import (
	"fmt"

	"github.com/ipld/go-ipld-prime/datamodel"
	"github.com/storacha/go-libstoracha/capabilities/types"
	"github.com/storacha/go-ucanto/core/ipld"
	"github.com/storacha/go-ucanto/core/result/failure"
	"github.com/storacha/go-ucanto/core/schema"
	"github.com/storacha/go-ucanto/did"
	"github.com/storacha/go-ucanto/ucan"
	"github.com/storacha/go-ucanto/validator"
)

const AuthorizeAbility = "access/authorize"

// CapabilityRequest represents a request for a specific capability in an
// `access/authorize` invocation.
type CapabilityRequest struct {
	// If set to `"*"` it corresponds to "sudo" access.
	Can string
}

// AuthorizeCaveats represents the caveats required to perform an
// access/authorize invocation.
type AuthorizeCaveats struct {
	// DID of the Account authorization is requested from.
	Iss *did.DID
	// Capabilities agent wishes to be granted.
	Att []CapabilityRequest
}

func (pc AuthorizeCaveats) ToIPLD() (datamodel.Node, error) {
	return ipld.WrapWithRecovery(&pc, AuthorizeCaveatsType(), types.Converters...)
}

var AuthorizeCaveatsReader = schema.Struct[AuthorizeCaveats](AuthorizeCaveatsType(), nil, types.Converters...)

// AuthorizeOk represents the successful response for a access/authorize
// invocation.
type AuthorizeOk struct {
	Request    ipld.Link
	Expiration ucan.UTCUnixTimestamp
}

func (po AuthorizeOk) ToIPLD() (datamodel.Node, error) {
	return ipld.WrapWithRecovery(&po, AuthorizeOkType(), types.Converters...)
}

var AuthorizeOkReader = schema.Struct[AuthorizeOk](AuthorizeOkType(), nil, types.Converters...)

// Authorize can be invoked by an agent to request set of capabilities from the
// account.
var Authorize = validator.NewCapability(
	AuthorizeAbility,
	schema.DIDString(),
	AuthorizeCaveatsReader,
	AuthorizeDerive,
)

func AuthorizeDerive(claimed, delegated ucan.Capability[AuthorizeCaveats]) failure.Failure {
	if fail := equalWith(claimed, delegated); fail != nil {
		return fail
	}

	if fail := equalIss(claimed, delegated); fail != nil {
		return fail
	}

	if fail := subsetCapabilities(claimed.Nb().Att, delegated.Nb().Att); fail != nil {
		return fail
	}

	return nil
}

// equalWith validates that the claimed capability's `with` field matches the delegated one.
func equalWith(claimed, delegated ucan.Capability[AuthorizeCaveats]) failure.Failure {
	if claimed.With() != delegated.With() {
		return schema.NewSchemaError(fmt.Sprintf(
			"resource '%s' doesn't match delegated '%s'",
			claimed.With(), delegated.With(),
		))
	}

	return nil
}

// equalIss checks if the issuer matches between two capabilities.
func equalIss(claimed, delegated ucan.Capability[AuthorizeCaveats]) failure.Failure {
	if delegated.Nb().Iss != claimed.Nb().Iss {
		return schema.NewSchemaError(fmt.Sprintf(
			"claimed issuer '%v' doesn't match delegated '%v'",
			claimed.Nb().Iss, delegated.Nb().Iss,
		))
	}

	return nil
}

// subsetCapabilities checks if the headers match between two capabilities.
func subsetCapabilities(claimed, delegated []CapabilityRequest) failure.Failure {
	delegatedCaps := make(map[string]bool)
	for _, cap := range delegated {
		delegatedCaps[cap.Can] = true
	}

	if delegatedCaps["*"] {
		// If everything is allowed, no need to check further because it contains
		// all the capabilities.
		return nil
	}

	// Otherwise we compute delta between what is allowed and what is requested.
	escalated := make([]string, 0, len(claimed))
	for _, cap := range claimed {
		if !delegatedCaps[cap.Can] {
			escalated = append(escalated, cap.Can)
		}
	}

	if len(escalated) > 0 {
		return schema.NewSchemaError(fmt.Sprintf(
			"unauthorized nb.att.can %v",
			escalated,
		))
	}

	return nil
}

package access

import (
	"fmt"

	"github.com/storacha/go-libstoracha/capabilities/types"
	"github.com/storacha/go-ucanto/core/result/failure"
	"github.com/storacha/go-ucanto/core/schema"
	"github.com/storacha/go-ucanto/ucan"
	"github.com/storacha/go-ucanto/validator"
)

const AccessAbility = "access/*"

// Access capability definition
// This capability can only be delegated (but not invoked) allowing audience to
// derive any `access/` prefixed capability for the space identified by the DID
// in the `with` field.
var Access = validator.NewCapability(
	AccessAbility,
	schema.DIDString(),
	schema.Struct[struct{}](nil, nil, types.Converters...),
	nil,
)

// CapabilityRequest represents a request for a specific capability in an
// `access/authorize` invocation.
type CapabilityRequest struct {
	// If set to `"*"` it corresponds to "sudo" access.
	Can string
}

// equalWith validates that the claimed capability's `with` field matches the delegated one.
func equalWith[Caveats any](claimed, delegated ucan.Capability[Caveats]) failure.Failure {
	if claimed.With() != delegated.With() {
		return schema.NewSchemaError(fmt.Sprintf(
			"resource '%s' doesn't match delegated '%s'",
			claimed.With(), delegated.With(),
		))
	}

	return nil
}

// equal checks if the values are equal. constraint should describe the field,
// and is used in the error message if the values don't match.
func equal(claimed, delegated any, constraint string) failure.Failure {
	if delegated != claimed {
		return schema.NewSchemaError(fmt.Sprintf(
			"claimed %s '%v' doesn't match delegated '%v'",
			constraint, claimed, delegated,
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

package access

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
	Iss did.DID
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
	return nil
}

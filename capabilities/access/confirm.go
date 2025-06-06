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

const ConfirmAbility = "access/confirm"

// ConfirmCaveats represents the caveats required to perform an
// access/confirm invocation.
type ConfirmCaveats struct {
	// Link to the `access/authorize` request that this delegation was created
	// for.
	Cause ipld.Link
	Iss   did.DID
	Aud   did.DID
	Att   []CapabilityRequest
}

func (pc ConfirmCaveats) ToIPLD() (datamodel.Node, error) {
	return ipld.WrapWithRecovery(&pc, ConfirmCaveatsType(), types.Converters...)
}

var ConfirmCaveatsReader = schema.Struct[ConfirmCaveats](ConfirmCaveatsType(), nil, types.Converters...)

type DelegationsModel struct {
	Keys   []string
	Values map[string][]byte
}

// ConfirmOk represents the successful response for a access/confirm
// invocation.
type ConfirmOk struct {
	Delegations DelegationsModel
}

func (po ConfirmOk) ToIPLD() (datamodel.Node, error) {
	return ipld.WrapWithRecovery(&po, ConfirmOkType(), types.Converters...)
}

var ConfirmOkReader = schema.Struct[ConfirmOk](ConfirmOkType(), nil, types.Converters...)

// Confirm can be invoked by an agent to request set of capabilities from the
// account.
var Confirm = validator.NewCapability(
	ConfirmAbility,
	schema.DIDString(),
	ConfirmCaveatsReader,
	ConfirmDerive,
)

func ConfirmDerive(claimed, delegated ucan.Capability[ConfirmCaveats]) failure.Failure {
	if fail := equalWith(claimed, delegated); fail != nil {
		return fail
	}

	if fail := equal(claimed.Nb().Iss, delegated.Nb().Iss, "iss"); fail != nil {
		return fail
	}

	if fail := equal(claimed.Nb().Aud, delegated.Nb().Aud, "aud"); fail != nil {
		return fail
	}

	if fail := subsetCapabilities(claimed.Nb().Att, delegated.Nb().Att); fail != nil {
		return fail
	}

	if fail := equal(claimed.Nb().Cause.String(), delegated.Nb().Cause.String(), "cause"); fail != nil {
		return fail
	}

	return nil
}

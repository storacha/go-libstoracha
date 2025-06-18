package access

import (
	"fmt"

	"github.com/ipld/go-ipld-prime/datamodel"
	"github.com/storacha/go-libstoracha/capabilities/types"
	"github.com/storacha/go-ucanto/core/ipld"
	"github.com/storacha/go-ucanto/core/result/failure"
	"github.com/storacha/go-ucanto/core/schema"
	"github.com/storacha/go-ucanto/ucan"
	"github.com/storacha/go-ucanto/validator"
)

const DelegateAbility = "access/delegate"

// DelegateCaveats represents the caveats required to perform an
// access/delegate invocation.
type DelegateCaveats struct {
	// The delegations to store.
	Delegations DelegationLinksModel
}

func (pc DelegateCaveats) ToIPLD() (datamodel.Node, error) {
	return ipld.WrapWithRecovery(&pc, DelegateCaveatsType(), types.Converters...)
}

var DelegateCaveatsReader = schema.Struct[DelegateCaveats](DelegateCaveatsType(), nil, types.Converters...)

// DelegateOk represents the successful response for a access/delegate
// invocation.
type DelegateOk struct {
}

func (po DelegateOk) ToIPLD() (datamodel.Node, error) {
	return ipld.WrapWithRecovery(&po, DelegateOkType(), types.Converters...)
}

var DelegateOkReader = schema.Struct[DelegateOk](DelegateOkType(), nil, types.Converters...)

// Delegate can be invoked by an agent to request set of capabilities from the
// account.
var Delegate = validator.NewCapability(
	DelegateAbility,
	schema.DIDString(schema.WithMethod("key")),
	DelegateCaveatsReader,
	DelegateDerive,
)

func DelegateDerive(claimed, delegated ucan.Capability[DelegateCaveats]) failure.Failure {
	if fail := equalWith(claimed, delegated); fail != nil {
		return fail
	}

	if fail := subsetDelegations(claimed.Nb().Delegations, delegated.Nb().Delegations); fail != nil {
		return fail
	}

	return nil
}

func subsetDelegations(claimed, delegated DelegationLinksModel) failure.Failure {
	disallowed := make([]string, 0, len(claimed.Values))
	for claimedCid := range claimed.Values {
		if delegated.Values[claimedCid] == nil {
			disallowed = append(disallowed, claimedCid)
		}
	}

	if len(disallowed) > 0 {
		return schema.NewSchemaError(fmt.Sprintf(
			"unauthorized delegations %v",
			disallowed,
		))
	}

	return nil
}

package access

import (
	"github.com/ipld/go-ipld-prime/datamodel"
	"github.com/storacha/go-libstoracha/capabilities/types"
	"github.com/storacha/go-ucanto/core/ipld"
	"github.com/storacha/go-ucanto/core/schema"
	"github.com/storacha/go-ucanto/validator"
)

const ClaimAbility = "access/claim"

type ClaimCaveats struct {
}

func (pc ClaimCaveats) ToIPLD() (datamodel.Node, error) {
	return ipld.WrapWithRecovery(&pc, ClaimCaveatsType(), types.Converters...)
}

var ClaimCaveatsReader = schema.Struct[ClaimCaveats](ClaimCaveatsType(), nil, types.Converters...)

// ClaimOk represents the successful response for a access/claim
// invocation.
type ClaimOk struct {
	Delegations DelegationsModel
}

func (po ClaimOk) ToIPLD() (datamodel.Node, error) {
	return ipld.WrapWithRecovery(&po, ClaimOkType(), types.Converters...)
}

var ClaimOkReader = schema.Struct[ClaimOk](ClaimOkType(), nil, types.Converters...)

// Claim can be invoked by an agent to request set of capabilities from the
// account.
var Claim = validator.NewCapability(
	ClaimAbility,
	schema.Or(
		schema.DIDString(schema.WithMethod("key")),
		schema.DIDString(schema.WithMethod("mailto")),
	),
	ClaimCaveatsReader,
	nil,
)

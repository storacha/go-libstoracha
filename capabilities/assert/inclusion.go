package assert

import (
	"github.com/ipld/go-ipld-prime/datamodel"
	"github.com/storacha/go-libstoracha/capabilities/types"
	"github.com/storacha/go-ucanto/core/ipld"
	"github.com/storacha/go-ucanto/core/schema"
	"github.com/storacha/go-ucanto/validator"
)

// InclusionAbility claims that a CID includes the contents claimed in another
// CID.
const InclusionAbility = "assert/inclusion"

type InclusionCaveats struct {
	Content  types.HasMultihash
	Includes ipld.Link
	Proof    *ipld.Link
}

func (ic InclusionCaveats) ToIPLD() (datamodel.Node, error) {
	return ipld.WrapWithRecovery(&ic, InclusionCaveatsType(), types.Converters...)
}

var InclusionCaveatsReader = schema.Struct[InclusionCaveats](InclusionCaveatsType(), nil, types.Converters...)

var Inclusion = validator.NewCapability(InclusionAbility, schema.DIDString(),
	InclusionCaveatsReader, nil)

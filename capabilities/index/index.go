package index

import (
	_ "embed"
	"fmt"

	ipldprime "github.com/ipld/go-ipld-prime"
	"github.com/ipld/go-ipld-prime/datamodel"
	ipldschema "github.com/ipld/go-ipld-prime/schema"
	"github.com/storacha/go-libstoracha/capabilities/types"
	"github.com/storacha/go-ucanto/core/ipld"
	"github.com/storacha/go-ucanto/core/result/failure"
	"github.com/storacha/go-ucanto/core/schema"
	"github.com/storacha/go-ucanto/ucan"
	"github.com/storacha/go-ucanto/validator"
)

//go:embed index.ipldsch
var assertSchema []byte

var assertTS = mustLoadTS()

func mustLoadTS() *ipldschema.TypeSystem {
	ts, err := ipldprime.LoadSchemaBytes(assertSchema)
	if err != nil {
		panic(fmt.Errorf("loading assert schema: %w", err))
	}
	return ts
}

func AddCaveatsType() ipldschema.Type {
	return assertTS.TypeByName("AddCaveats")
}

var IndexAbility = "space/index/*"

// Index capability definition
// This capability can only be delegated (but not invoked) allowing audience to
// derive any `space/index/` prefixed capability for the space identified by the DID
// in the `with` field.
var Index = validator.NewCapability(
	IndexAbility,
	schema.DIDString(),
	schema.Struct[struct{}](nil, nil, types.Converters...),
	func(claimed, delegated ucan.Capability[struct{}]) failure.Failure {
		return equalWith(claimed, delegated)
	},
)

// AddCaveats represents the arguments for the space/index/add capability
type AddCaveats struct {
	// Link is the Content Archive (CAR) containing the `Index`.
	Index ucan.Link
}

func (ic AddCaveats) ToIPLD() (datamodel.Node, error) {
	return ipld.WrapWithRecovery(&ic, AddCaveatsType(), types.Converters...)
}

var AddAbility = "space/index/add"

var AddCaveatsReader = schema.Struct[AddCaveats](AddCaveatsType(), nil, types.Converters...)

// Add capability definition
// This capability allows an agent to submit verifiable claims about content-addressable data
// to be published on the InterPlanetary Network Indexer (IPNI), making it publicly queryable.
var Add = validator.NewCapability(
	AddAbility,
	schema.DIDString(),
	schema.Struct[AddCaveats](nil, nil, types.Converters...),
	func(claimed, delegated ucan.Capability[AddCaveats]) failure.Failure {
		if fail := equalWith(claimed, delegated); fail != nil {
			return fail
		}

		if fail := equalIndex(claimed, delegated); fail != nil {
			return fail
		}

		return nil
	},
)

// equalWith validates that the claimed capability's `with` field matches the delegated one.
func equalWith[Caveats any](claimed, delegated ucan.Capability[Caveats]) failure.Failure {
	if claimed.With() != delegated.With() {
		return schema.NewSchemaError(fmt.Sprintf(
			"Resource '%s' doesn't match delegated '%s'",
			claimed.With(), delegated.With(),
		))
	}

	return nil
}

// equalIndex validates that the claimed capability's `index` field matches the
// delegated one's, if any
func equalIndex(claimed, delegated ucan.Capability[AddCaveats]) failure.Failure {
	claimedCaveats := claimed.Nb()
	delegatedCaveats := delegated.Nb()

	// If delegated doesn't specify an index, allow any index
	if delegatedCaveats.Index != nil && claimedCaveats.Index != nil {
		if claimedCaveats.Index.String() != delegatedCaveats.Index.String() {
			return schema.NewSchemaError(fmt.Sprintf(
				"index '%s' doesn't match delegated '%s'",
				claimedCaveats.Index.String(),
				delegatedCaveats.Index.String(),
			))
		}
	}

	return nil
}

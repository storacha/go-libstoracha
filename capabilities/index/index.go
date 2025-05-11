package index

import (
	"fmt"

	"github.com/storacha/go-libstoracha/capabilities/types"
	"github.com/storacha/go-ucanto/core/result/failure"
	"github.com/storacha/go-ucanto/core/schema"
	"github.com/storacha/go-ucanto/ucan"
	"github.com/storacha/go-ucanto/validator"
)

// IndexCaveats represents the arguments for the space/index/add capability
type IndexCaveats struct {
	// Link is the Content Archive (CAR) containing the `Index`.
	Index ucan.Link
}

var IndexAbility = "space/index/*"
var AddAbility = "space/index/add"

// Index capability definition
// This capability can only be delegated (but not invoked) allowing audience to
// derive any `space/index/` prefixed capability for the space identified by the DID
// in the `with` field.
var Index = validator.NewCapability(
	IndexAbility,
	schema.DIDString(),
	schema.Struct[struct{}](nil, nil, types.Converters...),
	func(claimed, delegated ucan.Capability[struct{}]) failure.Failure {
		return equalWith(claimed.With(), delegated.With())
	},
)

// Add capability definition
// This capability allows an agent to submit verifiable claims about content-addressable data
// to be published on the InterPlanetary Network Indexer (IPNI), making it publicly queryable.
var Add = validator.NewCapability(
	AddAbility,
	schema.DIDString(),
	schema.Struct[IndexCaveats](nil, nil, types.Converters...),
	func(claimed, delegated ucan.Capability[IndexCaveats]) failure.Failure {
		// Check if the `with` fields are equal
		if err := equalWith(claimed.With(), delegated.With()); err != nil {
			return err
		}

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
	},
)

// equalWith validates that the claimed capability's `with` field matches the delegated one.
func equalWith(claimed, delegated string) failure.Failure {
	if claimed != delegated {
		return schema.NewSchemaError(fmt.Sprintf(
			"resource '%s' doesn't match delegated '%s'",
			claimed, delegated,
		))
	}
	return nil
}

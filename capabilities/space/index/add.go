package index

import (
	_ "embed"
	"fmt"

	"github.com/ipld/go-ipld-prime/datamodel"
	"github.com/storacha/go-libstoracha/capabilities/types"
	"github.com/storacha/go-ucanto/core/ipld"
	"github.com/storacha/go-ucanto/core/result/failure"
	"github.com/storacha/go-ucanto/core/schema"
	"github.com/storacha/go-ucanto/ucan"
	"github.com/storacha/go-ucanto/validator"

	"github.com/storacha/go-ucanto/core/receipt"
)

const AddAbility = "space/index/add"

// AddCaveats represents the arguments for the space/index/add capability
type AddCaveats struct {
	// Link is the Content Archive (CAR) containing the `Index`.
	Index ucan.Link
}

func (ic AddCaveats) ToIPLD() (datamodel.Node, error) {
	return ipld.WrapWithRecovery(&ic, AddCaveatsType(), types.Converters...)
}

var AddCaveatsReader = schema.Struct[AddCaveats](AddCaveatsType(), nil, types.Converters...)

type AddOk struct {
}

func (ao AddOk) ToIPLD() (datamodel.Node, error) {
	return ipld.WrapWithRecovery(&ao, AddOkType(), types.Converters...)
}

type AddReceipt receipt.Receipt[AddOk, failure.Failure]

type AddReceiptReader receipt.ReceiptReader[AddOk, failure.Failure]

func NewAddReceiptReader() (AddReceiptReader, error) {
	return receipt.NewReceiptReader[AddOk, failure.Failure](indexSchema)
}

var AddOkReader = schema.Struct[AddOk](AddOkType(), nil, types.Converters...)

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

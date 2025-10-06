package access

import (
	"fmt"

	"github.com/ipld/go-ipld-prime/datamodel"
	"github.com/storacha/go-libstoracha/capabilities/types"
	"github.com/storacha/go-ucanto/core/ipld"
	"github.com/storacha/go-ucanto/core/receipt"
	"github.com/storacha/go-ucanto/core/result/failure"
	"github.com/storacha/go-ucanto/core/schema"
	"github.com/storacha/go-ucanto/ucan"
	"github.com/storacha/go-ucanto/validator"
)

const GrantAbility = "access/grant"

// GrantCaveats are the caveats required to perform an access/grant invocation.
type GrantCaveats struct {
	// Att are the capabilities agent wishes to be granted.
	Att []CapabilityRequest
	// Cause is an OPTIONAL link to a UCAN that provides context for the grant
	// request. The linked UCAN MUST be included in the invocation.
	Cause ucan.Link
}

func (gc GrantCaveats) ToIPLD() (datamodel.Node, error) {
	return ipld.WrapWithRecovery(&gc, GrantCaveatsType(), types.Converters...)
}

var GrantCaveatsReader = schema.Struct[GrantCaveats](GrantCaveatsType(), nil, types.Converters...)

// GrantOk represents the successful response for a access/grant invocation.
type GrantOk struct {
	Delegations DelegationsModel
}

func (gok GrantOk) ToIPLD() (datamodel.Node, error) {
	return ipld.WrapWithRecovery(&gok, GrantOkType(), types.Converters...)
}

var GrantOkReader = schema.Struct[GrantOk](GrantOkType(), nil, types.Converters...)

type GrantError struct {
	Name    string
	Message string
}

func (ge GrantError) ToIPLD() (datamodel.Node, error) {
	return ipld.WrapWithRecovery(&ge, GrantErrorType(), types.Converters...)
}

type GrantReceipt receipt.Receipt[GrantOk, GrantError]
type GrantReceiptReader receipt.ReceiptReader[GrantOk, GrantError]

func NewGrantReceiptReader() (GrantReceiptReader, error) {
	return receipt.NewReceiptReaderFromTypes[GrantOk, GrantError](GrantOkType(), GrantErrorType())
}

// Grant is a capability that allows an agent to request capabilities from the
// invocation executor.
var Grant = validator.NewCapability(
	GrantAbility,
	schema.DIDString(),
	GrantCaveatsReader,
	GrantDerive,
)

func GrantDerive(claimed, delegated ucan.Capability[GrantCaveats]) failure.Failure {
	return schema.NewSchemaError(fmt.Sprintf("%s cannot be delegated", GrantAbility))
}

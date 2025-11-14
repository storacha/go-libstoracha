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

const (
	GrantAbility = "access/grant"
	// UnknownAbilityErrorName is the name given to an error where the ability
	// requested to be granted is unknown to the service.
	UnknownAbilityErrorName = "UnknownAbility"
	// MissingCapabilityErrorName is the name given to an error caused by the list
	// of requested capabilities being empty.
	MissingCapabilityErrorName = "MissingCapability"
	// UnknownCauseErrorName is the name given to an error where the cause
	// invocation sent as context for the delegation is not recognised.
	UnknownCauseErrorName = "UnknownCause"
	// MissingCauseErrorName is the name given to an error where a required cause
	// invocation has not been provided in the invocation to request a grant.
	MissingCauseErrorName = "MissingCause"
	// InvalidCauseErrorName is the name given to an error where the cause
	// invocation has been determined to be invalid is some way. See the error
	// message for details.
	InvalidCauseErrorName = "InvalidCause"
	// UnauthorizedCauseErrorName is the name given to an error where the cause
	// invocation failed UCAN validation.
	UnauthorizedCauseErrorName = "UnauthorizedCause"
)

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
	ErrorName string
	Message   string
}

func NewUnknownAbilityError(ability string) GrantError {
	return GrantError{
		ErrorName: UnknownAbilityErrorName,
		Message:   fmt.Sprintf("unknown ability: %s", ability),
	}
}

func NewMissingCapabilityError() GrantError {
	return GrantError{
		ErrorName: MissingCapabilityErrorName,
		Message:   "grant requires one or more capabilities",
	}
}

func NewMissingCauseError() GrantError {
	return GrantError{
		ErrorName: MissingCauseErrorName,
		Message:   "grant requires supporting contextual invocation",
	}
}

func NewUnknownCauseError() GrantError {
	return GrantError{
		ErrorName: UnknownCauseErrorName,
		Message:   "unknown cause invocation",
	}
}

func NewInvalidCauseError(msg string) GrantError {
	return GrantError{
		ErrorName: InvalidCauseErrorName,
		Message:   fmt.Sprintf("invalid cause invocation: %s", msg),
	}
}

func NewUnauthorizedCauseError(err validator.Unauthorized) GrantError {
	return GrantError{
		ErrorName: UnauthorizedCauseErrorName,
		Message:   err.Error(),
	}
}

func (ge GrantError) Name() string {
	return ge.ErrorName
}

func (ge GrantError) Error() string {
	return ge.Message
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

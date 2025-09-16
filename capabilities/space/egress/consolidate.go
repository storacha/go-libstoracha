package egress

import (
	"github.com/ipld/go-ipld-prime/datamodel"
	captypes "github.com/storacha/go-libstoracha/capabilities/types"
	"github.com/storacha/go-ucanto/core/ipld"
	"github.com/storacha/go-ucanto/core/receipt"
	"github.com/storacha/go-ucanto/core/schema"
	"github.com/storacha/go-ucanto/ucan"
	"github.com/storacha/go-ucanto/validator"
)

const ConsolidateAbility = "space/egress/consolidate"

type ConsolidateCaveats struct {
	Cause ucan.Link
}

func (cc ConsolidateCaveats) ToIPLD() (datamodel.Node, error) {
	return ipld.WrapWithRecovery(&cc, ConsolidateCaveatsType(), captypes.Converters...)
}

var ConsolidateCaveatsReader = schema.Struct[ConsolidateCaveats](ConsolidateCaveatsType(), nil, captypes.Converters...)

type ReceiptError struct {
	Name    string
	Message string
	Receipt ucan.Link
}

type ConsolidateOk struct {
	Errors []ReceiptError
}

func (co ConsolidateOk) ToIPLD() (datamodel.Node, error) {
	return ipld.WrapWithRecovery(&co, ConsolidateOkType(), captypes.Converters...)
}

var ConsolidateOkReader = schema.Struct[ConsolidateOk](ConsolidateOkType(), nil, captypes.Converters...)

type ConsolidateError struct {
	Name    string
	Message string
}

func NewConsolidateError(msg string) ConsolidateError {
	return ConsolidateError{
		Name:    "EgressConsolidateError",
		Message: msg,
	}
}

func (ce ConsolidateError) ToIPLD() (datamodel.Node, error) {
	return ipld.WrapWithRecovery(&ce, ConsolidateErrorType(), captypes.Converters...)
}

var ConsolidateErrorReader = schema.Struct[ConsolidateError](ConsolidateErrorType(), nil, captypes.Converters...)

type ConsolidateReceipt receipt.Receipt[ConsolidateOk, ConsolidateError]

type ConsolidateReceiptReader receipt.ReceiptReader[ConsolidateOk, ConsolidateError]

func NewConsolidateReceiptReader() (ConsolidateReceiptReader, error) {
	return receipt.NewReceiptReader[ConsolidateOk, ConsolidateError](egressSchema)
}

// EgressTrack capability definition
// This capability allows a storage node to request recording egress for content it has served.
var Consolidate = validator.NewCapability(
	ConsolidateAbility,
	schema.DIDString(),
	ConsolidateCaveatsReader,
	validator.DefaultDerives,
)

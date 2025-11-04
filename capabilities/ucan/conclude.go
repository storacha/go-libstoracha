package ucan

import (
	"fmt"
	"time"

	"github.com/ipld/go-ipld-prime/datamodel"
	"github.com/storacha/go-libstoracha/capabilities/types"
	"github.com/storacha/go-ucanto/core/ipld"
	"github.com/storacha/go-ucanto/core/result/failure"
	"github.com/storacha/go-ucanto/core/schema"
	"github.com/storacha/go-ucanto/ucan"
	"github.com/storacha/go-ucanto/validator"
)

const ConcludeAbility = "ucan/conclude"

// ConcludeCaveats represents the caveats required to perform a ucan/conclude invocation.
type ConcludeCaveats struct {
	// Receipt is the CID of the content with the receipt.
	Receipt ipld.Link
}

func (cc ConcludeCaveats) ToIPLD() (datamodel.Node, error) {
	return ipld.WrapWithRecovery(&cc, ConcludeCaveatsType(), types.Converters...)
}

var ConcludeCaveatsReader = schema.Struct[ConcludeCaveats](ConcludeCaveatsType(), nil, types.Converters...)

// ConcludeOk represents the successful response for a ucan/conclude invocation.
type ConcludeOk struct {
	// Time is the timestamp when ucan/conclude invocation was completed.
	Time time.Time
}

func (co ConcludeOk) ToIPLD() (datamodel.Node, error) {
	return ipld.WrapWithRecovery(&co, ConcludeOkType(), types.Converters...)
}

var ConcludeOkReader = schema.Struct[ConcludeOk](ConcludeOkType(), nil, types.Converters...)

// Conclude is a capability that represents a receipt using a special UCAN capability.
var Conclude = validator.NewCapability(
	ConcludeAbility,
	schema.DIDString(),
	ConcludeCaveatsReader,
	func(claimed, delegated ucan.Capability[ConcludeCaveats]) failure.Failure {
		// Check if the with field matches
		if fail := equalWith(claimed, delegated); fail != nil {
			return fail
		}

		// Check if the receipt matches
		if fail := equalReceipt(claimed, delegated); fail != nil {
			return fail
		}

		return nil
	},
)

// equalWith validates that the claimed capability's `with` field matches the delegated one.
func equalWith(claimed, delegated ucan.UnknownCapability) failure.Failure {
	if claimed.With() != delegated.With() {
		return schema.NewSchemaError(fmt.Sprintf(
			"Resource '%s' doesn't match delegated '%s'",
			claimed.With(), delegated.With(),
		))
	}

	return nil
}

// equalReceipt checks if the receipt matches between two capabilities.
func equalReceipt(claimed, delegated ucan.Capability[ConcludeCaveats]) failure.Failure {
	if delegated.Nb().Receipt != claimed.Nb().Receipt {
		return schema.NewSchemaError(fmt.Sprintf(
			"claimed receipt '%s' doesn't match delegated '%s'",
			claimed.Nb().Receipt, delegated.Nb().Receipt,
		))
	}

	return nil
}

package content

import (
	"bytes"
	"errors"
	"fmt"

	"github.com/ipld/go-ipld-prime/datamodel"
	"github.com/multiformats/go-multibase"
	mh "github.com/multiformats/go-multihash"
	"github.com/storacha/go-ucanto/core/ipld"
	"github.com/storacha/go-ucanto/core/receipt"
	"github.com/storacha/go-ucanto/core/result/failure"
	fdm "github.com/storacha/go-ucanto/core/result/failure/datamodel"
	"github.com/storacha/go-ucanto/core/schema"
	"github.com/storacha/go-ucanto/ucan"
	"github.com/storacha/go-ucanto/validator"

	"github.com/storacha/go-libstoracha/capabilities/types"
)

const RetrieveAbility = "space/content/retrieve"

// RetrieveCaveats represents the caveats required to perform a space/content/retrieve invocation.
type RetrieveCaveats struct {
	Blob  BlobDigest
	Range Range
}

type BlobDigest struct {
	Digest mh.Multihash
}

// Range represents a range of byte offsets to retrieve.
// `Start` is the start offset from which to extract bytes. `End` is the offset at which extraction should end.
// Both offsets are inclusive.
type Range struct {
	Start uint64
	End   uint64
}

func (rc RetrieveCaveats) ToIPLD() (datamodel.Node, error) {
	return ipld.WrapWithRecovery(&rc, RetrieveCaveatsType(), types.Converters...)
}

var RetrieveCaveatsReader = schema.Struct[RetrieveCaveats](RetrieveCaveatsType(), nil, types.Converters...)

// RetrieveOk represents the result of a successful space/content/retrieve invocation.
type RetrieveOk struct{}

func (ro RetrieveOk) ToIPLD() (datamodel.Node, error) {
	return ipld.WrapWithRecovery(&ro, RetrieveOkType(), types.Converters...)
}

type RetrieveReceipt receipt.Receipt[RetrieveOk, failure.Failure]

type RetrieveReceiptReader receipt.ReceiptReader[RetrieveOk, failure.Failure]

func NewRetrieveReceiptReader() (RetrieveReceiptReader, error) {
	return receipt.NewReceiptReader[RetrieveOk, failure.Failure](contentSchema)
}

var RetrieveOkReader = schema.Struct[RetrieveOk](RetrieveOkType(), nil, types.Converters...)

type NotFoundError struct {
	name    string
	message string
}

func NewNotFoundError(msg string) NotFoundError {
	return NotFoundError{
		name:    "NotFound",
		message: msg,
	}
}

func (nfe NotFoundError) Name() string {
	return nfe.name
}

func (nfe NotFoundError) Error() string {
	return nfe.message
}

func (nfe NotFoundError) ToIPLD() (datamodel.Node, error) {
	return ipld.WrapWithRecovery(&nfe, NotFoundErrorType(), types.Converters...)
}

var NotFoundErrorReader = schema.Mapped(
	schema.Struct[fdm.FailureModel](fdm.FailureType(), nil, types.Converters...),
	func(f fdm.FailureModel) (NotFoundError, failure.Failure) {
		if f.Name == nil {
			return NotFoundError{}, failure.FromError(errors.New("missing error name"))
		}
		if *f.Name != "NotFound" {
			return NotFoundError{}, failure.FromError(fmt.Errorf("incorrect name: %s, expected: NotFound", *f.Name))
		}
		return NewNotFoundError(f.Message), nil
	},
)

type RangeNotSatisfiableError struct {
	name    string
	message string
}

func NewRangeNotSatisfiableError(msg string) RangeNotSatisfiableError {
	return RangeNotSatisfiableError{
		name:    "RangeNotSatisfiable",
		message: msg,
	}
}

func (rnse RangeNotSatisfiableError) Name() string {
	return rnse.name
}

func (rnse RangeNotSatisfiableError) Error() string {
	return rnse.message
}

func (rnse RangeNotSatisfiableError) ToIPLD() (datamodel.Node, error) {
	return ipld.WrapWithRecovery(&rnse, RangeNotSatisfiableErrorType(), types.Converters...)
}

var RangeNotSatisfiableErrorReader = schema.Mapped(
	schema.Struct[fdm.FailureModel](fdm.FailureType(), nil, types.Converters...),
	func(f fdm.FailureModel) (RangeNotSatisfiableError, failure.Failure) {
		if f.Name == nil {
			return RangeNotSatisfiableError{}, failure.FromError(errors.New("missing error name"))
		}
		if *f.Name != "RangeNotSatisfiable" {
			return RangeNotSatisfiableError{}, failure.FromError(fmt.Errorf("incorrect name: %s, expected: NotFound", *f.Name))
		}
		return NewRangeNotSatisfiableError(f.Message), nil
	},
)

// Retrieve is a capability that allows the agent to retrieve a byte range from a blob in a space.
var Retrieve = validator.NewCapability(
	RetrieveAbility,
	schema.DIDString(),
	RetrieveCaveatsReader,
	func(claimed, delegated ucan.Capability[RetrieveCaveats]) failure.Failure {
		fail := equalWith(claimed, delegated)
		if fail != nil {
			return fail
		}

		fail = equalDigest(claimed, delegated)
		if fail != nil {
			return fail
		}

		return validRange(claimed, delegated)
	},
)

// equalWith validates that the claimed capability's `with` field matches the delegated one.
func equalWith(claimed, delegated ucan.Capability[RetrieveCaveats]) failure.Failure {
	if claimed.With() != delegated.With() {
		return schema.NewSchemaError(fmt.Sprintf(
			"Resource '%s' doesn't match delegated '%s'",
			claimed.With(), delegated.With(),
		))
	}

	return nil
}

// equalDigest validates that the claimed digest capability matches the delegated one.
func equalDigest(claimed, delegated ucan.Capability[RetrieveCaveats]) failure.Failure {
	claimedDigest := claimed.Nb().Blob.Digest
	delegatedDigest := delegated.Nb().Blob.Digest

	// Check if the claimed digest matches
	if !bytes.Equal(delegatedDigest, claimedDigest) {
		claimedDigest, _ := multibase.Encode(multibase.Base58BTC, claimedDigest)
		delegatedDigest, _ := multibase.Encode(multibase.Base58BTC, delegatedDigest)

		return schema.NewSchemaError(fmt.Sprintf(
			"Digest %s violates imposed %s constraint",
			claimedDigest, delegatedDigest,
		))
	}

	return nil
}

func validRange(claimed, delegated ucan.Capability[RetrieveCaveats]) failure.Failure {
	claimedRange := claimed.Nb().Range
	delegatedRange := delegated.Nb().Range

	if claimedRange.Start < delegatedRange.Start {
		return schema.NewSchemaError(fmt.Sprintf(
			"Start offset %d violates imposed %d constraint",
			claimedRange.Start, delegatedRange.Start,
		))
	}

	if claimedRange.End > delegatedRange.End {
		return schema.NewSchemaError(fmt.Sprintf(
			"End offset %d violates imposed %d constraint",
			claimedRange.End, delegatedRange.End,
		))
	}

	return nil
}

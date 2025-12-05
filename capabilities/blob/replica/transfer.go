package replica

import (
	"github.com/ipld/go-ipld-prime/datamodel"
	"github.com/storacha/go-ucanto/did"

	"github.com/storacha/go-ucanto/core/ipld"
	"github.com/storacha/go-ucanto/core/receipt"
	"github.com/storacha/go-ucanto/core/result/failure"
	"github.com/storacha/go-ucanto/core/schema"
	"github.com/storacha/go-ucanto/ucan"
	"github.com/storacha/go-ucanto/validator"

	"github.com/storacha/go-libstoracha/capabilities/types"
)

const TransferAbility = "blob/replica/transfer"

var _ ipld.Builder = (*TransferCaveats)(nil)

type TransferCaveats struct {
	// Space contains the did to transfer Blob to.
	Space did.DID
	// Blob is the blob to be transferred.
	Blob types.Blob
	// Site contains a location commitment indicating where the Blob must be
	// transferred from.
	Site ucan.Link
	// Cause contains the `blob/replica/allocate` invocation that initiated this transfer.
	Cause ucan.Link
}

type TransferOk struct {
	// Site contains the location commitment indicate where the Blob has been
	// transferred to.
	Site ucan.Link
	// PDP optionally contains the PDP invocation will complete when aggregation
	// is complete and the piece is accepted.
	PDP *ucan.Link
}

func (t TransferOk) ToIPLD() (datamodel.Node, error) {
	return ipld.WrapWithRecovery(&t, TransferOkType(), types.Converters...)
}

var TransferOkReader = schema.Struct[TransferOk](TransferOkType(), nil, types.Converters...)

type TransferError struct {
	Name    string
	Message string
}

func NewTransferError(msg string) TransferError {
	return TransferError{
		Name:    "TransferError",
		Message: msg,
	}
}

func (t TransferError) ToIPLD() (datamodel.Node, error) {
	return ipld.WrapWithRecovery(&t, TransferErrorType(), types.Converters...)
}

var TransferErrorReader = schema.Struct[TransferError](TransferErrorType(), nil, types.Converters...)

func (tc TransferCaveats) ToIPLD() (datamodel.Node, error) {
	return ipld.WrapWithRecovery(&tc, TransferCaveatsType(), types.Converters...)
}

type TransferReceipt receipt.Receipt[TransferOk, failure.Failure]
type TransferReceiptReader receipt.ReceiptReader[TransferOk, failure.Failure]

func NewTransferReceiptReader() (TransferReceiptReader, error) {
	return receipt.NewReceiptReader[TransferOk, failure.Failure](replicaSchema, types.Converters...)
}

var TransferCaveatsReader = schema.Struct[TransferCaveats](TransferCaveatsType(), nil, types.Converters...)

// Transfer is a capability that allows an agent to transfer a Blob for replication
// into a space identified by did:key in the `with` field.
var Transfer = validator.NewCapability(
	TransferAbility,
	schema.DIDString(),
	TransferCaveatsReader,
	validator.DefaultDerives,
)

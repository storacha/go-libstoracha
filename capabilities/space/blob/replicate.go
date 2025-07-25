package blob

import (
	"github.com/ipld/go-ipld-prime/datamodel"
	"github.com/storacha/go-ucanto/core/ipld"
	"github.com/storacha/go-ucanto/core/receipt"
	"github.com/storacha/go-ucanto/core/result/failure"
	"github.com/storacha/go-ucanto/core/schema"
	"github.com/storacha/go-ucanto/validator"

	"github.com/storacha/go-libstoracha/capabilities/types"
)

const ReplicateAbility = "space/blob/replicate"

var _ ipld.Builder = (*ReplicateCaveats)(nil)

type ReplicateCaveats struct {
	// Blob is the blob that must be replicated.
	Blob types.Blob
	// Replicas is the number of replicas to ensure.
	// e.g. Replicas: 2 will ensure 3 copies of the data exist in a network.
	Replicas uint
	// Site contains a location commitment indicating where the Blob must be
	// fetched from.
	Site ipld.Link
}

func (rc ReplicateCaveats) ToIPLD() (datamodel.Node, error) {
	return ipld.WrapWithRecovery(&rc, ReplicateCaveatsType(), types.Converters...)
}

var ReplicateCaveatsReader = schema.Struct[ReplicateCaveats](ReplicateCaveatsType(), nil, types.Converters...)

// ReplicateOk represents the result of a successful space/blob/replicate invocation.
type ReplicateOk struct {
	// Site resolves to additional locations for the blob.
	// Their selector MUST be ".out.ok.site" and they link to a receipt of a
	// "blob/replica/transfer" task.
	Site []types.Promise
}

func (ro ReplicateOk) ToIPLD() (datamodel.Node, error) {
	return ipld.WrapWithRecovery(&ro, ReplicateOkType(), types.Converters...)
}

type ReplicateReceipt receipt.Receipt[ReplicateOk, failure.Failure]
type ReplicateReceiptReader receipt.ReceiptReader[ReplicateOk, failure.Failure]

func NewReplicateReceiptReader() (ReplicateReceiptReader, error) {
	return receipt.NewReceiptReader[ReplicateOk, failure.Failure](blobSchema)
}

var ReplicateOkReader = schema.Struct[ReplicateOk](ReplicateOkType(), nil, types.Converters...)

// Replicate is a capability that allows an agent to replicate a Blob into a
// space identified by did:key in the `with` field.
//
// A Replicate capability may only be invoked after a `blob/accept` receipt has
// been receieved, indicating the source node has successfully received the blob.
// Each Replicate task MUST target a different node, and they MUST NOT target
// the original upload target.
//
// The Replicate task receipt includes async tasks for `blob/replica/allocate`
// and `blob/replica/transfer`. Successful completion of the
// `blob/replica/transfer` task indicates the replication target has transferred
// and stored the blob. The number of `blob/replica/allocate` and
// `blob/replica/transfer tasks corresponds directly to number of replicas
// requested.
var Replicate = validator.NewCapability(
	ReplicateAbility,
	schema.DIDString(),
	ReplicateCaveatsReader,
	validator.DefaultDerives,
)

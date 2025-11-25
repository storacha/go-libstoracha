package replica

import (
	"github.com/ipld/go-ipld-prime/datamodel"
	"github.com/storacha/go-ucanto/core/ipld"
	"github.com/storacha/go-ucanto/core/receipt"
	"github.com/storacha/go-ucanto/core/result/failure"
	"github.com/storacha/go-ucanto/core/schema"
	"github.com/storacha/go-ucanto/did"
	"github.com/storacha/go-ucanto/ucan"
	"github.com/storacha/go-ucanto/validator"

	"github.com/storacha/go-libstoracha/capabilities/types"
)

const AllocateAbility = "blob/replica/allocate"

// AllocateSiteSelector is the selector for extracting the location commitment
// link for a new site from a "blob/replica/transfer" receipt.
const AllocateSiteSelector = ".out.ok.site"

var _ ipld.Builder = (*AllocateCaveats)(nil)

type AllocateCaveats struct {
	// Space contains the did to allocate Blob in.
	Space did.DID
	// Blob is the blob to be allocated.
	Blob types.Blob
	// Site contains a location commitment indicating where the Blob must be
	// fetched from.
	Site ucan.Link
	// Cause contains the `space/blob/replicate` invocation that caused this allocation.
	Cause ucan.Link
}

type AllocateOk struct {
	// Size is the number of bytes allocated for a Blob.
	Size uint64
	// Site resolves to an additional location for the blob.
	// The selector MUST be ".out.ok.site" i.e. [AllocateSiteSelector] and it
	// links to a receipt of a "blob/replica/transfer" task.
	Site types.Promise
}

func (a AllocateOk) ToIPLD() (ipld.Node, error) {
	return ipld.WrapWithRecovery(&a, AllocateOkType(), types.Converters...)
}

var AllocateOkReader = schema.Struct[AllocateOk](AllocateOkType(), nil, types.Converters...)

func (ac AllocateCaveats) ToIPLD() (datamodel.Node, error) {
	return ipld.WrapWithRecovery(&ac, AllocateCaveatsType(), types.Converters...)
}

var AllocateCaveatsReader = schema.Struct[AllocateCaveats](AllocateCaveatsType(), nil, types.Converters...)

type AllocateReceipt receipt.Receipt[AllocateOk, failure.Failure]
type AllocateReceiptReader receipt.ReceiptReader[AllocateOk, failure.Failure]

func NewAllocateReceiptReader() (AllocateReceiptReader, error) {
	return receipt.NewReceiptReader[AllocateOk, failure.Failure](replicaSchema)
}

// Allocate is a capability that allows an agent to allocate a Blob for replication
// into a space identified by did:key in the `with` field.
//
// The Allocate task receipt includes an async task that will be performed by
// a storage node - `blob/replica/transfer`. The `blob/replica/transfer` task is
// completed when the storage node has transferred the blob from its location to the storage node.
var Allocate = validator.NewCapability(
	AllocateAbility,
	schema.DIDString(),
	AllocateCaveatsReader,
	validator.DefaultDerives,
)

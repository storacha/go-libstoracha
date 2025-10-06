package blob

import (
	"bytes"
	"fmt"

	"github.com/ipld/go-ipld-prime/datamodel"
	"github.com/multiformats/go-multihash"
	"github.com/storacha/go-libstoracha/capabilities/types"
	"github.com/storacha/go-libstoracha/digestutil"
	"github.com/storacha/go-ucanto/core/ipld"
	"github.com/storacha/go-ucanto/core/receipt"
	"github.com/storacha/go-ucanto/core/result/failure"
	"github.com/storacha/go-ucanto/core/schema"
	"github.com/storacha/go-ucanto/ucan"
	"github.com/storacha/go-ucanto/validator"
)

const RetrieveAbility = "blob/retrieve"

// Blob represents a blob to be retrieved.
type Blob struct {
	// Digest is the multihash of the blob bytes, uniquely identifying the blob.
	Digest multihash.Multihash
}

// RetrieveCaveats are the caveats required to perform an blob/retrieve invocation.
type RetrieveCaveats struct {
	Blob Blob
}

func (gc RetrieveCaveats) ToIPLD() (datamodel.Node, error) {
	return ipld.WrapWithRecovery(&gc, RetrieveCaveatsType(), types.Converters...)
}

var RetrieveCaveatsReader = schema.Struct[RetrieveCaveats](RetrieveCaveatsType(), nil, types.Converters...)

// RetrieveOk represents the successful response for a blob/retrieve invocation.
type RetrieveOk struct{}

func (ro RetrieveOk) ToIPLD() (datamodel.Node, error) {
	return ipld.WrapWithRecovery(&ro, RetrieveOkType())
}

var RetrieveOkReader = schema.Struct[RetrieveOk](RetrieveOkType(), nil)

type RetrieveError struct {
	Name    string
	Message string
}

func (re RetrieveError) Error() string {
	return re.Message
}

func (re RetrieveError) ToIPLD() (datamodel.Node, error) {
	return ipld.WrapWithRecovery(&re, RetrieveErrorType())
}

type RetrieveReceipt receipt.Receipt[RetrieveOk, RetrieveError]
type RetrieveReceiptReader receipt.ReceiptReader[RetrieveOk, RetrieveError]

func NewRetrieveReceiptReader() (RetrieveReceiptReader, error) {
	return receipt.NewReceiptReaderFromTypes[RetrieveOk, RetrieveError](RetrieveOkType(), RetrieveErrorType())
}

// Retrieve is a service capability that allows an authorized agent to retrieve
// blob bytes from another agent that is hosting/storing the blob. i.e. there is
// no space authorization - the resource is the agent storing the blob.
var Retrieve = validator.NewCapability(
	RetrieveAbility,
	schema.DIDString(),
	RetrieveCaveatsReader,
	RetrieveDerive,
)

func RetrieveDerive(claimed, delegated ucan.Capability[RetrieveCaveats]) failure.Failure {
	if fail := validator.DefaultDerives(claimed, delegated); fail != nil {
		return fail
	}
	if !bytes.Equal(delegated.Nb().Blob.Digest, claimed.Nb().Blob.Digest) {
		return schema.NewSchemaError(fmt.Sprintf(
			"claimed digest %v doesn't match delegated digest %v",
			digestutil.Format(claimed.Nb().Blob.Digest),
			digestutil.Format(delegated.Nb().Blob.Digest),
		))
	}
	return nil
}

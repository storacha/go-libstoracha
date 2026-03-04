package shard

import (
	"fmt"

	"github.com/ipld/go-ipld-prime/datamodel"
	"github.com/storacha/go-ucanto/core/ipld"
	"github.com/storacha/go-ucanto/core/receipt"
	"github.com/storacha/go-ucanto/core/result/failure"
	"github.com/storacha/go-ucanto/core/schema"
	"github.com/storacha/go-ucanto/ucan"
	"github.com/storacha/go-ucanto/validator"

	"github.com/storacha/go-libstoracha/capabilities/types"
)

const (
	ListAbility = "upload/shard/list"
	// UploadNotFoundErrorName is the name given to an error where the upload
	// associated with the invocation is not found.
	UploadNotFoundErrorName = "UploadNotFound"
)

type ListCaveats struct {
	Root   ipld.Link
	Cursor *string
	Size   *uint64
}

func (lc ListCaveats) ToIPLD() (datamodel.Node, error) {
	return ipld.WrapWithRecovery(&lc, ListCaveatsType(), types.Converters...)
}

var ListCaveatsReader = schema.Struct[ListCaveats](ListCaveatsType(), nil, types.Converters...)

type ListOk struct {
	Cursor  *string
	Size    uint64
	Results []ipld.Link
}

func (lo ListOk) ToIPLD() (datamodel.Node, error) {
	return ipld.WrapWithRecovery(&lo, ListOkType(), types.Converters...)
}

var ListOkReader = schema.Struct[ListOk](ListOkType(), nil, types.Converters...)

type ListError struct {
	ErrorName string
	Message   string
}

func NewUploadNotFoundError(root ipld.Link) ListError {
	return ListError{
		ErrorName: UploadNotFoundErrorName,
		Message:   fmt.Sprintf("upload not found: %s", root.String()),
	}
}

func (e ListError) Name() string {
	return e.ErrorName
}

func (e ListError) Error() string {
	return e.Message
}

func (e ListError) ToIPLD() (datamodel.Node, error) {
	return ipld.WrapWithRecovery(&e, ListErrorType(), types.Converters...)
}

type ListReceipt receipt.Receipt[ListOk, ListError]
type ListReceiptReader receipt.ReceiptReader[ListOk, ListError]

func NewListReceiptReader() (ListReceiptReader, error) {
	return receipt.NewReceiptReaderFromTypes[ListOk, ListError](ListOkType(), ListErrorType(), types.Converters...)
}

var List = validator.NewCapability(
	ListAbility,
	schema.DIDString(),
	ListCaveatsReader,
	ListDerive,
)

func ListDerive(claimed, delegated ucan.Capability[ListCaveats]) failure.Failure {
	if fail := validator.DefaultDerives(claimed, delegated); fail != nil {
		return fail
	}

	if claimed.Nb().Root.String() != delegated.Nb().Root.String() {
		return failure.FromError(fmt.Errorf("constraint violation: %q violates imposed root constraint %q", claimed.Nb().Root.String(), delegated.Nb().Root.String()))
	}

	if delegated.Nb().Cursor != nil {
		if claimed.Nb().Cursor == nil {
			return failure.FromError(fmt.Errorf("constraint escalation: cursor escalates imposed constraint %q", *delegated.Nb().Cursor))
		} else if *claimed.Nb().Cursor != *delegated.Nb().Cursor {
			return failure.FromError(fmt.Errorf("constraint violation: %q violates imposed cursor constraint %q", *claimed.Nb().Cursor, *delegated.Nb().Cursor))
		}
	}

	if delegated.Nb().Size != nil {
		if claimed.Nb().Size == nil {
			return failure.FromError(fmt.Errorf("constraint escalation: size escalates imposed constraint %d", *delegated.Nb().Size))
		} else if *claimed.Nb().Size != *delegated.Nb().Size {
			return failure.FromError(fmt.Errorf("constraint violation: %d violates imposed size constraint %d", *claimed.Nb().Size, *delegated.Nb().Size))
		}
	}

	return nil
}

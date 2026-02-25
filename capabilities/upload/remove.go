package upload

import (
	"github.com/ipld/go-ipld-prime/datamodel"
	"github.com/storacha/go-ucanto/core/ipld"
	"github.com/storacha/go-ucanto/core/receipt"
	"github.com/storacha/go-ucanto/core/result/failure"
	"github.com/storacha/go-ucanto/core/schema"
	"github.com/storacha/go-ucanto/ucan"
	"github.com/storacha/go-ucanto/validator"

	"github.com/storacha/go-libstoracha/capabilities/types"
)

const RemoveAbility = "upload/remove"

type RemoveCaveats struct {
	Root ipld.Link
}

func (rc RemoveCaveats) ToIPLD() (datamodel.Node, error) {
	return ipld.WrapWithRecovery(&rc, RemoveCaveatsType(), types.Converters...)
}

var RemoveCaveatsReader = schema.Struct[RemoveCaveats](RemoveCaveatsType(), nil, types.Converters...)

type RemoveOk struct {
	Root   ipld.Link
	Shards []ipld.Link
}

func (ro RemoveOk) ToIPLD() (datamodel.Node, error) {
	return ipld.WrapWithRecovery(&ro, RemoveOkType(), types.Converters...)
}

var RemoveOkReader = schema.Struct[RemoveOk](RemoveOkType(), nil, types.Converters...)

type RemoveReceipt receipt.Receipt[RemoveOk, failure.Failure]
type RemoveReceiptReader receipt.ReceiptReader[RemoveOk, failure.Failure]

func NewRemoveReceiptReader() (RemoveReceiptReader, error) {
	return receipt.NewReceiptReader[RemoveOk, failure.Failure](uploadSchema, types.Converters...)
}

var Remove = validator.NewCapability(
	RemoveAbility,
	schema.DIDString(),
	RemoveCaveatsReader,
	func(claimed, delegated ucan.Capability[RemoveCaveats]) failure.Failure {
		if err := validateSpaceDID(claimed.With()); err != nil {
			return err
		}

		if fail := validator.DefaultDerives(claimed, delegated); fail != nil {
			return fail
		}

		// Allow derivation from upload/* wildcard capability
		if delegated.Can() == UploadAbility {
			return nil
		}

		if fail := equalRoot(claimed.Nb().Root, delegated.Nb().Root); fail != nil {
			return fail
		}

		return nil
	},
)

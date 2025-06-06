package upload

import (
	"time"

	"github.com/ipld/go-ipld-prime/datamodel"
	"github.com/storacha/go-ucanto/core/ipld"
	"github.com/storacha/go-ucanto/core/receipt"
	"github.com/storacha/go-ucanto/core/result/failure"
	"github.com/storacha/go-ucanto/core/schema"
	"github.com/storacha/go-ucanto/ucan"
	"github.com/storacha/go-ucanto/validator"

	"github.com/storacha/go-libstoracha/capabilities/types"
)

const GetAbility = "upload/get"

type GetCaveats struct {
	Root ipld.Link
}

func (gc GetCaveats) ToIPLD() (datamodel.Node, error) {
	return ipld.WrapWithRecovery(&gc, GetCaveatsType(), types.Converters...)
}

var GetCaveatsReader = schema.Struct[GetCaveats](GetCaveatsType(), nil, types.Converters...)

type GetOk struct {
	Root       ipld.Link
	Shards     []ipld.Link
	InsertedAt time.Time
	UpdatedAt  time.Time
}

func (ok GetOk) ToIPLD() (datamodel.Node, error) {
	return ipld.WrapWithRecovery(&ok, GetOkType(), types.Converters...)
}

type GetReceipt receipt.Receipt[GetOk, failure.Failure]
type GetReceiptReader receipt.ReceiptReader[GetOk, failure.Failure]

func NewGetReceiptReader() (GetReceiptReader, error) {
	return receipt.NewReceiptReader[GetOk, failure.Failure](uploadSchema)
}

var GetOkReader = schema.Struct[GetOk](GetOkType(), nil, types.Converters...)

var Get = validator.NewCapability(
	GetAbility,
	schema.DIDString(),
	GetCaveatsReader,
	func(claimed, delegated ucan.Capability[GetCaveats]) failure.Failure {
		if err := validateSpaceDID(claimed.With()); err != nil {
			return err
		}

		if fail := validator.DefaultDerives(claimed, delegated); fail != nil {
			return fail
		}

		if fail := equalRoot(claimed.Nb().Root, delegated.Nb().Root); fail != nil {
			return fail
		}
		return nil
	},
)

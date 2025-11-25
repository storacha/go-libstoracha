package consumer

import (
	"github.com/ipld/go-ipld-prime/datamodel"
	"github.com/storacha/go-ucanto/core/ipld"
	"github.com/storacha/go-ucanto/core/receipt"
	"github.com/storacha/go-ucanto/core/result/failure"
	"github.com/storacha/go-ucanto/core/schema"
	"github.com/storacha/go-ucanto/validator"

	"github.com/storacha/go-libstoracha/capabilities/types"
)

const GetAbility = "consumer/get"

type GetCaveats struct {
	Consumer string
}

func (gc GetCaveats) ToIPLD() (datamodel.Node, error) {
	return ipld.WrapWithRecovery(&gc, GetCaveatsType(), types.Converters...)
}

var GetCaveatsReader = schema.Struct[GetCaveats](GetCaveatsType(), nil, types.Converters...)

type GetOk struct {
	Consumer string
	Provider string
}

func (go_ GetOk) ToIPLD() (datamodel.Node, error) {
	return ipld.WrapWithRecovery(&go_, GetOkType(), types.Converters...)
}

type GetReceipt receipt.Receipt[GetOk, failure.Failure]
type GetReceiptReader receipt.ReceiptReader[GetOk, failure.Failure]

func NewGetReceiptReader() (GetReceiptReader, error) {
	return receipt.NewReceiptReader[GetOk, failure.Failure](consumerSchema)
}

var GetOkReader = schema.Struct[GetOk](GetOkType(), nil, types.Converters...)

var Get = validator.NewCapability(
	GetAbility,
	schema.DIDString(),
	GetCaveatsReader,
	validator.DefaultDerives[GetCaveats],
)

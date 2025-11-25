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

const HasAbility = "consumer/has"

type HasCaveats struct {
	Consumer string
}

func (hc HasCaveats) ToIPLD() (datamodel.Node, error) {
	return ipld.WrapWithRecovery(&hc, HasCaveatsType(), types.Converters...)
}

var HasCaveatsReader = schema.Struct[HasCaveats](HasCaveatsType(), nil, types.Converters...)

type HasOk struct {
	Has bool
}

func (ho HasOk) ToIPLD() (datamodel.Node, error) {
	return ipld.WrapWithRecovery(&ho, HasOkType(), types.Converters...)
}

type HasReceipt receipt.Receipt[HasOk, failure.Failure]
type HasReceiptReader receipt.ReceiptReader[HasOk, failure.Failure]

func NewHasReceiptReader() (HasReceiptReader, error) {
	return receipt.NewReceiptReader[HasOk, failure.Failure](consumerSchema)
}

var HasOkReader = schema.Struct[HasOk](HasOkType(), nil, types.Converters...)

var Has = validator.NewCapability(
	HasAbility,
	schema.DIDString(),
	HasCaveatsReader,
	validator.DefaultDerives[HasCaveats],
)

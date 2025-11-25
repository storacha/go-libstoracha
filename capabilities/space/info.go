package space

import (
	"github.com/ipld/go-ipld-prime/datamodel"
	"github.com/storacha/go-ucanto/core/ipld"
	"github.com/storacha/go-ucanto/core/receipt"
	"github.com/storacha/go-ucanto/core/result/failure"
	"github.com/storacha/go-ucanto/core/schema"
	"github.com/storacha/go-ucanto/validator"

	"github.com/storacha/go-libstoracha/capabilities/types"
)

const InfoAbility = "space/info"

// InfoCaveats represents the caveats for space/info (no caveats needed)
type InfoCaveats struct{}

func (ic InfoCaveats) ToIPLD() (datamodel.Node, error) {
	return ipld.WrapWithRecovery(&ic, InfoCaveatsType(), types.Converters...)
}

var InfoCaveatsReader = schema.Struct[InfoCaveats](InfoCaveatsType(), nil, types.Converters...)

type InfoOk struct {
	Did       string
	Providers []string
}

func (io InfoOk) ToIPLD() (datamodel.Node, error) {
	return ipld.WrapWithRecovery(&io, InfoOkType(), types.Converters...)
}

type InfoReceipt receipt.Receipt[InfoOk, failure.Failure]
type InfoReceiptReader receipt.ReceiptReader[InfoOk, failure.Failure]

func NewInfoReceiptReader() (InfoReceiptReader, error) {
	return receipt.NewReceiptReader[InfoOk, failure.Failure](spaceSchema)
}

var InfoOkReader = schema.Struct[InfoOk](InfoOkType(), nil, types.Converters...)

var Info = validator.NewCapability(
	InfoAbility,
	schema.DIDString(),
	InfoCaveatsReader,
	validator.DefaultDerives[InfoCaveats],
)

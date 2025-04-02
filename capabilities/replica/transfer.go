package replica

import (
	"github.com/ipld/go-ipld-prime/datamodel"
	"github.com/storacha/go-ucanto/did"

	"github.com/storacha/go-ucanto/core/ipld"
	"github.com/storacha/go-ucanto/core/schema"
	"github.com/storacha/go-ucanto/ucan"
	"github.com/storacha/go-ucanto/validator"

	"github.com/storacha/go-libstoracha/capabilities/blob"
	"github.com/storacha/go-libstoracha/capabilities/types"
)

const TransferAbility = "replica/transfer"

var _ ipld.Builder = (*TransferCaveats)(nil)

type TransferCaveats struct {
	Space    did.DID
	Blob     blob.Blob
	Location ucan.Link
	Cause    ucan.Link
}

type TransferOk struct {
	Site ucan.Link
	PDP  *ucan.Link
}

func (t TransferOk) ToIPLD() (datamodel.Node, error) {
	return ipld.WrapWithRecovery(&t, TransferOkType(), types.Converters...)
}

func (tc TransferCaveats) ToIPLD() (datamodel.Node, error) {
	return ipld.WrapWithRecovery(&tc, TransferCaveatsType(), types.Converters...)
}

var TransferCaveatsReader = schema.Struct[TransferCaveats](TransferCaveatsType(), nil, types.Converters...)
var Transfer = validator.NewCapability(
	TransferAbility,
	schema.DIDString(),
	TransferCaveatsReader,
	validator.DefaultDerives,
)

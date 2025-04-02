package replica

import (
	"github.com/multiformats/go-multihash"
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
	Blob     blob.Blob
	Location multihash.Multihash
	Cause    ucan.Link
}

func (tc TransferCaveats) ToIPLD() (ipld.Node, error) {
	return ipld.WrapWithRecovery(&tc, TransferCaveatsType(), types.Converters...)
}

var TransferCaveatsReader = schema.Struct[TransferCaveats](TransferCaveatsType(), nil, types.Converters...)
var Transfer = validator.NewCapability(
	TransferAbility,
	schema.DIDString(),
	TransferCaveatsReader,
	validator.DefaultDerives,
)

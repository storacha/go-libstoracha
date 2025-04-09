package blob

import (
	"github.com/ipld/go-ipld-prime/datamodel"
	"github.com/storacha/go-ucanto/core/ipld"
	"github.com/storacha/go-ucanto/core/schema"
	"github.com/storacha/go-ucanto/validator"

	"github.com/storacha/go-libstoracha/capabilities/types"
)

const ReplicateAbility = "blob/replicate"

var _ ipld.Builder = (*ReplicateCaveats)(nil)

type ReplicateCaveats struct {
	Blob     Blob
	Replicas uint
	Location ipld.Link
}

func (rc ReplicateCaveats) ToIPLD() (datamodel.Node, error) {
	return ipld.WrapWithRecovery(&rc, ReplicateCaveatsType(), types.Converters...)
}

var ReplicateCaveatsReader = schema.Struct[ReplicateCaveats](ReplicateCaveatsType(), nil, types.Converters...)
var Replicate = validator.NewCapability(
	ReplicateAbility,
	schema.DIDString(),
	ReplicateCaveatsReader,
	validator.DefaultDerives,
)

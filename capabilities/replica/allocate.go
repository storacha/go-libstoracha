package replica

import (
	"github.com/ipld/go-ipld-prime/datamodel"
	"github.com/storacha/go-ucanto/core/ipld"
	"github.com/storacha/go-ucanto/core/schema"
	"github.com/storacha/go-ucanto/did"
	"github.com/storacha/go-ucanto/ucan"
	"github.com/storacha/go-ucanto/validator"

	"github.com/storacha/go-libstoracha/capabilities/blob"
	"github.com/storacha/go-libstoracha/capabilities/types"
)

const AllocateAbility = "replica/allocate"

var _ ipld.Builder = (*AllocateCaveats)(nil)

type AllocateCaveats struct {
	Space    did.DID
	Blob     blob.Blob
	Location ucan.Link
	Cause    ucan.Link
}

type AllocateOk struct {
	Size uint64
}

func (a AllocateOk) ToIPLD() (ipld.Node, error) {
	return ipld.WrapWithRecovery(&a, AllocateOkType(), types.Converters...)
}

func (ac AllocateCaveats) ToIPLD() (datamodel.Node, error) {
	return ipld.WrapWithRecovery(&ac, AllocateCaveatsType(), types.Converters...)
}

var AllocateCaveatsReader = schema.Struct[AllocateCaveats](AllocateCaveatsType(), nil, types.Converters...)
var Allocate = validator.NewCapability(
	AllocateAbility,
	schema.DIDString(),
	AllocateCaveatsReader,
	validator.DefaultDerives,
)

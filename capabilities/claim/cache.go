package claim

import (
	"github.com/ipld/go-ipld-prime/datamodel"
	"github.com/multiformats/go-multiaddr"
	"github.com/storacha/go-libstoracha/capabilities/types"
	"github.com/storacha/go-ucanto/core/ipld"
	"github.com/storacha/go-ucanto/core/schema"
	"github.com/storacha/go-ucanto/validator"
)

const CacheAbility = "claim/cache"

type Provider struct {
	Addresses []multiaddr.Multiaddr
}

type CacheCaveats struct {
	Claim    ipld.Link
	Provider Provider
}

func (cc CacheCaveats) ToIPLD() (datamodel.Node, error) {
	return ipld.WrapWithRecovery(&cc, CacheCaveatsType(), types.Converters...)
}

var CacheCaveatsReader = schema.Struct[CacheCaveats](CacheCaveatsType(), nil, types.Converters...)

var Cache = validator.NewCapability(CacheAbility, schema.DIDString(), CacheCaveatsReader, nil)

package claim

import (
	// for go:embed
	_ "embed"
	"fmt"

	ipldprime "github.com/ipld/go-ipld-prime"
	"github.com/ipld/go-ipld-prime/datamodel"
	ipldschema "github.com/ipld/go-ipld-prime/schema"
	"github.com/multiformats/go-multiaddr"
	"github.com/storacha/go-capabilities/pkg/types"
	"github.com/storacha/go-ucanto/core/ipld"
	"github.com/storacha/go-ucanto/core/schema"
	"github.com/storacha/go-ucanto/validator"
)

//go:embed claim.ipldsch
var claimSchema []byte

var claimTypeSystem = mustLoadTS()

func mustLoadTS() *ipldschema.TypeSystem {
	ts, err := ipldprime.LoadSchemaBytes(claimSchema)
	if err != nil {
		panic(fmt.Errorf("loading claim schema: %w", err))
	}
	return ts
}

func CacheCaveatsType() ipldschema.Type {
	return claimTypeSystem.TypeByName("CacheCaveats")
}

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

const CacheAbility = "claim/cache"

var CacheCaveatsReader = schema.Struct[CacheCaveats](CacheCaveatsType(), nil, types.Converters...)

var Cache = validator.NewCapability(CacheAbility, schema.DIDString(), CacheCaveatsReader, nil)

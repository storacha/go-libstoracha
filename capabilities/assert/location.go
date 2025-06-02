package assert

import (
	// for schema embed
	_ "embed"
	"net/url"

	"github.com/ipld/go-ipld-prime/datamodel"
	"github.com/storacha/go-libstoracha/capabilities/types"
	"github.com/storacha/go-ucanto/core/ipld"
	"github.com/storacha/go-ucanto/core/result/failure"
	"github.com/storacha/go-ucanto/core/schema"
	"github.com/storacha/go-ucanto/did"
	"github.com/storacha/go-ucanto/validator"
)

const LocationAbility = "assert/location"

type Range struct {
	Offset uint64
	Length *uint64
}

type LocationCaveats struct {
	Content  types.HasMultihash
	Location []url.URL
	Range    *Range
	Space    did.DID
}

// Space field used to be optional, so we maintain compatibility
type LegacyLocationCaveats struct {
	Content  types.HasMultihash
	Location []url.URL
	Range    *Range
	Space    *did.DID
}

func (lc LocationCaveats) ToIPLD() (datamodel.Node, error) {
	space := &lc.Space
	if lc.Space == did.Undef {
		space = nil
	}
	return ipld.WrapWithRecovery(&LegacyLocationCaveats{
		lc.Content, lc.Location, lc.Range, space,
	}, LocationCaveatsType(), types.Converters...)
}

var LocationCaveatsReader = schema.Mapped(schema.Struct[LegacyLocationCaveats](LocationCaveatsType(), nil, types.Converters...),
	func(llc LegacyLocationCaveats) (LocationCaveats, failure.Failure) {
		space := did.Undef
		if llc.Space != nil {
			space = *llc.Space
		}
		return LocationCaveats{llc.Content, llc.Location, llc.Range, space}, nil
	})

var Location = validator.NewCapability(LocationAbility, schema.DIDString(), LocationCaveatsReader, nil)

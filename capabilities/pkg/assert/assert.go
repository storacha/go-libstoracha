package assert

import (
	"fmt"
	"net/url"

	// for schema embed
	_ "embed"

	ipldprime "github.com/ipld/go-ipld-prime"
	"github.com/ipld/go-ipld-prime/datamodel"
	ipldschema "github.com/ipld/go-ipld-prime/schema"
	"github.com/storacha/go-capabilities/pkg/types"
	"github.com/storacha/go-ucanto/core/ipld"
	"github.com/storacha/go-ucanto/core/result/failure"
	"github.com/storacha/go-ucanto/core/schema"
	"github.com/storacha/go-ucanto/did"
	"github.com/storacha/go-ucanto/validator"
)

//go:embed assert.ipldsch
var assertSchema []byte

var assertTS = mustLoadTS()

func mustLoadTS() *ipldschema.TypeSystem {
	ts, err := ipldprime.LoadSchemaBytes(assertSchema)
	if err != nil {
		panic(fmt.Errorf("loading assert schema: %w", err))
	}
	return ts
}

func LocationCaveatsType() ipldschema.Type {
	return assertTS.TypeByName("LocationCaveats")
}

func InclusionCaveatsType() ipldschema.Type {
	return assertTS.TypeByName("InclusionCaveats")
}

func IndexCaveatsType() ipldschema.Type {
	return assertTS.TypeByName("IndexCaveats")
}

func PartitionCaveatsType() ipldschema.Type {
	return assertTS.TypeByName("PartitionCaveats")
}

func RelationCaveatsType() ipldschema.Type {
	return assertTS.TypeByName("RelationCaveats")
}

func EqualsCaveatsType() ipldschema.Type {
	return assertTS.TypeByName("EqualsCaveats")
}

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

const LocationAbility = "assert/location"

var LocationCaveatsReader = schema.Mapped(schema.Struct[LegacyLocationCaveats](LocationCaveatsType(), nil, types.Converters...),
	func(llc LegacyLocationCaveats) (LocationCaveats, failure.Failure) {
		space := did.Undef
		if llc.Space != nil {
			space = *llc.Space
		}
		return LocationCaveats{llc.Content, llc.Location, llc.Range, space}, nil
	})

var Location = validator.NewCapability(LocationAbility, schema.DIDString(), LocationCaveatsReader, nil)

/**
 * Claims that a CID includes the contents claimed in another CID.
 */

type InclusionCaveats struct {
	Content  types.HasMultihash
	Includes ipld.Link
	Proof    *ipld.Link
}

func (ic InclusionCaveats) ToIPLD() (datamodel.Node, error) {
	return ipld.WrapWithRecovery(&ic, InclusionCaveatsType(), types.Converters...)
}

const InclusionAbility = "assert/inclusion"

var InclusionCaveatsReader = schema.Struct[InclusionCaveats](InclusionCaveatsType(), nil, types.Converters...)

var Inclusion = validator.NewCapability(InclusionAbility, schema.DIDString(),
	InclusionCaveatsReader, nil)

/**
 * Claims that a content graph can be found in blob(s) that are identified and
 * indexed in the given index CID.
 */

type IndexCaveats struct {
	Content ipld.Link
	Index   ipld.Link
}

func (ic IndexCaveats) ToIPLD() (datamodel.Node, error) {
	return ipld.WrapWithRecovery(&ic, IndexCaveatsType(), types.Converters...)
}

const IndexAbility = "assert/index"

var IndexCaveatsReader = schema.Struct[IndexCaveats](IndexCaveatsType(), nil, types.Converters...)

var Index = validator.NewCapability(IndexAbility, schema.DIDString(), IndexCaveatsReader, nil)

/**
 * Claims that a CID's graph can be read from the blocks found in parts.
 */

type PartitionCaveats struct {
	Content types.HasMultihash
	Blocks  *ipld.Link
	Parts   []ipld.Link
}

func (pc PartitionCaveats) ToIPLD() (datamodel.Node, error) {
	return ipld.WrapWithRecovery(&pc, PartitionCaveatsType(), types.Converters...)
}

const PartitionAbility = "assert/partition"

var PartitionCaveatsReader = schema.Struct[PartitionCaveats](PartitionCaveatsType(), nil, types.Converters...)
var Partition = validator.NewCapability(
	PartitionAbility,
	schema.DIDString(),
	PartitionCaveatsReader, nil)

/**
 * Claims that a CID links to other CIDs.
 */

type RelationPartInclusion struct {
	Content ipld.Link
	Parts   *[]ipld.Link
}

type RelationPart struct {
	Content  ipld.Link
	Includes *RelationPartInclusion
}

type RelationCaveats struct {
	Content  types.HasMultihash
	Children []ipld.Link
	Parts    []RelationPart
}

func (rc RelationCaveats) ToIPLD() (datamodel.Node, error) {
	return ipld.WrapWithRecovery(&rc, RelationCaveatsType(), types.Converters...)
}

const RelationAbility = "assert/relation"

var RelationCaveatsReader = schema.Struct[RelationCaveats](RelationCaveatsType(), nil, types.Converters...)

var Relation = validator.NewCapability(
	RelationAbility,
	schema.DIDString(),
	RelationCaveatsReader,
	nil,
)

/**
 * Claim data is referred to by another CID and/or multihash. e.g CAR CID & CommP CID
 */

type EqualsCaveats struct {
	Content types.HasMultihash
	Equals  ipld.Link
}

func (ec EqualsCaveats) ToIPLD() (datamodel.Node, error) {
	return ipld.WrapWithRecovery(&ec, EqualsCaveatsType(), types.Converters...)
}

const EqualsAbility = "assert/equals"

var EqualsCaveatsReader = schema.Struct[EqualsCaveats](EqualsCaveatsType(), nil, types.Converters...)

var Equals = validator.NewCapability(
	EqualsAbility,
	schema.DIDString(),
	EqualsCaveatsReader,
	nil,
)

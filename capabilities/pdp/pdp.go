package pdp

import (
	// for go:embed
	_ "embed"
	"fmt"

	"github.com/filecoin-project/go-data-segment/merkletree"
	ipldprime "github.com/ipld/go-ipld-prime"
	"github.com/ipld/go-ipld-prime/datamodel"
	ipldschema "github.com/ipld/go-ipld-prime/schema"
	"github.com/storacha/go-libstoracha/capabilities/types"
	"github.com/storacha/go-libstoracha/piece/piece"
	"github.com/storacha/go-ucanto/core/ipld"
	"github.com/storacha/go-ucanto/core/schema"
	"github.com/storacha/go-ucanto/validator"
)

const AcceptAbility = "pdp/accept"

//go:embed pdp.ipldsch
var pdpSchema []byte

var pdpTS = mustLoadTS()

func mustLoadTS() *ipldschema.TypeSystem {
	ts, err := ipldprime.LoadSchemaBytes(pdpSchema)
	if err != nil {
		panic(fmt.Errorf("loading blob schema: %w", err))
	}
	return ts
}

func AcceptCaveatsType() ipldschema.Type {
	return pdpTS.TypeByName("AcceptCaveats")
}

func AcceptOkType() ipldschema.Type {
	return pdpTS.TypeByName("AcceptOk")
}

func InfoCaveatsType() ipldschema.Type {
	return pdpTS.TypeByName("InfoCaveats")
}

func InfoOkType() ipldschema.Type {
	return pdpTS.TypeByName("InfoOk")
}

type AcceptCaveats struct {
	Piece piece.PieceLink
}

func (ac AcceptCaveats) ToIPLD() (datamodel.Node, error) {
	return ipld.WrapWithRecovery(&ac, AcceptCaveatsType(), types.Converters...)
}

var AcceptCaveatsReader = schema.Struct[AcceptCaveats](AcceptCaveatsType(), nil, types.Converters...)

var Accept = validator.NewCapability(
	AcceptAbility,
	schema.DIDString(),
	AcceptCaveatsReader,
	validator.DefaultDerives,
)

type AcceptOk struct {
	Aggregate      piece.PieceLink
	InclusionProof merkletree.ProofData
	Piece          piece.PieceLink
}

func (ao AcceptOk) ToIPLD() (datamodel.Node, error) {
	return ipld.WrapWithRecovery(&ao, AcceptOkType(), types.Converters...)
}

const InfoAbility = "pdp/info"

type InfoCaveats struct {
	Piece piece.PieceLink
}

func (ic InfoCaveats) ToIPLD() (datamodel.Node, error) {
	return ipld.WrapWithRecovery(&ic, InfoCaveatsType(), types.Converters...)
}

var InfoCaveatsReader = schema.Struct[InfoCaveats](InfoCaveatsType(), nil, types.Converters...)

var Info = validator.NewCapability(
	InfoAbility,
	schema.DIDString(),
	InfoCaveatsReader,
	validator.DefaultDerives,
)

type InfoAcceptedAggregate struct {
	Aggregate      piece.PieceLink
	InclusionProof merkletree.ProofData
}

type InfoOk struct {
	Piece      piece.PieceLink
	Aggregates []InfoAcceptedAggregate
}

func (io InfoOk) ToIPLD() (datamodel.Node, error) {
	return ipld.WrapWithRecovery(&io, InfoOkType(), types.Converters...)
}

package filecoin

import (
	"github.com/ipld/go-ipld-prime/datamodel"
	"github.com/storacha/go-libstoracha/capabilities/types"
	"github.com/storacha/go-ucanto/core/ipld"
	"github.com/storacha/go-ucanto/core/result/failure"
	"github.com/storacha/go-ucanto/core/schema"
	"github.com/storacha/go-ucanto/ucan"
	"github.com/storacha/go-ucanto/validator"
)

const (
	InfoAbility = "filecoin/info"
)

type ProofData []byte

type InclusionProof struct {
	Subtree ProofData `ipld:"subtree"`
	Index   ProofData `ipld:"index"`
}

type SingletonMarketSource struct {
	DealID uint64 `ipld:"dealID"`
}

type DealMetadata struct {
	DataType   uint64                `ipld:"dataType"`
	DataSource SingletonMarketSource `ipld:"dataSource"`
}

type InfoCaveats struct {
	Piece datamodel.Link `ipld:"piece"`
}

func (ic InfoCaveats) ToIPLD() (datamodel.Node, error) {
	return ipld.WrapWithRecovery(&ic, InfoCaveatsType(), types.Converters...)
}

var InfoCaveatsReader = schema.Struct[InfoCaveats](InfoCaveatsType(), nil, types.Converters...)

type InfoAcceptedAggregate struct {
	Aggregate datamodel.Link `ipld:"aggregate"`
	Inclusion InclusionProof `ipld:"inclusion"`
}

type InfoAcceptedDeal struct {
	Aggregate datamodel.Link `ipld:"aggregate"`
	Aux       DealMetadata   `ipld:"aux"`
	Provider  string         `ipld:"provider"`
}

type InfoOk struct {
	Piece      datamodel.Link          `ipld:"piece"`
	Aggregates []InfoAcceptedAggregate `ipld:"aggregates"`
	Deals      []InfoAcceptedDeal      `ipld:"deals"`
}

func (io InfoOk) ToIPLD() (datamodel.Node, error) {
	return ipld.WrapWithRecovery(&io, InfoOkType(), types.Converters...)
}

var InfoOkReader = schema.Struct[InfoOk](InfoOkType(), nil, types.Converters...)

// Info is a capability that allows an agent to _request_ info about a content piece in Filecoin deals.
var Info = validator.NewCapability(
	InfoAbility,
	schema.DIDString(),
	InfoCaveatsReader,
	func(claimed, delegated ucan.Capability[InfoCaveats]) failure.Failure {
		return validateFilecoinCapability(claimed, delegated, func(claimedNb, delegatedNb InfoCaveats) failure.Failure {
			return equalPieceLink(claimedNb.Piece, delegatedNb.Piece)
		})
	},
)
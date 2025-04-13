package filecoin_test

import (
	"testing"

	"github.com/ipfs/go-cid"
	cidlink "github.com/ipld/go-ipld-prime/linking/cid"
	"github.com/storacha/go-libstoracha/capabilities/filecoin"
	"github.com/stretchr/testify/require"
)

func TestInfoCapability(t *testing.T) {
	capability := filecoin.Info

	t.Run("has correct ability", func(t *testing.T) {
		require.Equal(t, "filecoin/info", capability.Can())
	})
}

func TestInfoCaveatsRoundTrip(t *testing.T) {
	pieceCid, err := cid.Parse("bafybeihykhetgzaibu2vkbzycmhjvuahgk7yb3p5d7sh6d6ze4mhnnjaga")
	require.NoError(t, err)
	pieceLink := cidlink.Link{Cid: pieceCid}
	
	nb := filecoin.InfoCaveats{
		Piece: pieceLink,
	}

	node, err := nb.ToIPLD()
	require.NoError(t, err)

	rnb, err := filecoin.InfoCaveatsReader.Read(node)
	require.NoError(t, err)
	require.Equal(t, nb.Piece.String(), rnb.Piece.String())
}

func TestInfoOkRoundTrip(t *testing.T) {
	pieceCid, err := cid.Parse("bafybeihykhetgzaibu2vkbzycmhjvuahgk7yb3p5d7sh6d6ze4mhnnjaga")
	require.NoError(t, err)
	pieceLink := cidlink.Link{Cid: pieceCid}
	
	aggregateCid, err := cid.Parse("bafybeid46f7zggioxjm5p2ze2l6s6wbqvoo4gzbdzuibgwbhe5iopu2aiy")
	require.NoError(t, err)
	aggregateLink := cidlink.Link{Cid: aggregateCid}
	
	ok := filecoin.InfoOk{
		Piece: pieceLink,
		Aggregates: []filecoin.InfoAcceptedAggregate{
			{
				Aggregate: aggregateLink,
				Inclusion: filecoin.InclusionProof{
					Subtree: []byte{1, 2, 3},
					Index:   []byte{4, 5, 6},
				},
			},
		},
		Deals: []filecoin.InfoAcceptedDeal{
			{
				Aggregate: aggregateLink,
				Aux: filecoin.DealMetadata{
					DataType: 1,
					DataSource: filecoin.SingletonMarketSource{
						DealID: 12345,
					},
				},
				Provider: "f01234",
			},
		},
	}

	node, err := ok.ToIPLD()
	require.NoError(t, err)

	rok, err := filecoin.InfoOkReader.Read(node)
	require.NoError(t, err)
	require.Equal(t, ok.Piece.String(), rok.Piece.String())
	require.Len(t, rok.Aggregates, 1)
	require.Equal(t, ok.Aggregates[0].Aggregate.String(), rok.Aggregates[0].Aggregate.String())
	require.Len(t, rok.Deals, 1)
	require.Equal(t, ok.Deals[0].Provider, rok.Deals[0].Provider)
}
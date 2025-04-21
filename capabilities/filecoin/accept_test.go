package filecoin_test

import (
	"testing"

	"github.com/ipfs/go-cid"
	cidlink "github.com/ipld/go-ipld-prime/linking/cid"
	"github.com/storacha/go-libstoracha/capabilities/filecoin"
	"github.com/stretchr/testify/require"
)

func TestAcceptCapability(t *testing.T) {
	capability := filecoin.Accept

	t.Run("has correct ability", func(t *testing.T) {
		require.Equal(t, "filecoin/accept", capability.Can())
	})
}

func TestAcceptCaveatsRoundTrip(t *testing.T) {
	contentCid, err := cid.Parse("bafybeigdyrzt5sfp7udm7hu76uh7y26nf3efuylqabf3oclgtqy55fbzdi")
	require.NoError(t, err)
	contentLink := cidlink.Link{Cid: contentCid}

	pieceCid, err := cid.Parse("bafybeihykhetgzaibu2vkbzycmhjvuahgk7yb3p5d7sh6d6ze4mhnnjaga")
	require.NoError(t, err)
	pieceLink := cidlink.Link{Cid: pieceCid}

	nb := filecoin.AcceptCaveats{
		Content: contentLink,
		Piece:   pieceLink,
	}

	node, err := nb.ToIPLD()
	require.NoError(t, err)

	rnb, err := filecoin.AcceptCaveatsReader.Read(node)
	require.NoError(t, err)
	require.Equal(t, nb.Content.String(), rnb.Content.String())
	require.Equal(t, nb.Piece.String(), rnb.Piece.String())
}

func TestAcceptOkRoundTrip(t *testing.T) {
	pieceCid, err := cid.Parse("bafybeihykhetgzaibu2vkbzycmhjvuahgk7yb3p5d7sh6d6ze4mhnnjaga")
	require.NoError(t, err)
	pieceLink := cidlink.Link{Cid: pieceCid}

	aggregateCid, err := cid.Parse("bafybeid46f7zggioxjm5p2ze2l6s6wbqvoo4gzbdzuibgwbhe5iopu2aiy")
	require.NoError(t, err)
	aggregateLink := cidlink.Link{Cid: aggregateCid}

	ok := filecoin.AcceptOk{
		Piece:     pieceLink,
		Aggregate: aggregateLink,
		Inclusion: filecoin.InclusionProof{
			Subtree: []byte{1, 2, 3},
			Index:   []byte{4, 5, 6},
		},
		Aux: filecoin.DealMetadata{
			DataType: 1,
			DataSource: filecoin.SingletonMarketSource{
				DealID: 12345,
			},
		},
	}

	node, err := ok.ToIPLD()
	require.NoError(t, err)

	rok, err := filecoin.AcceptOkReader.Read(node)
	require.NoError(t, err)
	require.Equal(t, ok.Piece.String(), rok.Piece.String())
	require.Equal(t, ok.Aggregate.String(), rok.Aggregate.String())
	require.Equal(t, ok.Aux.DataType, rok.Aux.DataType)
	require.Equal(t, ok.Aux.DataSource.DealID, rok.Aux.DataSource.DealID)
}

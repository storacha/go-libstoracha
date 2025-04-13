package filecoin_test

import (
	"testing"

	"github.com/ipfs/go-cid"
	cidlink "github.com/ipld/go-ipld-prime/linking/cid"
	"github.com/storacha/go-libstoracha/capabilities/filecoin"
	"github.com/stretchr/testify/require"
)

func TestOfferCapability(t *testing.T) {
	capability := filecoin.Offer

	t.Run("has correct ability", func(t *testing.T) {
		require.Equal(t, "filecoin/offer", capability.Can())
	})
}

func TestOfferCaveatsRoundTrip(t *testing.T) {
	contentCid, err := cid.Parse("bafybeigdyrzt5sfp7udm7hu76uh7y26nf3efuylqabf3oclgtqy55fbzdi")
	require.NoError(t, err)
	contentLink := cidlink.Link{Cid: contentCid}

	pieceCid, err := cid.Parse("bafybeihykhetgzaibu2vkbzycmhjvuahgk7yb3p5d7sh6d6ze4mhnnjaga")
	require.NoError(t, err)
	pieceLink := cidlink.Link{Cid: pieceCid}
	
	nb := filecoin.OfferCaveats{
		Content: contentLink,
		Piece:   pieceLink,
	}

	node, err := nb.ToIPLD()
	require.NoError(t, err)

	rnb, err := filecoin.OfferCaveatsReader.Read(node)
	require.NoError(t, err)
	require.Equal(t, nb.Content.String(), rnb.Content.String())
	require.Equal(t, nb.Piece.String(), rnb.Piece.String())
}

func TestOfferOkRoundTrip(t *testing.T) {
	pieceCid, err := cid.Parse("bafybeihykhetgzaibu2vkbzycmhjvuahgk7yb3p5d7sh6d6ze4mhnnjaga")
	require.NoError(t, err)
	pieceLink := cidlink.Link{Cid: pieceCid}
	
	ok := filecoin.OfferOk{
		Piece: pieceLink,
	}

	node, err := ok.ToIPLD()
	require.NoError(t, err)

	rok, err := filecoin.OfferOkReader.Read(node)
	require.NoError(t, err)
	require.Equal(t, ok.Piece.String(), rok.Piece.String())
}
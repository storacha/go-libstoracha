package pdp_test

import (
	"testing"

	"github.com/filecoin-project/go-data-segment/merkletree"
	"github.com/storacha/go-libstoracha/capabilities/pdp"
	"github.com/storacha/go-libstoracha/testutil"
	"github.com/stretchr/testify/require"
)

func TestRoundTripAcceptCaveats(t *testing.T) {
	bytes := testutil.RandomBytes(t, 256)
	digest := testutil.MultihashFromBytes(t, bytes)

	ac := pdp.AcceptCaveats{
		Blob: digest,
	}

	node, err := ac.ToIPLD()
	require.NoError(t, err)

	rac, err := pdp.AcceptCaveatsReader.Read(node)
	require.NoError(t, err)
	require.Equal(t, ac, rac)
}

func TestRoundTripAcceptOk(t *testing.T) {
	piece := testutil.RandomPiece(t, 1024)
	aggregate := testutil.RandomPiece(t, 2048)

	// Create some test proof nodes
	node1 := merkletree.ZeroCommitmentForLevel(0)
	node2 := merkletree.ZeroCommitmentForLevel(1)

	inclusionProof := merkletree.ProofData{
		Path:  []merkletree.Node{node1, node2},
		Index: 42,
	}

	ao := pdp.AcceptOk{
		Piece:          piece,
		Aggregate:      aggregate,
		InclusionProof: inclusionProof,
	}

	node, err := ao.ToIPLD()
	require.NoError(t, err)

	rao, err := pdp.AcceptOkReader.Read(node)
	require.NoError(t, err)
	require.Equal(t, ao.Piece, rao.Piece)
	require.Equal(t, ao.Aggregate, rao.Aggregate)
	require.Equal(t, ao.InclusionProof, rao.InclusionProof)
}

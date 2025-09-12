package content_test

import (
	"testing"

	"github.com/storacha/go-libstoracha/capabilities/space/content"
	"github.com/storacha/go-libstoracha/testutil"

	"github.com/stretchr/testify/require"
)

func TestRoundTripRetrieveCaveats(t *testing.T) {
	digest := testutil.RandomMultihash(t)
	nb := content.RetrieveCaveats{
		Blob: content.BlobDigest{
			Digest: digest,
		},
		Range: content.Range{
			Start: 123,
			End:   456,
		},
	}

	node, err := nb.ToIPLD()
	require.NoError(t, err)

	rnb, err := content.RetrieveCaveatsReader.Read(node)
	require.NoError(t, err)
	require.Equal(t, nb, rnb)
}

func TestRoundTripRetrieveOk(t *testing.T) {
	ok := content.RetrieveOk{}

	node, err := ok.ToIPLD()
	require.NoError(t, err)

	rok, err := content.RetrieveOkReader.Read(node)
	require.NoError(t, err)
	require.Equal(t, ok, rok)
}

package blob_test

import (
	"testing"

	"github.com/storacha/go-libstoracha/capabilities/internal/testutil"
	"github.com/storacha/go-libstoracha/capabilities/space/blob"
	"github.com/stretchr/testify/require"
)

func TestRoundTripAddCaveats(t *testing.T) {
	digest, bytes := testutil.RandomBytes(t, 256)
	nb := blob.AddCaveats{
		Blob: blob.Blob{
			Digest: digest,
			Size:   uint64(len(bytes)),
		},
	}

	node, err := nb.ToIPLD()
	require.NoError(t, err)

	rnb, err := blob.AddCaveatsReader.Read(node)
	require.NoError(t, err)
	require.Equal(t, nb, rnb)
}

func TestRoundTripAddOk(t *testing.T) {
	site := blob.Promise{
		UcanAwait: blob.Await{
			Selector: ".out.ok",
			Link:     testutil.RandomCID(t),
		},
	}
	ok := blob.AddOk{
		Site: site,
	}

	node, err := ok.ToIPLD()
	require.NoError(t, err)

	rok, err := blob.AddOkReader.Read(node)
	require.NoError(t, err)
	require.Equal(t, ok, rok)
}

package blob_test

import (
	"testing"

	"github.com/storacha/go-libstoracha/capabilities/types"
	"github.com/storacha/go-libstoracha/capabilities/web3.storage/blob"
	"github.com/storacha/go-libstoracha/testutil"
	"github.com/stretchr/testify/require"
)

func TestRoundTripAcceptCaveats(t *testing.T) {
	bytes := testutil.RandomBytes(t, 256)
	digest := testutil.MultihashFromBytes(t, bytes)
	nb := blob.AcceptCaveats{
		Space: testutil.RandomPrincipal(t).DID(),
		Blob: types.Blob{
			Digest: digest,
			Size:   uint64(len(bytes)),
		},
		TTL: 3600,
		Put: types.Promise{
			UcanAwait: types.Await{
				Selector: ".out.ok",
				Link:     testutil.RandomCID(t),
			},
		},
	}

	node, err := nb.ToIPLD()
	require.NoError(t, err)

	rnb, err := blob.AcceptCaveatsReader.Read(node)
	require.NoError(t, err)
	require.Equal(t, nb, rnb)
}

func TestRoundTripAcceptOk(t *testing.T) {
	ok := blob.AcceptOk{
		Site: testutil.RandomCID(t),
	}

	node, err := ok.ToIPLD()
	require.NoError(t, err)

	rok, err := blob.AcceptOkReader.Read(node)
	require.NoError(t, err)
	require.Equal(t, ok, rok)
}

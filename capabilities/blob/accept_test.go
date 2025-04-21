package blob_test

import (
	"testing"

	"github.com/storacha/go-libstoracha/capabilities/blob"
	"github.com/storacha/go-libstoracha/capabilities/internal/testutil"
	"github.com/storacha/go-libstoracha/capabilities/types"

	"github.com/stretchr/testify/require"
)

func TestRoundTripAcceptCaveats(t *testing.T) {
	digest, bytes := testutil.RandomBytes(t, 256)
	nb := blob.AcceptCaveats{
		Space: testutil.RandomPrincipal(t).DID(),
		Blob: types.Blob{
			Digest: digest,
			Size:   uint64(len(bytes)),
		},
		Put: blob.Promise{
			UcanAwait: blob.Await{
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

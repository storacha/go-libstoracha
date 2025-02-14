package blob

import (
	"testing"

	"github.com/storacha/go-libstoracha/capabilities/pkg/internal/testutil"
	"github.com/stretchr/testify/require"
)

func TestRoundTripAcceptCaveats(t *testing.T) {
	digest, bytes := testutil.RandomBytes(t, 256)
	nb := AcceptCaveats{
		Space: testutil.RandomPrincipal(t).DID(),
		Blob: Blob{
			Digest: digest,
			Size:   uint64(len(bytes)),
		},
		Put: Promise{
			UcanAwait: Await{
				Selector: ".out.ok",
				Link:     testutil.RandomCID(t),
			},
		},
	}

	node, err := nb.ToIPLD()
	require.NoError(t, err)

	rnb, err := AcceptCaveatsReader.Read(node)
	require.NoError(t, err)
	require.Equal(t, nb, rnb)
}

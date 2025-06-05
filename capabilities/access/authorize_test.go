package access_test

import (
	"testing"

	"github.com/storacha/go-libstoracha/capabilities/access"
	"github.com/storacha/go-libstoracha/internal/testutil"
	"github.com/storacha/go-ucanto/did"
	"github.com/stretchr/testify/require"
)

func TestRoundTripAuthorizeCaveats(t *testing.T) {
	issuer, err := did.Parse("did:mailto:example.com:alice")
	require.NoError(t, err)
	nb := access.AuthorizeCaveats{
		Iss: &issuer,
		Att: []access.CapabilityRequest{
			{Can: "stuff/do"},
			{Can: "stuff/say"},
		},
	}

	node, err := nb.ToIPLD()
	require.NoError(t, err)

	rnb, err := access.AuthorizeCaveatsReader.Read(node)
	require.NoError(t, err)
	require.Equal(t, nb, rnb)
}

func TestRoundTripAuthorizeOk(t *testing.T) {
	ok := access.AuthorizeOk{
		Request:    testutil.RandomCID(t),
		Expiration: 1234,
	}

	node, err := ok.ToIPLD()
	require.NoError(t, err)

	rok, err := access.AuthorizeOkReader.Read(node)
	require.NoError(t, err)
	require.Equal(t, ok, rok)
}

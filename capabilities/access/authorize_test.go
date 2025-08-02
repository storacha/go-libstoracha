package access_test

import (
	"testing"

	"github.com/storacha/go-libstoracha/capabilities/access"
	"github.com/storacha/go-libstoracha/testutil"
	"github.com/storacha/go-ucanto/ucan"
	"github.com/stretchr/testify/require"
)

func TestRoundTripAuthorizeCaveats(t *testing.T) {
	alice := "did:mailto:example.com:alice"
	nb := access.AuthorizeCaveats{
		Iss: &alice,
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

func TestAuthorizeDerive(t *testing.T) {
	alice := "did:mailto:example.com:alice"

	delegated := ucan.NewCapability(
		access.AuthorizeAbility,
		"did:mailto:example.com:alice",
		access.AuthorizeCaveats{
			Iss: &alice,
			Att: []access.CapabilityRequest{
				{Can: "stuff/do"},
				{Can: "stuff/say"},
			},
		},
	)

	t.Run("accepts an identical capability", func(t *testing.T) {
		claimed := ucan.NewCapability(
			delegated.Can(),
			delegated.With(),
			delegated.Nb(),
		)

		fail := access.AuthorizeDerive(claimed, delegated)
		require.NoError(t, fail)
	})

	t.Run("rejects the wrong resource", func(t *testing.T) {
		claimed := ucan.NewCapability(
			delegated.Can(),
			"did:mailto:example.com:bob",
			delegated.Nb(),
		)

		fail := access.AuthorizeDerive(claimed, delegated)
		require.Error(t, fail)
	})

	t.Run("rejects the wrong issuer", func(t *testing.T) {
		bob := "did:mailto:example.com:bob"

		claimed := ucan.NewCapability(
			delegated.Can(),
			delegated.With(),
			access.AuthorizeCaveats{
				Iss: &bob,
				Att: delegated.Nb().Att,
			},
		)

		fail := access.AuthorizeDerive(claimed, delegated)
		require.Error(t, fail)
	})

	t.Run("rejects non-delegated capabilities", func(t *testing.T) {
		claimed := ucan.NewCapability(
			delegated.Can(),
			delegated.With(),
			access.AuthorizeCaveats{
				Iss: delegated.Nb().Iss,
				Att: []access.CapabilityRequest{
					{Can: "stuff/do"},
					{Can: "stuff/yell"},
				},
			},
		)

		fail := access.AuthorizeDerive(claimed, delegated)
		require.Error(t, fail)
	})

	t.Run("accepts a subset of delegated capabilities", func(t *testing.T) {
		claimed := ucan.NewCapability(
			delegated.Can(),
			delegated.With(),
			access.AuthorizeCaveats{
				Iss: delegated.Nb().Iss,
				Att: []access.CapabilityRequest{
					{Can: "stuff/do"},
				},
			},
		)

		fail := access.AuthorizeDerive(claimed, delegated)
		require.NoError(t, fail)
	})

	t.Run("accepts any capabilities when wildcard is delegated", func(t *testing.T) {
		wildcardDelegated := ucan.NewCapability(
			delegated.Can(),
			delegated.With(),
			access.AuthorizeCaveats{
				Iss: delegated.Nb().Iss,
				Att: []access.CapabilityRequest{
					{Can: "*"},
				},
			},
		)

		claimed := ucan.NewCapability(
			delegated.Can(),
			delegated.With(),
			access.AuthorizeCaveats{
				Iss: delegated.Nb().Iss,
				Att: []access.CapabilityRequest{
					{Can: "stuff/do"},
					{Can: "stuff/yell"},
				},
			},
		)

		fail := access.AuthorizeDerive(claimed, wildcardDelegated)
		require.NoError(t, fail)
	})
}

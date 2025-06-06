package access_test

import (
	"testing"

	"github.com/ipfs/go-cid"
	"github.com/storacha/go-libstoracha/capabilities/access"
	"github.com/storacha/go-libstoracha/internal/testutil"
	"github.com/storacha/go-ucanto/core/ipld/codec/cbor"
	"github.com/storacha/go-ucanto/did"
	"github.com/storacha/go-ucanto/ucan"
	"github.com/stretchr/testify/require"
)

func TestRoundTripConfirmCaveats(t *testing.T) {
	alice := testutil.Must(did.Parse("did:mailto:example.com:alice"))(t)
	bob := testutil.Must(did.Parse("did:mailto:example.com:bob"))(t)

	nb := access.ConfirmCaveats{
		Cause: testutil.RandomCID(t),
		Iss:   alice,
		Aud:   bob,
		Att: []access.CapabilityRequest{
			{Can: "stuff/do"},
			{Can: "stuff/say"},
		},
	}

	node, err := nb.ToIPLD()
	require.NoError(t, err)

	rnb, err := access.ConfirmCaveatsReader.Read(node)
	require.NoError(t, err)
	require.Equal(t, nb, rnb)
}

func TestRoundTripConfirmOk(t *testing.T) {
	delegations := access.DelegationsModel{
		Keys:   make([]string, 0, 2),
		Values: make(map[string][]byte),
	}
	for range 2 {
		mh, bytes := testutil.RandomBytes(t, 256)
		delegations.Keys = append(delegations.Keys, cid.NewCidV1(cbor.Code, mh).String())
		delegations.Values[cid.NewCidV1(cbor.Code, mh).String()] = bytes
	}
	ok := access.ConfirmOk{
		Delegations: delegations,
	}

	node, err := ok.ToIPLD()
	require.NoError(t, err)

	rok, err := access.ConfirmOkReader.Read(node)
	require.NoError(t, err)
	require.Equal(t, ok, rok)
}

func TestConfirmDerive(t *testing.T) {
	alice := testutil.Must(did.Parse("did:mailto:example.com:alice"))(t)
	bob := testutil.Must(did.Parse("did:mailto:example.com:bob"))(t)

	delegated := ucan.NewCapability(
		access.ConfirmAbility,
		"did:mailto:example.com:alice",
		access.ConfirmCaveats{
			Cause: testutil.RandomCID(t),
			Iss:   alice,
			Aud:   bob,
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

		fail := access.ConfirmDerive(claimed, delegated)
		require.NoError(t, fail)
	})

	t.Run("rejects the wrong resource", func(t *testing.T) {
		claimed := ucan.NewCapability(
			delegated.Can(),
			"did:mailto:example.com:bob",
			delegated.Nb(),
		)

		fail := access.ConfirmDerive(claimed, delegated)
		require.Error(t, fail)
	})

	t.Run("rejects the wrong issuer", func(t *testing.T) {
		claimed := ucan.NewCapability(
			delegated.Can(),
			delegated.With(),
			access.ConfirmCaveats{
				Cause: delegated.Nb().Cause,
				Iss:   bob,
				Aud:   delegated.Nb().Aud,
				Att:   delegated.Nb().Att,
			},
		)

		fail := access.ConfirmDerive(claimed, delegated)
		require.Error(t, fail)
	})

	t.Run("rejects the wrong audience", func(t *testing.T) {
		claimed := ucan.NewCapability(
			delegated.Can(),
			delegated.With(),
			access.ConfirmCaveats{
				Cause: delegated.Nb().Cause,
				Iss:   delegated.Nb().Iss,
				Aud:   alice,
				Att:   delegated.Nb().Att,
			},
		)

		fail := access.ConfirmDerive(claimed, delegated)
		require.Error(t, fail)
	})

	t.Run("rejects non-delegated capabilities", func(t *testing.T) {
		claimed := ucan.NewCapability(
			delegated.Can(),
			delegated.With(),
			access.ConfirmCaveats{
				Cause: delegated.Nb().Cause,
				Iss:   delegated.Nb().Iss,
				Aud:   delegated.Nb().Aud,
				Att: []access.CapabilityRequest{
					{Can: "stuff/do"},
					{Can: "stuff/yell"},
				},
			},
		)

		fail := access.ConfirmDerive(claimed, delegated)
		require.Error(t, fail)
	})

	t.Run("accepts a subset of delegated capabilities", func(t *testing.T) {
		claimed := ucan.NewCapability(
			delegated.Can(),
			delegated.With(),
			access.ConfirmCaveats{
				Cause: delegated.Nb().Cause,
				Iss:   delegated.Nb().Iss,
				Aud:   delegated.Nb().Aud,
				Att: []access.CapabilityRequest{
					{Can: "stuff/do"},
				},
			},
		)

		fail := access.ConfirmDerive(claimed, delegated)
		require.NoError(t, fail)
	})

	t.Run("accepts any capabilities when wildcard is delegated", func(t *testing.T) {
		wildcardDelegated := ucan.NewCapability(
			delegated.Can(),
			delegated.With(),
			access.ConfirmCaveats{
				Cause: delegated.Nb().Cause,
				Iss:   delegated.Nb().Iss,
				Aud:   delegated.Nb().Aud,
				Att: []access.CapabilityRequest{
					{Can: "*"},
				},
			},
		)

		claimed := ucan.NewCapability(
			delegated.Can(),
			delegated.With(),
			access.ConfirmCaveats{
				Cause: delegated.Nb().Cause,
				Iss:   delegated.Nb().Iss,
				Aud:   delegated.Nb().Aud,
				Att: []access.CapabilityRequest{
					{Can: "stuff/do"},
					{Can: "stuff/yell"},
				},
			},
		)

		fail := access.ConfirmDerive(claimed, wildcardDelegated)
		require.NoError(t, fail)
	})

	t.Run("rejects a non-matching cause", func(t *testing.T) {
		claimed := ucan.NewCapability(
			delegated.Can(),
			delegated.With(),
			access.ConfirmCaveats{
				Cause: testutil.RandomCID(t),
				Iss:   delegated.Nb().Iss,
				Aud:   delegated.Nb().Aud,
				Att:   delegated.Nb().Att,
			},
		)

		fail := access.ConfirmDerive(claimed, delegated)
		require.Error(t, fail)
	})
}

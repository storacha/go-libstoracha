package access_test

import (
	"testing"

	"github.com/storacha/go-libstoracha/capabilities/access"
	"github.com/storacha/go-libstoracha/internal/testutil"
	"github.com/storacha/go-ucanto/ucan"
	"github.com/stretchr/testify/require"
)

func buildDelegationLinksModel(t *testing.T, links ...ucan.Link) access.DelegationLinksModel {
	delegations := access.DelegationLinksModel{
		Keys:   make([]string, 0, len(links)),
		Values: make(map[string]ucan.Link),
	}

	for _, link := range links {
		delegations.Keys = append(delegations.Keys, link.String())
		delegations.Values[link.String()] = link
	}

	return delegations
}

func TestRoundTripDelegateCaveats(t *testing.T) {
	delegations := buildDelegationLinksModel(t, testutil.RandomCID(t), testutil.RandomCID(t))
	nb := access.DelegateCaveats{
		Delegations: delegations,
	}

	node, err := nb.ToIPLD()
	require.NoError(t, err)

	rnb, err := access.DelegateCaveatsReader.Read(node)
	require.NoError(t, err)
	require.Equal(t, nb, rnb)
}

func TestDelegateDerive(t *testing.T) {
	// alice := testutil.Must(did.Parse("did:mailto:example.com:alice"))(t)
	// bob := testutil.Must(did.Parse("did:mailto:example.com:bob"))(t)

	del1Cid := testutil.RandomCID(t)
	del2Cid := testutil.RandomCID(t)

	delegated := ucan.NewCapability(
		access.DelegateAbility,
		"did:mailto:example.com:alice",
		access.DelegateCaveats{
			Delegations: buildDelegationLinksModel(t, del1Cid, del2Cid),
		},
	)

	t.Run("accepts an identical capability", func(t *testing.T) {
		claimed := ucan.NewCapability(
			delegated.Can(),
			delegated.With(),
			delegated.Nb(),
		)

		fail := access.DelegateDerive(claimed, delegated)
		require.NoError(t, fail)
	})

	t.Run("rejects the wrong resource", func(t *testing.T) {
		claimed := ucan.NewCapability(
			delegated.Can(),
			"did:mailto:example.com:bob",
			delegated.Nb(),
		)

		fail := access.DelegateDerive(claimed, delegated)
		require.Error(t, fail)
	})

	// 	t.Run("rejects the wrong issuer", func(t *testing.T) {
	// 		claimed := ucan.NewCapability(
	// 			delegated.Can(),
	// 			delegated.With(),
	// 			access.DelegateCaveats{
	// 				Cause: delegated.Nb().Cause,
	// 				Iss:   bob,
	// 				Aud:   delegated.Nb().Aud,
	// 				Att:   delegated.Nb().Att,
	// 			},
	// 		)

	// 		fail := access.DelegateDerive(claimed, delegated)
	// 		require.Error(t, fail)
	// 	})

	// 	t.Run("rejects the wrong audience", func(t *testing.T) {
	// 		claimed := ucan.NewCapability(
	// 			delegated.Can(),
	// 			delegated.With(),
	// 			access.DelegateCaveats{
	// 				Cause: delegated.Nb().Cause,
	// 				Iss:   delegated.Nb().Iss,
	// 				Aud:   alice,
	// 				Att:   delegated.Nb().Att,
	// 			},
	// 		)

	// 		fail := access.DelegateDerive(claimed, delegated)
	// 		require.Error(t, fail)
	// 	})

	t.Run("rejects non-delegated capabilities", func(t *testing.T) {
		claimed := ucan.NewCapability(
			delegated.Can(),
			delegated.With(),
			access.DelegateCaveats{
				Delegations: buildDelegationLinksModel(t, del1Cid, testutil.RandomCID(t)),
			},
		)

		fail := access.DelegateDerive(claimed, delegated)
		require.Error(t, fail)
	})

	t.Run("accepts a subset of delegated capabilities", func(t *testing.T) {
		claimed := ucan.NewCapability(
			delegated.Can(),
			delegated.With(),
			access.DelegateCaveats{
				Delegations: buildDelegationLinksModel(t, del1Cid),
			},
		)

		fail := access.DelegateDerive(claimed, delegated)
		require.NoError(t, fail)
	})
}

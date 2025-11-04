package ucan_test

import (
	"testing"

	ucancap "github.com/storacha/go-libstoracha/capabilities/ucan"
	"github.com/storacha/go-libstoracha/testutil"
	"github.com/storacha/go-ucanto/core/dag/blockstore"
	"github.com/storacha/go-ucanto/core/delegation"
	"github.com/storacha/go-ucanto/core/invocation"
	"github.com/storacha/go-ucanto/core/message"
	"github.com/storacha/go-ucanto/did"
	"github.com/storacha/go-ucanto/principal/absentee"
	"github.com/storacha/go-ucanto/transport/car"
	"github.com/storacha/go-ucanto/ucan"
	"github.com/storacha/go-ucanto/validator"
	"github.com/stretchr/testify/require"
)

func TestAttest(t *testing.T) {
	alice := testutil.Alice
	account := absentee.From(testutil.Must(did.Parse("did:mailto:web.mail:alice"))(t))
	space := testutil.Mallory
	service := testutil.WebService

	// delegation from space to mailto account
	adminDlg, err := delegation.Delegate(
		space,
		account,
		[]ucan.Capability[ucan.NoCaveats]{
			ucan.NewCapability("*", "ucan:*", ucan.NoCaveats{}),
		},
	)
	require.NoError(t, err)

	// delegation from mailto account to alice key
	accountDlg, err := delegation.Delegate(
		account,
		alice,
		[]ucan.Capability[ucan.NoCaveats]{
			ucan.NewCapability("blob/add", space.DID().String(), ucan.NoCaveats{}),
		},
	)
	require.NoError(t, err)

	// attestation for account delegation
	attestCaveats := ucancap.AttestCaveats{
		Proof: accountDlg.Link(),
	}
	attestDlg, err := ucancap.Attest.Delegate(
		service,
		alice,
		service.DID().String(),
		attestCaveats,
	)
	require.NoError(t, err)

	// invocation of ability, passing the required proofs
	inv, err := invocation.Invoke(
		alice,
		service,
		ucan.NewCapability("blob/add", space.DID().String(), ucan.NoCaveats{}),
		delegation.WithProof(
			delegation.FromDelegation(attestDlg),
			delegation.FromDelegation(accountDlg),
			delegation.FromDelegation(adminDlg),
		),
	)
	require.NoError(t, err)

	msg := roundTripAgentMessage(t, inv)

	br, err := blockstore.NewBlockReader(blockstore.WithBlocksIterator(msg.Blocks()))
	require.NoError(t, err)

	d, err := delegation.NewDelegationView(attestDlg.Link(), br)
	require.NoError(t, err)

	match, err := ucancap.Attest.Match(validator.NewSource(d.Capabilities()[0], d))
	require.NoError(t, err)
	require.Equal(t, attestCaveats.Proof, match.Value().Nb().Proof)
}

func roundTripAgentMessage(t *testing.T, inv invocation.Invocation) message.AgentMessage {
	t.Helper()
	inMsg, err := message.Build([]invocation.Invocation{inv}, nil)
	require.NoError(t, err)

	outCodec := car.NewOutboundCodec()
	req, err := outCodec.Encode(inMsg)
	require.NoError(t, err)

	inCodec, err := car.NewInboundCodec().Accept(req)
	require.NoError(t, err)

	outMsg, err := inCodec.Decoder().Decode(req)
	require.NoError(t, err)

	return outMsg
}

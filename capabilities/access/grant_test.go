package access_test

import (
	"io"
	"testing"

	"github.com/storacha/go-libstoracha/capabilities/access"
	"github.com/storacha/go-libstoracha/testutil"
	"github.com/storacha/go-ucanto/core/delegation"
	"github.com/storacha/go-ucanto/core/invocation"
	"github.com/storacha/go-ucanto/core/message"
	"github.com/storacha/go-ucanto/core/receipt"
	"github.com/storacha/go-ucanto/core/receipt/ran"
	"github.com/storacha/go-ucanto/core/result"
	"github.com/storacha/go-ucanto/transport/car"
	"github.com/storacha/go-ucanto/ucan"
	"github.com/stretchr/testify/require"
)

func TestGrant(t *testing.T) {
	grantCaveats := access.GrantCaveats{
		Att:   []access.CapabilityRequest{{Can: "admin/party"}},
		Cause: testutil.RandomCID(t),
	}

	inv, err := access.Grant.Invoke(
		testutil.Alice,
		testutil.Bob,
		testutil.Alice.DID().String(),
		grantCaveats,
	)
	require.NoError(t, err)

	t.Run("round trip", func(t *testing.T) {
		d0, err := delegation.Delegate(
			testutil.Bob,
			testutil.Alice,
			[]ucan.Capability[ucan.NoCaveats]{
				ucan.NewCapability(grantCaveats.Att[0].Can, testutil.Bob.DID().String(), ucan.NoCaveats{}),
			},
		)
		require.NoError(t, err)

		d0Bytes, err := io.ReadAll(d0.Archive())
		require.NoError(t, err)

		grantOk := access.GrantOk{
			Delegations: access.DelegationsModel{
				Keys:   []string{d0.Link().String()},
				Values: map[string][]byte{d0.Link().String(): d0Bytes},
			},
		}

		r0, err := receipt.Issue(
			testutil.Bob,
			result.Ok[access.GrantOk, access.GrantError](grantOk),
			ran.FromInvocation(inv),
		)
		require.NoError(t, err)

		// round trip the invocation and receipt in an agent message to ensure the
		// invocation can be encoded and the receipt decoded
		msg := roundTripAgentMessage(t, []invocation.Invocation{inv}, []receipt.AnyReceipt{r0})

		rcptLink, ok := msg.Get(inv.Link())
		require.True(t, ok)

		reader, err := access.NewGrantReceiptReader()
		require.NoError(t, err)

		r1, err := reader.Read(rcptLink, msg.Blocks())
		require.NoError(t, err)

		o, x := result.Unwrap(r1.Out())
		require.Empty(t, x)
		require.Len(t, o.Delegations.Keys, 1)
		require.Equal(t, d0.Link().String(), o.Delegations.Keys[0])

		_, err = delegation.Extract(o.Delegations.Values[d0.Link().String()])
		require.NoError(t, err)
	})

	t.Run("round trip with error", func(t *testing.T) {
		grantErr := access.GrantError{
			Name:    "Unauthorized",
			Message: "No, no you may not.",
		}

		r0, err := receipt.Issue(
			testutil.Bob,
			result.Error[access.GrantOk](grantErr),
			ran.FromInvocation(inv),
		)
		require.NoError(t, err)

		msg := roundTripAgentMessage(t, []invocation.Invocation{inv}, []receipt.AnyReceipt{r0})

		rcptLink, ok := msg.Get(inv.Link())
		require.True(t, ok)

		reader, err := access.NewGrantReceiptReader()
		require.NoError(t, err)

		r1, err := reader.Read(rcptLink, msg.Blocks())
		require.NoError(t, err)

		o, x := result.Unwrap(r1.Out())
		require.Empty(t, o)
		require.Equal(t, grantErr.Name, x.Name)
		require.Equal(t, grantErr.Message, x.Message)
	})
}

func roundTripAgentMessage(t *testing.T, invs []invocation.Invocation, rcpts []receipt.AnyReceipt) message.AgentMessage {
	t.Helper()
	inMsg, err := message.Build(invs, rcpts)
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

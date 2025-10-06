package blob_test

import (
	"testing"

	"github.com/storacha/go-libstoracha/capabilities/blob"
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

func TestRetrieve(t *testing.T) {
	retrieveCaveats := blob.RetrieveCaveats{
		Blob: blob.Blob{Digest: testutil.RandomMultihash(t)},
	}

	dlg, err := delegation.Delegate(
		testutil.Bob,
		testutil.Alice,
		[]ucan.Capability[ucan.NoCaveats]{
			ucan.NewCapability(blob.RetrieveAbility, testutil.Bob.DID().String(), ucan.NoCaveats{}),
		},
	)
	require.NoError(t, err)

	inv, err := blob.Retrieve.Invoke(
		testutil.Alice,
		testutil.Bob,
		testutil.Bob.DID().String(),
		retrieveCaveats,
		delegation.WithProof(delegation.FromDelegation(dlg)),
	)
	require.NoError(t, err)

	t.Run("round trip", func(t *testing.T) {
		retrieveOk := blob.RetrieveOk{}

		r0, err := receipt.Issue(
			testutil.Bob,
			result.Ok[blob.RetrieveOk, blob.RetrieveError](retrieveOk),
			ran.FromInvocation(inv),
		)
		require.NoError(t, err)

		// round trip the invocation and receipt in an agent message to ensure the
		// invocation can be encoded and the receipt decoded
		msg := roundTripAgentMessage(t, []invocation.Invocation{inv}, []receipt.AnyReceipt{r0})

		rcptLink, ok := msg.Get(inv.Link())
		require.True(t, ok)

		reader, err := blob.NewRetrieveReceiptReader()
		require.NoError(t, err)

		r1, err := reader.Read(rcptLink, msg.Blocks())
		require.NoError(t, err)

		o, x := result.Unwrap(r1.Out())
		require.Empty(t, x)
		require.Equal(t, retrieveOk, o)
	})

	t.Run("round trip with error", func(t *testing.T) {
		retrieveErr := blob.RetrieveError{
			Name:    "NotFound",
			Message: "I do not have it",
		}

		r0, err := receipt.Issue(
			testutil.Bob,
			result.Error[blob.RetrieveOk](retrieveErr),
			ran.FromInvocation(inv),
		)
		require.NoError(t, err)

		msg := roundTripAgentMessage(t, []invocation.Invocation{inv}, []receipt.AnyReceipt{r0})

		rcptLink, ok := msg.Get(inv.Link())
		require.True(t, ok)

		reader, err := blob.NewRetrieveReceiptReader()
		require.NoError(t, err)

		r1, err := reader.Read(rcptLink, msg.Blocks())
		require.NoError(t, err)

		o, x := result.Unwrap(r1.Out())
		require.Empty(t, o)
		require.Equal(t, retrieveErr.Name, x.Name)
		require.Equal(t, retrieveErr.Message, x.Message)
		require.Equal(t, retrieveErr.Error(), x.Error())
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

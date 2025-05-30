package index_test

import (
	"bytes"
	"testing"

	"github.com/ipfs/go-cid"
	"github.com/ipld/go-ipld-prime/codec/dagjson"
	cidlink "github.com/ipld/go-ipld-prime/linking/cid"
	basicnode "github.com/ipld/go-ipld-prime/node/basic"
	"github.com/storacha/go-libstoracha/capabilities/index"
	"github.com/stretchr/testify/require"
)

func TestRoundTripAddCaveats(t *testing.T) {
	testCid, err := cid.Parse("bafybeiduiecxoeiqs3gyc6r7v3lymmhserldnpw62qjnhmqsulqjxjmtzi")
	require.NoError(t, err)

	nb := index.AddCaveats{
		Index: cidlink.Link{Cid: testCid},
	}

	node, err := nb.ToIPLD()
	require.NoError(t, err)

	buf := bytes.NewBuffer([]byte{})
	err = dagjson.Encode(node, buf)
	require.NoError(t, err)

	builder := basicnode.Prototype.Any.NewBuilder()
	err = dagjson.Decode(builder, buf)
	require.NoError(t, err)

	rnb, err := index.AddCaveatsReader.Read(builder.Build())
	require.NoError(t, err)
	require.Equal(t, nb, rnb)
}

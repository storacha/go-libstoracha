package index_test

import (
	"bytes"
	"testing"

	"github.com/ipld/go-ipld-prime/codec/dagjson"
	basicnode "github.com/ipld/go-ipld-prime/node/basic"
	"github.com/storacha/go-libstoracha/capabilities/index"
	"github.com/storacha/go-libstoracha/capabilities/internal/testutil"
	"github.com/stretchr/testify/require"
)

func TestRoundTripAddCaveats(t *testing.T) {
	nb := index.AddCaveats{
		Index: testutil.RandomCID(t),
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

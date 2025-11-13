package types_test

import (
	"fmt"
	"math/big"
	"testing"

	"github.com/ipld/go-ipld-prime"
	"github.com/ipld/go-ipld-prime/codec/dagjson"
	"github.com/ipld/go-ipld-prime/schema"
	"github.com/storacha/go-libstoracha/capabilities/types"
	"github.com/stretchr/testify/require"
)

func testTypeSystem() *schema.TypeSystem {
	ts, err := types.LoadSchemaBytes([]byte(`
		type BigIntContainer struct {
			big BigInt
		}
		type BytesContainer struct {
			big Bytes
		}
	`))
	if err != nil {
		panic(fmt.Errorf("loading types schema: %w", err))
	}
	return ts
}

type bigIntContainer struct {
	Big *big.Int
}

type bytesContainer struct {
	Big []byte
}

func TestBigIntConverter(t *testing.T) {
	testCases := []struct {
		name string
		n    *big.Int
	}{
		{
			name: "positive",
			n:    big.NewInt(1),
		},
		{
			name: "negative",
			n:    big.NewInt(-1),
		},
		{
			name: "zero",
			n:    big.NewInt(0),
		},
	}

	for _, tc := range testCases {
		t.Run("roundtrip "+tc.name, func(t *testing.T) {
			in := bigIntContainer{Big: tc.n}
			buf, err := ipld.Marshal(dagjson.Encode, &in, testTypeSystem().TypeByName("BigIntContainer"), types.Converters...)
			require.NoError(t, err)
			t.Log(string(buf))

			var out bigIntContainer
			_, err = ipld.Unmarshal(buf, dagjson.Decode, &out, testTypeSystem().TypeByName("BigIntContainer"), types.Converters...)
			require.NoError(t, err)

			require.Equal(t, tc.n, out.Big)
		})
	}

	t.Run("decode empty bytes", func(t *testing.T) {
		in := bytesContainer{}
		buf, err := ipld.Marshal(dagjson.Encode, &in, testTypeSystem().TypeByName("BytesContainer"), types.Converters...)
		require.NoError(t, err)
		t.Log(string(buf))

		var out bigIntContainer
		_, err = ipld.Unmarshal(buf, dagjson.Decode, &out, testTypeSystem().TypeByName("BigIntContainer"), types.Converters...)
		require.NoError(t, err)

		require.Equal(t, big.NewInt(0), out.Big)
	})

	t.Run("decode invalid bytes", func(t *testing.T) {
		in := bytesContainer{Big: []byte{2}}
		buf, err := ipld.Marshal(dagjson.Encode, &in, testTypeSystem().TypeByName("BytesContainer"), types.Converters...)
		require.NoError(t, err)
		t.Log(string(buf))

		var out bigIntContainer
		_, err = ipld.Unmarshal(buf, dagjson.Decode, &out, testTypeSystem().TypeByName("BigIntContainer"), types.Converters...)
		require.ErrorIs(t, err, types.ErrInvalidSign)
	})
}

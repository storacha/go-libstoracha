package encryption_test

import (
	"testing"

	"github.com/storacha/go-libstoracha/capabilities/space/encryption"
	"github.com/storacha/go-libstoracha/testutil"
	"github.com/stretchr/testify/require"
)

func TestRoundTripKeyDecryptCaveats(t *testing.T) {
	nb := encryption.KeyDecryptCaveats{
		Key: testutil.RandomBytes(t, 8),
	}

	node, err := nb.ToIPLD()
	require.NoError(t, err)

	rnb, err := encryption.KeyDecryptCaveatsReader.Read(node)
	require.NoError(t, err)
	require.Equal(t, nb, rnb)
}

func TestRoundTripKeyDecryptCaveats_EmptyKey(t *testing.T) {
	nb := encryption.KeyDecryptCaveats{
		Key: []byte{},
	}

	node, err := nb.ToIPLD()
	require.NoError(t, err)

	rnb, err := encryption.KeyDecryptCaveatsReader.Read(node)
	require.NoError(t, err)
	require.Equal(t, nb, rnb)
}

func TestRoundTripKeyDecryptOk(t *testing.T) {
	ok := encryption.KeyDecryptOk{
		DecryptedSymmetricKey: "dGVzdC1iYXNlNjQtZW5jb2RlZC1wYXlsb2Fk",
	}

	node, err := ok.ToIPLD()
	require.NoError(t, err)

	rok, err := encryption.KeyDecryptOkReader.Read(node)
	require.NoError(t, err)
	require.Equal(t, ok, rok)
}

func TestRoundTripKeyDecryptCaveats_LargeKey(t *testing.T) {
	// Simulate a 256-byte RSA-OAEP ciphertext
	nb := encryption.KeyDecryptCaveats{Key: testutil.RandomBytes(t, 256)}

	node, err := nb.ToIPLD()
	require.NoError(t, err)

	rnb, err := encryption.KeyDecryptCaveatsReader.Read(node)
	require.NoError(t, err)
	require.Equal(t, nb, rnb)
}

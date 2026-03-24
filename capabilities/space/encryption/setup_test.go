package encryption_test

import (
	"testing"

	"github.com/storacha/go-libstoracha/capabilities/space/encryption"
	"github.com/stretchr/testify/require"
)

func TestRoundTripSetupCaveats_NoOptionalFields(t *testing.T) {
	nb := encryption.SetupCaveats{}

	node, err := nb.ToIPLD()
	require.NoError(t, err)

	rnb, err := encryption.SetupCaveatsReader.Read(node)
	require.NoError(t, err)
	require.Equal(t, nb, rnb)
}

func TestRoundTripSetupCaveats_WithOptionalFields(t *testing.T) {
	location := "us-central1"
	keyring := "my-keyring"
	nb := encryption.SetupCaveats{
		Location: &location,
		Keyring:  &keyring,
	}

	node, err := nb.ToIPLD()
	require.NoError(t, err)

	rnb, err := encryption.SetupCaveatsReader.Read(node)
	require.NoError(t, err)
	require.Equal(t, nb, rnb)
	require.Equal(t, location, *rnb.Location)
	require.Equal(t, keyring, *rnb.Keyring)
}

func TestRoundTripSetupCaveats_OnlyLocation(t *testing.T) {
	location := "europe-west1"
	nb := encryption.SetupCaveats{
		Location: &location,
	}

	node, err := nb.ToIPLD()
	require.NoError(t, err)

	rnb, err := encryption.SetupCaveatsReader.Read(node)
	require.NoError(t, err)
	require.Equal(t, location, *rnb.Location)
	require.Nil(t, rnb.Keyring)
}

func TestRoundTripSetupOk(t *testing.T) {
	ok := encryption.SetupOk{
		PublicKey: "MIIBIjANBgkqhkiG9w0BAQEFAAOCAQ8AMIIBCgKCAQEA",
		Provider:  "storacha",
		Algorithm: "RSA-OAEP-256",
	}

	node, err := ok.ToIPLD()
	require.NoError(t, err)

	rok, err := encryption.SetupOkReader.Read(node)
	require.NoError(t, err)
	require.Equal(t, ok, rok)
}

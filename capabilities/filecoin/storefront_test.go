package filecoin_test

import (
	"testing"

	"github.com/storacha/go-libstoracha/capabilities/filecoin"
	"github.com/stretchr/testify/require"
)

func TestFilecoinCapability(t *testing.T) {
	capability := filecoin.Filecoin

	t.Run("has correct ability", func(t *testing.T) {
		require.Equal(t, "filecoin/*", capability.Can())
	})
	
	t.Run("validates DID format", func(t *testing.T) {
		validDid := "did:key:z6MkrZ1r5XBFZjB9WFxjbpGZdBLTZ5MsSKz2Ur9aCiX4HHcF"
		invalidDid := "not-a-did"
		
		err := filecoin.ValidateSpaceDID(validDid)
		require.Nil(t, err)
		
		err = filecoin.ValidateSpaceDID(invalidDid)
		require.NotNil(t, err)
		require.Contains(t, err.Error(), "invalid DID format")
	})
}
package upload_test

import (
	"testing"

	"github.com/storacha/go-ucanto/core/ipld"
	"github.com/storacha/go-ucanto/ucan"
	"github.com/storacha/go-libstoracha/capabilities/upload"
	"github.com/storacha/go-libstoracha/internal/testutil"
	"github.com/stretchr/testify/require"
)

func TestUploadCapability(t *testing.T) {
	capability := upload.Upload

	t.Run("has correct ability", func(t *testing.T) {
		require.Equal(t, "upload/*", capability.Can())
	})
}

func TestUploadWildcardDerivation(t *testing.T) {
	spaceDID := "did:key:z6MkhaXgBZDvotDkL5257faiztiGiC2QtKLGpbnnEGta2doK"
	root := testutil.RandomCID(t)
	shards := []ipld.Link{testutil.RandomCID(t), testutil.RandomCID(t)}

	t.Run("upload/add can be derived from upload/*", func(t *testing.T) {
		// Create a wildcard capability delegation
		wildcardCap := ucan.NewCapability(upload.UploadAbility, spaceDID, struct{}{})
		
		// Create a specific upload/add capability
		addCap := ucan.NewCapability(upload.AddAbility, spaceDID, upload.AddCaveats{
			Root:   root,
			Shards: shards,
		})

		// Test derivation using the dedicated derivation function
		fail := upload.AddDerive(addCap, wildcardCap)
		require.NoError(t, fail, "upload/add should be derivable from upload/*")
	})

	t.Run("upload/get can be derived from upload/*", func(t *testing.T) {
		// Create a wildcard capability delegation
		wildcardCap := ucan.NewCapability(upload.UploadAbility, spaceDID, struct{}{})
		
		// Create a specific upload/get capability
		getCap := ucan.NewCapability(upload.GetAbility, spaceDID, upload.GetCaveats{
			Root: root,
		})

		// Test derivation using the dedicated derivation function
		fail := upload.GetDerive(getCap, wildcardCap)
		require.NoError(t, fail, "upload/get should be derivable from upload/*")
	})

	t.Run("upload/remove can be derived from upload/*", func(t *testing.T) {
		// Create a wildcard capability delegation
		wildcardCap := ucan.NewCapability(upload.UploadAbility, spaceDID, struct{}{})
		
		// Create a specific upload/remove capability
		removeCap := ucan.NewCapability(upload.RemoveAbility, spaceDID, upload.RemoveCaveats{
			Root: root,
		})

		// Test derivation using the dedicated derivation function
		fail := upload.RemoveDerive(removeCap, wildcardCap)
		require.NoError(t, fail, "upload/remove should be derivable from upload/*")
	})

	t.Run("upload/list can be derived from upload/*", func(t *testing.T) {
		// Create a wildcard capability delegation
		wildcardCap := ucan.NewCapability(upload.UploadAbility, spaceDID, struct{}{})
		
		// Create a specific upload/list capability
		listCap := ucan.NewCapability(upload.ListAbility, spaceDID, upload.ListCaveats{
			Cursor: nil,
			Size:   nil,
			Pre:    nil,
		})

		// Test derivation using the dedicated derivation function
		fail := upload.ListDerive(listCap, wildcardCap)
		require.NoError(t, fail, "upload/list should be derivable from upload/*")
	})

	t.Run("wildcard only allows same space DID", func(t *testing.T) {
		differentSpaceDID := "did:key:z6MkrZ1r5XBFZjBU34qyD8fueMbMRkKw17BZaq2ivKFjnz2z"
		
		// Create a wildcard capability delegation for one space
		wildcardCap := ucan.NewCapability(upload.UploadAbility, spaceDID, struct{}{})
		
		// Create a specific upload/add capability for a different space
		addCap := ucan.NewCapability(upload.AddAbility, differentSpaceDID, upload.AddCaveats{
			Root:   root,
			Shards: shards,
		})

		// Test that derivation fails when space DIDs don't match
		fail := upload.AddDerive(addCap, wildcardCap)
		require.Error(t, fail, "upload/add should not be derivable from upload/* with different space DID")
	})
}

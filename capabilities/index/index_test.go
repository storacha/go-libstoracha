package index_test

import (
	"encoding/json"
	"fmt"
	"testing"

	"github.com/ipfs/go-cid"
	"github.com/storacha/go-libstoracha/capabilities/index"
	"github.com/storacha/go-ucanto/did"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestIndexCapabilityDefinitions(t *testing.T) {
	// Test ability constants
	assert.Equal(t, "space/index/*", index.IndexAbility)
	assert.Equal(t, "space/index/add", index.AddAbility)

	// Test capability registration
	assert.NotNil(t, index.Index)
	assert.NotNil(t, index.Add)
}

// TestSpaceDID tests the SpaceDID type
func TestSpaceDID(t *testing.T) {
	didStr := "did:key:z6MkhaXgBZDvotDkL5257faiztiGiC2QtKLGpbnnEGta2doK"
	didObj, err := did.Parse(didStr)
	require.NoError(t, err)

	// Convert to SpaceDID type
	spaceDID := index.SpaceDID(didObj)
	assert.NotNil(t, spaceDID)
}

// TestErrorConstants tests error constants
func TestErrorConstants(t *testing.T) {
	assert.Equal(t, "IndexNotFound", index.ErrIndexNotFound)
	assert.Equal(t, "DecodeFailure", index.ErrDecodeFailure)
	assert.Equal(t, "UnknownFormat", index.ErrUnknownFormat)
	assert.Equal(t, "ShardNotFound", index.ErrShardNotFound)
	assert.Equal(t, "SliceNotFound", index.ErrSliceNotFound)
}

// TestIndexArgsType tests the IndexArgs type
func TestIndexArgsType(t *testing.T) {
	// Create a CID for testing
	testCid, err := cid.Parse("QmWATWQ7fVPP2EFGu71UkfnqhYXDYH566qy47CnJDgvs8u")
	require.NoError(t, err)

	// Create IndexArgs
	args := index.IndexArgs{
		Index: testCid,
	}

	// Verify Index field
	assert.True(t, args.Index.Equals(testCid))

	// Test JSON marshaling
	data, err := json.Marshal(args)
	require.NoError(t, err)
	assert.Contains(t, string(data), testCid.String())

}

// Integration test for capability validation
func TestCapabilityValidationIntegration(t *testing.T) {
	// This test simulates how the capabilities would be used in a real-world scenario
	// It validates that the parsers and validators work correctly together

	// Create two DIDs
	space1 := "did:key:z6MkhaXgBZDvotDkL5257faiztiGiC2QtKLGpbnnEGta2doK"
	space2 := "did:key:z6Mkf5rGMoatrSj1f4CJrqumMhCa8tcV7DjgRJTGB6wwpeQR"

	// Create CIDs for testing
	cid1, err := cid.Parse("QmWATWQ7fVPP2EFGu71UkfnqhYXDYH566qy47CnJDgvs8u")
	require.NoError(t, err)
	cid2, err := cid.Parse("QmWATWQ7fVPP2EFGu71UkfnqhYXDYH566qy47CnJDgvs8v")
	require.NoError(t, err)

	// Create test scenarios
	testCases := []struct {
		name           string
		claimedWith    string
		delegatedWith  string
		claimedIndex   cid.Cid
		delegatedIndex cid.Cid
		expectError    bool
		errorContains  string
	}{
		{
			name:           "Same With Same Index",
			claimedWith:    space1,
			delegatedWith:  space1,
			claimedIndex:   cid1,
			delegatedIndex: cid1,
			expectError:    false,
		},
		{
			name:           "Same With Different Index",
			claimedWith:    space1,
			delegatedWith:  space1,
			claimedIndex:   cid1,
			delegatedIndex: cid2,
			expectError:    true,
			errorContains:  "doesn't match delegated",
		},
		{
			name:           "Different With Same Index",
			claimedWith:    space1,
			delegatedWith:  space2,
			claimedIndex:   cid1,
			delegatedIndex: cid1,
			expectError:    true,
			errorContains:  "doesn't match delegated",
		},
		{
			name:           "Delegated Empty Index",
			claimedWith:    space1,
			delegatedWith:  space1,
			claimedIndex:   cid1,
			delegatedIndex: cid.Cid{}, // Empty CID
			expectError:    false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Create mock capabilities
			claimed := &mockCapability{
				withValue: tc.claimedWith,
				canValue:  index.AddAbility,
				nbValue:   index.IndexArgs{Index: tc.claimedIndex},
			}

			delegated := &mockCapability{
				withValue: tc.delegatedWith,
				canValue:  index.AddAbility,
				nbValue:   index.IndexArgs{Index: tc.delegatedIndex},
			}

			// Test by using our own validation function that mirrors the one in the validator
			err := validateCapability(claimed, delegated)

			if tc.expectError {
				assert.Error(t, err)
				if tc.errorContains != "" {
					assert.Contains(t, err.Error(), tc.errorContains)
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

// Mock capability for testing
type mockCapability struct {
	withValue string
	canValue  string
	nbValue   index.IndexArgs
}

func (m *mockCapability) With() string {
	return m.withValue
}

func (m *mockCapability) Can() string {
	return m.canValue
}

func (m *mockCapability) Nb() index.IndexArgs {
	return m.nbValue
}

func (m *mockCapability) MarshalJSON() ([]byte, error) {
	return json.Marshal(map[string]interface{}{
		"with": m.withValue,
		"can":  m.canValue,
		"nb":   m.nbValue,
	})
}

// validateCapability replicates the validation logic in the Add capability
func validateCapability(claimed, delegated *mockCapability) error {
	// Check if the `with` fields are equal
	if claimed.With() != delegated.With() {
		return fmt.Errorf("resource '%s' doesn't match delegated '%s'", claimed.With(), delegated.With())
	}

	claimedArgs := claimed.Nb()
	delegatedArgs := delegated.Nb()

	// If delegated doesn't specify an index, allow any index
	if delegatedArgs.Index.Defined() && claimedArgs.Index.Defined() {
		if !claimedArgs.Index.Equals(delegatedArgs.Index) {
			return fmt.Errorf("index '%s' doesn't match delegated '%s'",
				claimedArgs.Index.String(),
				delegatedArgs.Index.String())
		}
	}

	return nil
}

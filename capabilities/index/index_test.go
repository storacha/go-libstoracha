package index_test

import (
	"encoding/json"
	"fmt"
	"testing"

	"github.com/ipfs/go-cid"
	cidlink "github.com/ipld/go-ipld-prime/linking/cid"
	"github.com/storacha/go-libstoracha/capabilities/index"
	"github.com/storacha/go-ucanto/core/schema"
	"github.com/storacha/go-ucanto/ucan"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestIndexCaveatsType tests the IndexCaveats type
func TestIndexCaveatsType(t *testing.T) {
	// Create a CID for testing
	testCid, err := cid.Parse("QmWATWQ7fVPP2EFGu71UkfnqhYXDYH566qy47CnJDgvs8u")
	require.NoError(t, err)

	args := index.IndexCaveats{
		Index: cidlink.Link{Cid: testCid},
	}

	assert.Equal(t, testCid.String(), args.Index.String())

	data, err := json.Marshal(args)
	require.NoError(t, err)
	assert.Contains(t, string(data), testCid.String())
}

type testCapability struct {
	withValue string
	canValue  string
	nbValue   index.IndexCaveats
}

func (tc *testCapability) With() string {
	return tc.withValue
}

func (tc *testCapability) Can() string {
	return tc.canValue
}

func (tc *testCapability) Nb() index.IndexCaveats {
	return tc.nbValue
}

func (tc *testCapability) MarshalJSON() ([]byte, error) {
	return json.Marshal(map[string]interface{}{
		"with": tc.withValue,
		"can":  tc.canValue,
		"nb":   tc.nbValue,
	})
}

// Integration test for capability validation
func TestCapabilityValidationIntegration(t *testing.T) {
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
		claimedIndex   ucan.Link
		delegatedIndex ucan.Link
		expectError    bool
		errorContains  string
	}{
		{
			name:           "Same With Same Index",
			claimedWith:    space1,
			delegatedWith:  space1,
			claimedIndex:   cidlink.Link{Cid: cid1},
			delegatedIndex: cidlink.Link{Cid: cid1},
			expectError:    false,
		},
		{
			name:           "Same With Different Index",
			claimedWith:    space1,
			delegatedWith:  space1,
			claimedIndex:   cidlink.Link{Cid: cid1},
			delegatedIndex: cidlink.Link{Cid: cid2},
			expectError:    true,
			errorContains:  "doesn't match delegated",
		},
		{
			name:           "Different With Same Index",
			claimedWith:    space1,
			delegatedWith:  space2,
			claimedIndex:   cidlink.Link{Cid: cid1},
			delegatedIndex: cidlink.Link{Cid: cid1},
			expectError:    true,
			errorContains:  "doesn't match delegated",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Create capabilities that satisfy the ucan.Capability interface
			claimed := &testCapability{
				withValue: tc.claimedWith,
				canValue:  index.AddAbility,
				nbValue:   index.IndexCaveats{Index: tc.claimedIndex},
			}

			delegated := &testCapability{
				withValue: tc.delegatedWith,
				canValue:  index.AddAbility,
				nbValue:   index.IndexCaveats{Index: tc.delegatedIndex},
			}

			err := validateCapabilityWithExportedFunction(claimed, delegated)

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

func validateCapabilityWithExportedFunction(claimed, delegated *testCapability) error {
	// Check if the `with` fields are equal using the index.EqualWith function
	if err := index.EqualWith(claimed.With(), delegated.With()); err != nil {
		return err
	}

	claimedArgs := claimed.Nb()
	delegatedArgs := delegated.Nb()

	// If delegated doesn't specify an index, allow any index
	if delegatedArgs.Index != nil && claimedArgs.Index != nil {
		if claimedArgs.Index.String() != delegatedArgs.Index.String() {
			return schema.NewSchemaError(fmt.Sprintf(
				"index '%s' doesn't match delegated '%s'",
				claimedArgs.Index.String(),
				delegatedArgs.Index.String(),
			))
		}
	}

	return nil
}

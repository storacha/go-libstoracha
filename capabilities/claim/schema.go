package claim

import (
	// for go:embed
	_ "embed"
	"fmt"

	ipldschema "github.com/ipld/go-ipld-prime/schema"

	"github.com/storacha/go-libstoracha/capabilities/types"
)

//go:embed claim.ipldsch
var claimSchema []byte

var claimTypeSystem = mustLoadTS()

func mustLoadTS() *ipldschema.TypeSystem {
	ts, err := types.LoadSchemaBytes(claimSchema)
	if err != nil {
		panic(fmt.Errorf("loading claim schema: %w", err))
	}
	return ts
}

func CacheCaveatsType() ipldschema.Type {
	return claimTypeSystem.TypeByName("CacheCaveats")
}

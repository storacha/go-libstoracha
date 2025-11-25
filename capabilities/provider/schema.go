package provider

import (
	_ "embed"
	"fmt"

	"github.com/ipld/go-ipld-prime/schema"
	"github.com/storacha/go-libstoracha/capabilities/types"
)

//go:embed provider.ipldsch
var providerSchema []byte

var providerTS = mustLoadTS()

func mustLoadTS() *schema.TypeSystem {
	ts, err := types.LoadSchemaBytes(providerSchema)
	if err != nil {
		panic(fmt.Errorf("loading provider schema: %w", err))
	}
	return ts
}

func AddCaveatsType() schema.Type {
	return providerTS.TypeByName("AddCaveats")
}

func AddOkType() schema.Type {
	return providerTS.TypeByName("AddOk")
}

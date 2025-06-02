package pdp

import (
	// for go:embed
	_ "embed"
	"fmt"

	ipldschema "github.com/ipld/go-ipld-prime/schema"
	"github.com/storacha/go-libstoracha/capabilities/types"
)

//go:embed pdp.ipldsch
var pdpSchema []byte

var pdpTS = mustLoadTS()

func mustLoadTS() *ipldschema.TypeSystem {
	ts, err := types.LoadSchemaBytes(pdpSchema)
	if err != nil {
		panic(fmt.Errorf("loading blob schema: %w", err))
	}
	return ts
}

func AcceptCaveatsType() ipldschema.Type {
	return pdpTS.TypeByName("AcceptCaveats")
}

func AcceptOkType() ipldschema.Type {
	return pdpTS.TypeByName("AcceptOk")
}

func InfoCaveatsType() ipldschema.Type {
	return pdpTS.TypeByName("InfoCaveats")
}

func InfoOkType() ipldschema.Type {
	return pdpTS.TypeByName("InfoOk")
}

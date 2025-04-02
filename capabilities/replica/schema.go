package replica

import (
	_ "embed"
	"fmt"

	"github.com/ipld/go-ipld-prime/schema"

	"github.com/storacha/go-libstoracha/capabilities/types"
)

//go:embed replica.ipldsch
var replicaSchema []byte

var replicaTS = mustLoadTS()

func mustLoadTS() *schema.TypeSystem {
	ts, err := types.LoadSchemaBytes(replicaSchema)
	if err != nil {
		panic(fmt.Errorf("loading replica schema: %w", err))
	}
	return ts
}

func AllocateCaveatsType() schema.Type {
	return replicaTS.TypeByName("AllocateCaveats")
}

func TransferCaveatsType() schema.Type {
	return replicaTS.TypeByName("TransferCaveats")
}

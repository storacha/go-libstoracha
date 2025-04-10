package blob

import (
	// for schema embed
	_ "embed"
	"fmt"

	"github.com/ipld/go-ipld-prime/schema"

	"github.com/storacha/go-libstoracha/capabilities/types"
)

//go:embed blob.ipldsch
var blobSchema []byte

var blobTS = mustLoadTS()

func mustLoadTS() *schema.TypeSystem {
	ts, err := types.LoadSchemaBytes(blobSchema)
	if err != nil {
		panic(fmt.Errorf("loading blob schema: %w", err))
	}
	return ts
}

func AllocateCaveatsType() schema.Type {
	return blobTS.TypeByName("AllocateCaveats")
}

func AllocateOkType() schema.Type {
	return blobTS.TypeByName("AllocateOk")
}

func AcceptCaveatsType() schema.Type {
	return blobTS.TypeByName("AcceptCaveats")
}

func AcceptOkType() schema.Type {
	return blobTS.TypeByName("AcceptOk")
}
